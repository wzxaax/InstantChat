package kernal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func (s *Server) Handler(username string, conn net.Conn) {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Printf("新连接建立 | %s | 本地端口:%d | 协议:%s\n",
		remoteAddr,
		s.port,
		conn.RemoteAddr().Network())

	User := NewUser(username, conn, s)
	if User == nil {
		return
	}

	User.Online()

	// 此 channel 用于监听用户是否活跃
	isLive := make(chan bool)

	go ReceiveClientMessage(User, isLive)

	// select {} 会永久阻塞 Handler goroutine
	// select {}

	// 如果用户长时间没有活动（没有发消息），强制断开连接
	monitorUserActive(isLive, User)
}

func ReceiveClientMessage(user *User, isLive chan bool) {
	/*
		buf := make([]byte, 4096)
		for {
			n, err := user.conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(可去除'\n')
			msg := string(buf[:n])

			fmt.Printf("接收到%d字节数据: %s", n, msg)

			// 用户针对msg进行消息处理
			// user.DoMessage(msg)
		}
	*/

	remoteAddr := user.conn.RemoteAddr().(*net.TCPAddr)

	reader := bufio.NewReader(user.conn)
	var opErr *net.OpError
	for {
		// 读取直到遇到换行符
		// TODO 如果刚好消息内容里有\n，我们要对这个字符转义
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("连接关闭 | %s | 客户端正常断开\n", remoteAddr)
			} else if errors.As(err, &opErr) {
				fmt.Printf("连接关闭 | %s | 本地主动关闭\n", remoteAddr)
			} else {
				fmt.Printf("读取错误 | %s | 错误:%v\n", remoteAddr, err)
			}
			return
		}

		// 去除换行符后处理
		line = strings.TrimSpace(line)
		// 中文Windows的cmd使用GBK编码，需GBK转UTF-8才能在控制台正常显示
		// gbkDecoder := simplifiedchinese.GBK.NewDecoder()
		// Decodedline, _ := gbkDecoder.String(line)

		// fmt.Printf("收到行数据 | %s | 内容:%q\n", remoteAddr, line)

		/* 回显整行数据（直接回显获取到的GBK）
		if _, err := user.conn.Write([]byte(line + "\r\n> ")); err != nil {
			fmt.Printf("写入错误 | %s | 错误:%v\n", remoteAddr, err)
			return
		}
		*/

		// 用户针对msg进行消息处理
		user.DoMessage(line)

		// 用户的任意消息，代表当前用户是活跃的
		isLive <- true
	}
}

func monitorUserActive(isLive chan bool, user *User) {
	timer := time.NewTimer(1 * time.Minute)
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
			user.Offline()
			close(user.C)
			user.conn.Close()
			return
		}
	}
}
