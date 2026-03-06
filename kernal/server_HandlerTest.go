package kernal

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

import "golang.org/x/text/encoding/simplifiedchinese"

func (s *Server) testHandler(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Printf("新连接建立 | IP:%s Port:%d | 本地端口:%d | 协议:%s\n",
		remoteAddr.IP,
		remoteAddr.Port,
		s.port,
		conn.RemoteAddr().Network())

	// 连接生命周期管理
	defer func() {
		conn.Close()
		fmt.Printf("连接关闭 | %s\n", remoteAddr)
	}()

	// 发送欢迎信息（包含\r\n）
	conn.Write([]byte("Welcome to Server\r\n> "))

	reader := bufio.NewReader(conn)
	for {
		// 读取直到遇到换行符
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("读取错误 | %s | 错误:%v\n", remoteAddr, err)
			}
			return
		}

		// 去除换行符后处理
		line = strings.TrimSpace(line)
		// 中文Windows的cmd使用GBK编码，需GBK转UTF-8才能在控制台正常显示
		gbkDecoder := simplifiedchinese.GBK.NewDecoder()
		Decodedline, _ := gbkDecoder.String(line)
		fmt.Printf("收到行数据 | %s | 内容:%q\n", remoteAddr, Decodedline)

		// 回显整行数据（直接回显获取到的GBK）
		if _, err := conn.Write([]byte(line + "\r\n> ")); err != nil {
			fmt.Printf("写入错误 | %s | 错误:%v\n", remoteAddr, err)
			return
		}
	}
}
