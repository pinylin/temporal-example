package workflow

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"temporal-exmaple/config"
	lkf "temporal-exmaple/kafka"
)

var (
	CtxKeyWorkflowID = "workflowID"
)

func InitConsumerCron(tc *client.Client) *client.WorkflowRun {
	workflowID := "cron_" + config.Config.GetString("workflow.cron-consumer")
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "cron",
		CronSchedule: "*/2 * * * *",
	}

	_ = (*tc).CancelWorkflow(context.Background(), workflowID, "")
	tw, err := (*tc).ExecuteWorkflow(context.Background(), workflowOptions, CronConsumerWorkflow)
	if err != nil {
		slog.Error("Unable to start worker", err)
	}

	return &tw
}

func CronConsumerWorkflow(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Cron workflow started.", "StartTime", workflow.Now(ctx))
	count := 0
	var childWorkflows []workflow.Future
	for {
		count += 1
		if count >= 200 {
			break
		}
		//msg, err := kr.ReadMessage(c)
		//if err != nil {
		//	if !strings.HasSuffix(err.Error(),"EOF") {
		//		log.Printf("Error while reading message from kafka: %s\n", err)
		//		panic(err)
		//	} else {
		//		break
		//	}
		//}

		execution := workflow.GetInfo(ctx).WorkflowExecution
		childID := fmt.Sprintf("child_workflow:%v", execution.RunID) //, string(msg.Value))
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID: childID,
			//WorkflowID: string(msg.Value),
		}

		ctx = workflow.WithChildOptions(ctx, cwo)
		//ctx = context.WithValue(ctx, CtxKeyWorkflowID, string(msg.Value))
		childWorkflows = append(childWorkflows, workflow.ExecuteChildWorkflow(ctx, ChildCronConsumerWorkflow))
	}

	// 等待所有 ChildWorkflow 完成
	for _, childWorkflow := range childWorkflows {
		if err := childWorkflow.Get(ctx, nil); err != nil {
			return err
		}
	}

	return nil
}

func ChildCronConsumerWorkflow(ctx workflow.Context) (*CronResult, error) {
	workflow.GetLogger(ctx).Info("Cron workflow started.", "StartTime", workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	// Start from 0 for first cron job
	lastRunTime := time.Time{}
	// Check to see if there was a previous cron job
	if workflow.HasLastCompletionResult(ctx) {
		var lastResult CronResult
		if err := workflow.GetLastCompletionResult(ctx, &lastResult); err == nil {
			lastRunTime = lastResult.RunTime
		}
	}
	thisRunTime := workflow.Now(ctx)
	//workflowID := workflow.GetChildWorkflowOptions(ctx).WorkflowID

	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	topic := viper.GetString("kafka.topic")
	kr := lkf.GetKafkaConsumer(&topic)
	defer kr.Close()
	for {
		msg, err := kr.ReadMessage(c)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "EOF") {
				log.Printf("Error while reading message from kafka: %s\n", err)
				panic(err)
			} else {
				break
			}
		}
		ctx1 = workflow.WithValue(ctx1, CtxKeyWorkflowID, string(msg.Value))
		err = workflow.ExecuteActivity(ctx1, ConsumeActivityMsg, lastRunTime, thisRunTime).Get(ctx, nil)
		if err != nil {
			// Cron job failed
			// Next cron will still be scheduled by the Server
			workflow.GetLogger(ctx).Error("Cron job failed.", "Error", err)
			return nil, err
		}
	}

	return &CronResult{RunTime: thisRunTime}, nil
}

func ConsumeActivityMsg(ctx context.Context, lastRunTime, thisRunTime time.Time) error {
	activity.GetLogger(ctx).Info("Cron job running.", "lastRunTime_exclude", lastRunTime, "thisRunTime_include", thisRunTime)
	// Query database, call external API, or do any other non-deterministic action.
	workflowID := ctx.Value(CtxKeyWorkflowID).(string)
	slog.Info("Consume Message: %s\n", workflowID)

	return nil
}

func InitConsumerWorker(temporalClient *client.Client) *worker.Worker {
	queue := "cron"

	// use options to control max concurrent activity executions, retry policy and timeouts
	opts := worker.Options{
		MaxConcurrentActivityExecutionSize: 100,
	}
	tw := worker.New(*temporalClient, queue, opts)
	defer tw.Stop()

	tw.RegisterWorkflow(CronConsumerWorkflow)
	tw.RegisterWorkflow(ChildCronConsumerWorkflow)
	tw.RegisterWorkflow(ConsumerWorkflow)
	tw.RegisterActivity(ConsumeActivityMsg)

	err := tw.Run(worker.InterruptCh())
	if err != nil {
		slog.Error("Unable to start temporal worker", err)
	}

	return &tw
}

func ConsumerWorkflow(ctx workflow.Context) (*CronResult, error) {

	workflow.GetLogger(ctx).Info("Cron workflow started.", "StartTime", workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	// Start from 0 for first cron job
	lastRunTime := time.Time{}
	// Check to see if there was a previous cron job
	if workflow.HasLastCompletionResult(ctx) {
		var lastResult CronResult
		if err := workflow.GetLastCompletionResult(ctx, &lastResult); err == nil {
			lastRunTime = lastResult.RunTime
		}
	}
	thisRunTime := workflow.Now(ctx)

	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	topic := viper.GetString("kafka.topic")
	kr := lkf.GetKafkaConsumer(&topic)
	defer kr.Close()
	for {
		msg, err := kr.ReadMessage(c)
		if err != nil {
			if !strings.HasSuffix(err.Error(),"EOF") {
				log.Printf("Error while reading message from kafka: %s\n", err)
				panic(err)
			} else {
				break
			}
		}
		ctx1 = workflow.WithValue(ctx1, CtxKeyWorkflowID, string(msg.Value))
		err = workflow.ExecuteActivity(ctx1, ConsumeActivityMsg, lastRunTime, thisRunTime).Get(ctx, nil)
		if err != nil {
			// Cron job failed
			// Next cron will still be scheduled by the Server
			workflow.GetLogger(ctx).Error("Cron job failed.", "Error", err)
			return nil, err
		}
	}

	return &CronResult{RunTime: thisRunTime}, nil
}
