package workflow

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/pborman/uuid"
	"github.com/segmentio/kafka-go"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"temporal-exmaple/config"
	lkf "temporal-exmaple/kafka"
)

func InitProducerCron(tc *client.Client) *client.WorkflowRun {
	workflowID := "cron_" + config.Config.GetString("workflow.cron-producer")
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "cron",
		CronSchedule: "* * * * *",
	}

	_ = (*tc).CancelWorkflow(context.Background(), workflowID, "")
	we, err := (*tc).ExecuteWorkflow(context.Background(), workflowOptions, CronProducerWorkflow)
	if err != nil {
		slog.Error("Unable to start worker", err)
	} else {
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}

	return &we
}

func CronProducerWorkflow(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Cron workflow started.", "StartTime", workflow.Now(ctx))

	// var childWorkflows []workflow.Future
	for i := 0; i < 200; i++ {
		execution := workflow.GetInfo(ctx).WorkflowExecution
		childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID: childID,
		}
		ctx = workflow.WithChildOptions(ctx, cwo)
		// childWorkflows = append(childWorkflows, workflow.ExecuteChildWorkflow(ctx, ChildCronProducerWorkflow))
		workflow.ExecuteChildWorkflow(ctx, ChildCronProducerWorkflow)
	}

	// 等待所有 ChildWorkflow 完成
	// for _, childWorkflow := range childWorkflows {
	// 	if err := childWorkflow.Get(ctx, nil); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func ChildCronProducerWorkflow(ctx workflow.Context) error {
	//workflowID := ctx.Value(CtxKeyWorkflowID).(string)
	workflowID := "cron_" + uuid.New()
	kw := lkf.NewKafkaWriter()
	defer kw.Close()
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("workflowID-%s", workflowID)),
		Value: []byte(workflowID),
	}
	err := kw.WriteMessages(context.Background(), msg)
	if err != nil {
		slog.Error(err.Error())
	}
	return err
}

func PushActivityMsg(ctx context.Context, lastRunTime, thisRunTime time.Time) error {
	activity.GetLogger(ctx).Info("Cron job running.", "lastRunTime_exclude", lastRunTime, "thisRunTime_include", thisRunTime)
	// Query database, call external API, or do any other non-deterministic action.

	//workflowID := ctx.Value(CtxKeyWorkflowID).(string)
	workflowID := "cron_" + uuid.New()
	kw := lkf.NewKafkaWriter()
	defer kw.Close()
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("workflowID-%s", workflowID)),
		Value: []byte(workflowID),
	}
	err := kw.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error(err.Error())
	}
	return err
}

func InitProducerWorker(temporalClient *client.Client) *worker.Worker {
	queue := "cron"

	// use options to control max concurrent activity executions, retry policy and timeouts
	opts := worker.Options{
		MaxConcurrentActivityExecutionSize: 100,
	}
	tw := worker.New(*temporalClient, queue, opts)
	defer tw.Stop()

	tw.RegisterWorkflow(CronProducerWorkflow)
	tw.RegisterWorkflow(ChildCronProducerWorkflow)
	tw.RegisterActivity(PushActivityMsg)

	err := tw.Run(worker.InterruptCh())
	if err != nil {
		slog.Error("Unable to start temporal worker", err)
	}

	return &tw
}
