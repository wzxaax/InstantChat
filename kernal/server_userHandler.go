package kernal

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func (s *Server) userHandler(username string, conn net.Conn) {
	// 调试信息
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Printf("新连接建立 | %s | 本地端口:%d | 协议:%s\n",
		remoteAddr,
		s.port,
		conn.RemoteAddr().Network())

	// 创建新的 User 对象
	User := NewUser(username, conn, s)
	if User == nil {
		return
	}

	// 用户上线
	User.Online()

	// 监听用户的 channel, 一旦有消息, 直接通过 tcp conn 发送给客户端
	go User.ListenMessage()

	// 此 channel 用于监测用户是否活跃
	isLive := make(chan bool)

	// 开启新协程，来接收用户发送的消息、执行相应处理
	go ReceiveClientMessage(User, isLive)

	// select {} 会永久阻塞 Handler goroutine

	// 如果用户长时间没有活动（没有发送有效消息），强制断开连接
	monitorUserActive(isLive, User)
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		// 使用 WritePack 向TCP连接中写入一个带长度前缀的数据包
		WritePack(user.conn, []byte(msg))
	}
}

func ReceiveClientMessage(user *User, isLive chan bool) {
	remoteAddr := user.conn.RemoteAddr().(*net.TCPAddr)

	for {
		// 使用 ReadPack 从TCP连接中读取一个数据包，从中提取用户数据
		data, err := ReadPack(user.conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("连接关闭 | %s | 客户端正常断开\n", remoteAddr)
			} else {
				fmt.Printf("读取错误 | %s | 连接已被强制关闭\n", remoteAddr)
			}
			return
		}

		line := strings.TrimSpace(string(data))

		// 针对消息进行处理
		user.DoMessage(line)

		if line != "who" {
			// 用户的任意有效消息，代表当前用户是活跃的
			isLive <- true
		}
	}
}

func monitorUserActive(isLive chan bool, user *User) {
	timer := time.NewTimer(1 * time.Minute) // 超时时长为1分钟
	defer timer.Stop()

	for {
		select {
		case <-isLive:
			if !timer.Stop() { // 若计时器已触发或已停止, timer.Stop()==false, 此时必须排空旧事件
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(1 * time.Minute) // 安全重置计时器

		case <-timer.C:
			data, _ := json.Marshal(map[string]interface{}{
				"type": "connectTimeout",
				"msg":  "连接超时，您已被踢出",
			})
			user.C <- string(data)
			time.Sleep(1 * time.Second)
			// 用户上线
			user.Offline()
			close(user.C)
			user.conn.Close()
			return
		}
	}
}
