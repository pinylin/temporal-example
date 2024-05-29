package kafka

import (
	"fmt"
	"log/slog"
	"sync"
	"temporal-exmaple/config"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/IBM/sarama"
)

//type Message kafka.Message

const (
	LastOffset  int64 = -1 // The most recent offset available for a partition.
	FirstOffset int64 = -2 // The least recent offset available for a partition.
)

var kwOnce sync.Once
var kw *kafka.Writer

//var krOnce sync.Once
//var kr *kafka.Reader

func InitKafkaProducer() {
	kwOnce.Do(func() {
		kafkaURL := config.Config.GetString("kafka.host")
		topic := config.Config.GetString("kafka.topic")
		//slog.Info("%s", kafkaURL)
		//slog.Info("%s", topic)

		kw = &kafka.Writer{
			Addr:     kafka.TCP(kafkaURL),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	})
}

func GetKafkaWriter() *kafka.Writer {
	kwOnce.Do(func() {
		kafkaURL := config.Config.GetString("kafka.host")
		topic := config.Config.GetString("kafka.topic")
		//slog.Info("%s", kafkaURL)
		//slog.Info("%s", topic)

		kw = &kafka.Writer{
			Addr:     kafka.TCP(kafkaURL),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	})
	return kw
}

func GetKafkaConsumer(topic *string) *kafka.Reader {
	kafkaURL := config.Config.GetString("kafka.host")

	//topic := viper.GetString("kafka.topic")
	brokers := []string{kafkaURL}
	//groupID := "consumer-group-1"
	groupID := config.Config.GetString("kafka.consumer-group")

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	consumerConf := kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       *topic,
		StartOffset: LastOffset,
		Dialer:      dialer,
		MinBytes:    1, // same value of Shopify/sarama
		MaxBytes:    57671680,
	}

	kr := kafka.NewReader(consumerConf)

	slog.Info("Kafka consumer initialised for topic: %s\n", *topic)

	return kr
}

func CreateTopic()  {
	kafkaURL := config.Config.GetString("kafka.host")
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin([]string{kafkaURL}, config)
	if err != nil {
		panic(err)
	}
	defer admin.Close()

	// 创建 Topic
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	err = admin.CreateTopic("my-topic", topicDetail, false)
	if err != nil {
		panic(err)
	}

	fmt.Println("Topic 'my-topic' created successfully")
}

func NewKafkaWriter() *kafka.Writer {
	kafkaURL := config.Config.GetString("kafka.host")
	topic := config.Config.GetString("kafka.topic")
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}