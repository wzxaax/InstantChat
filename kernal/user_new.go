package kernal

import (
	"encoding/json"
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
	// 默认用客户端地址当作用户名
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	_, ok := server.OnlineMap[username]
	if ok {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "userName_alreadyUsed",
			"msg":  "此用户名已被使用",
		})
		conn.Write(data)
		time.Sleep(1 * time.Second)
		close(user.C)
		conn.Close()
		return nil
	}

	if username != "" {
		user.Name = username
	}

	// 监听用户的 channel, 一旦有消息, 直接通过 conn 发送给客户端
	go user.ListenMessage()

	return user
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C

		user.conn.Write([]byte(msg))
		// 中文Windows的cmd使用GBK编码
		// 服务端的消息默认使用UTF-8编码，需强制统一编码
		// gbkMsg, _ := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(msg))
		// user.conn.Write(append(gbkMsg, '\r', '\n'))
	}
}
