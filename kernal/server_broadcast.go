package kernal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func (s *Server) ListenBroadcastMessage() {
	// 为当前服务器生成唯一的消费者组ID，确保所有节点都能收到全量的广播消息
	groupID := "server_node_" + uuid.New().String()
	// 订阅 Kafka 的全局广播频道
	reader := NewKafkaReader(BroadcastTopic, groupID)

	// 启动一个协程去持续接收 Kafka 广播频道里的消息
	go func() {
		defer reader.Close()
		for {
			// 读取消息
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				fmt.Printf("Kafka BroadcastMsg Consumer 出错: %v\n", err)
				continue
			}

			// 将接收到的广播消息发送给所有的本地在线用户
			data, _ := json.Marshal(map[string]interface{}{
				"type": "Broadcast",
				"msg":  string(msg.Value),
			})

			s.mapLock.RLock()
			for _, client := range s.OnlineMap {
				client.C <- string(data)
			}
			s.mapLock.RUnlock()
		}
	}()
}

func BroadCast(user *User, msg string) {
	sendMsg := user.Name + ":" + msg
	// 将广播消息发布到全局广播频道 Kafka Topic
	writer := NewKafkaWriter(BroadcastTopic)
	defer writer.Close()

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(user.Name), // 可选 -- 按用户名做分区
			Value: []byte(sendMsg),
		},
	)
	if err != nil {
		fmt.Printf("Kafka BroadcastMsg Producer 失败: %v\n", err)
	}
}
