package kernal

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// 全局 Kafka Brokers 地址
var KafkaBrokers = []string{"localhost:9092"}

const (
	// 私聊频道
	PrivateChatTopic = "channel_private_chat"
	// 广播频道
	BroadcastTopic = "channel_broadcast"
)

// 测试连接并尝试获取 metadata 确保 Kafka 存活
func InitKafka() {
	conn, err := kafka.DialContext(context.Background(), "tcp", KafkaBrokers[0])
	if err != nil {
		fmt.Printf("Kafka 连接失败，请确保您已经启动了 Kafka 服务器: %v\n", err)
		return
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		fmt.Printf("Kafka 获取 Brokers 失败，请检查配置: %v\n", err)
		return
	}

	fmt.Println("======== Kafka 连接成功，准备就绪！========")
}

// NewKafkaWriter 返回一个向指定 Topic 发送消息的 Writer
func NewKafkaWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(KafkaBrokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true, // 允许自动创建 Topic
		WriteTimeout:           10 * time.Second,
	}
}

// NewKafkaReader 返回一个从指定 Topic 消费消息的 Reader
func NewKafkaReader(topic string, groupID string) *kafka.Reader {
	config := kafka.ReaderConfig{
		Brokers:   KafkaBrokers,
		Topic:     topic,
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	}

	// 如果提供了 groupID，则使用消费者组模式（每个服务端实例拥有不同的 groupID，从而都能收到全量消息）
	if groupID != "" {
		config.GroupID = groupID
		// config.Partition = 0   # 使用 GroupID 时不能指定 Partition
		// ### 这个很重要，确保新启动的消费者从最新的消息开始消费，而不是从头消费引发消息风暴
		config.StartOffset = kafka.LastOffset
	} else {
		// 单点接收，不使用群组，直接读最新（目前未使用）
		// ### 设置 StartOffset 为 LastOffset 可以避免读取历史遗留消息 （kafka-go 没有 GroupID 时，默认从 FirstOffset 读取）
		config.StartOffset = kafka.LastOffset
		config.Partition = 0 // 单点模式下需要手动指定读取的分区
	}

	return kafka.NewReader(config)
}
