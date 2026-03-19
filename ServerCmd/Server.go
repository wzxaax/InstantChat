package main

import "InstantChat/kernal"

func main() {
	kernal.InitKafka()                            // 初始化 Kafka 连接
	kernal.InitRedis()                            // 初始化 Redis 连接
	Server := kernal.NewServer("127.0.0.1", 8888) // 创建聊天服务器
	Server.Start()                                // 启动聊天服务器
}
