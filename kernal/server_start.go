package kernal

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	// 服务器的 ip+端口
	ip   string
	port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
}

func NewServer(ip string, port int) *Server {
	return &Server{
		ip:        ip,
		port:      port,
		OnlineMap: make(map[string]*User),
	}
}

func (s *Server) Start() {
	// 在指定 ip+端口 上创建 TCP 监听 Socket, 等待客户端连接
	listener, err := net.Listen("tcp", s.ip+":"+strconv.Itoa(s.port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	defer listener.Close()

	// 服务器接收私聊消息, 若发现有发往本地在线用户的消息, 进行转发
	go s.ListenPrivateMessage()
	// 服务器接收广播消息, 一旦有消息, 转发给所有本地在线用户
	go s.ListenBroadcastMessage()

	// 服务器持续接受新连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}

		go func(c net.Conn) {
			// 获取用户名
			var username string
			if username, err = getInitialUserName(c); err != nil {
				data, _ := json.Marshal(map[string]interface{}{
					"type": "connectFail",
					"msg":  "登录失败",
				})
				WritePack(c, []byte(data))
				return
			}
			// 创建 User 对象、用户上线、监听发给用户的消息、接收用户发出的消息
			s.userHandler(username, c)
		}(conn)
	}
}

func getInitialUserName(conn net.Conn) (string, error) {
	data, err := ReadPack(conn)
	if err != nil {
		return "", err
	}

	username := strings.TrimSpace(string(data))
	return username, nil
}
