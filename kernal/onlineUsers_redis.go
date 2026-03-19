package kernal

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

const (
	/* 全局的在线用户数据 Key */
	OnlineUsersKey = "instant_chat:online_users"
)

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 没有密码则留空
		DB:       0,  // 默认数据库
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis 连接失败，请确保您已经启动了 Redis 服务器: %v\n", err)
		return
	}

	fmt.Println("======== Redis 连接成功，准备就绪！========")
}
