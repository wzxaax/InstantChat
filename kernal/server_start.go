package kernal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	ip   string
	port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 此 channel 用于消息广播
	BroadcastMessage chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		ip:               ip,
		port:             port,
		OnlineMap:        make(map[string]*User),
		BroadcastMessage: make(chan string),
	}
}

func (s *Server) Start() {
	// 在指定 ip 和 端口 上创建 Tcp 监听 socket, 等待客户端连接
	listener, err := net.Listen("tcp", s.ip+":"+strconv.Itoa(s.port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	defer listener.Close()

	// 监听广播消息的 channel, 一旦有消息, 发送给所有在线用户
	go s.ListenBroadcastMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}

		username := getInitialUserName(conn)

		// go s.testHandler(conn)
		go s.Handler(username, conn)
	}
}

func getInitialUserName(conn net.Conn) string {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	reader := bufio.NewReader(conn)
	var opErr *net.OpError

	// 读取第一条消息作为 username
	username, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Printf("连接关闭 | %s | 客户端正常断开\n", remoteAddr)
		} else if errors.As(err, &opErr) {
			fmt.Printf("连接关闭 | %s | 本地主动关闭\n", remoteAddr)
		} else {
			fmt.Printf("读取错误 | %s | 错误:%v\n", remoteAddr, err)
		}
		return ""
	}
	username = strings.TrimSpace(username)
	fmt.Println("用户连接:", username)
	return username

	/*
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("读取 username 出错:", err)
			return ""
		}
		username := string(buf[:n])
		username = strings.TrimSpace(username)
		fmt.Println("用户连接:", username)
		return username
	*/
}
