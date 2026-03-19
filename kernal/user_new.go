package kernal

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(username string, conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 1. 先检查 Redis 全局在线名单
	exists, err := RedisClient.SIsMember(ctx, OnlineUsersKey, username).Result()
	if err != nil {
		fmt.Printf("Redis 查询失败: %v\n", err)
	}

	// 2. 只有无重名，才允许继续
	_, ok := server.OnlineMap[username] // 保留本地检查作为双重保障
	if ok || exists {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "userName_alreadyUsed",
			"msg":  "此用户名已被使用",
		})
		WritePack(conn, data)
		time.Sleep(1 * time.Second)
		close(user.C)
		conn.Close()
		return nil
	}

	// 设置用户名
	user.Name = username

	return user
}
