package kernal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func (s *Server) ListenPrivateMessage() {
	// 为当前服务器生成唯一的消费者组ID，确保所有节点都能收到全量的私聊消息并进行本地分发
	groupID := "server_node_" + uuid.New().String()
	// 订阅 Kafka 的全局私聊频道
	reader := NewKafkaReader(PrivateChatTopic, groupID)

	// 启动一个协程去持续接收 Kafka 私聊频道里的消息
	go func() {
		defer reader.Close()
		for {
			// 读取消息
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				fmt.Printf("Kafka PrivateMsg Consumer 出错: %v\n", err)
				continue
			}

			// 解析消息内容，获取目标用户名
			var data map[string]interface{}
			err = json.Unmarshal(msg.Value, &data)
			if err != nil {
				fmt.Printf("解析私聊消息失败: %v\n", err)
				continue
			}
			targetName, ok := data["target"].(string)
			if !ok {
				continue
			}

			// 检查目标用户是否在当前节点的在线列表中
			s.mapLock.RLock()
			targetUser, online := s.OnlineMap[targetName]
			s.mapLock.RUnlock()

			if online {
				// 如果目标用户连接到了本地服务器，则投递消息
				targetUser.C <- string(msg.Value)
			}
		}
	}()
}

func Private(remoteName string, data []byte) {
	// 将私聊消息发布到全局私聊频道 Kafka Topic
	writer := NewKafkaWriter(PrivateChatTopic)
	defer writer.Close()

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(remoteName), // 以目标用户名为Key，保证同一目标用户的消息被投递到同一分区，从而保证顺序性
			Value: data,
		},
	)
	if err != nil {
		fmt.Printf("Kafka PrivateMsg Producer 失败: %v\n", err)
	}
}
