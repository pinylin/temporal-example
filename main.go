package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"temporal-exmaple/kafka"

	//"github.com/spf13/viper"
	//"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	//"go.temporal.io/sdk/worker"

	"temporal-exmaple/config"
	"temporal-exmaple/workflow"
)

func initTemporalClient() *client.Client {
	clientOptions := client.Options{
		Logger: log.NewStructuredLogger(
			slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}))),
		//HostPort: client.DefaultHostPort,
		HostPort: config.Config.GetString("temporal.host"),
	}
	slog.Info(config.Config.GetString("temporal.host"))
	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		panic(err)
	}
	return &temporalClient
}

func main() {

	config.InitConfig()
	kafka.InitKafkaProducer()
	//kafka.CreateTopic()
	//isMaster := os.Getenv("IS_MASTER")
	podName := os.Getenv("POD_NAME")
	isMaster := strings.HasSuffix(podName, "-0")
	slog.Warn(fmt.Sprintf("podName: %v,    IS_MASTER: %v", podName, isMaster))
	tc := initTemporalClient()
	if isMaster {
		slog.Warn("master init...")
		
		workflow.InitProducerCron(tc)
		workflow.InitConsumerCron(tc)
	}

	go workflow.InitWorker(tc)
	//go workflow.InitProducerWorker(tc)
	//go workflow.InitConsumerWorker(tc)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down after cleanup")

	// Close temporal worker and client
	(*tc).Close()

}
