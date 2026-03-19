package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"InstantChat/kernal"

	"github.com/gorilla/websocket"
)

// 定义一个 WebSocket 的升级器 (Upgrader) 实例
var upgrader = websocket.Upgrader{
	// 关闭了跨域安全检查，允许来自**任何域名（Origin）**的网页客户端连接到这个 WebSocket 服务器
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	// 将常规 HTTP 请求升级为一个全双工通信的 WebSocket 连接对象
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	// 连接 TCP 聊天服务器
	tcpConn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "connectFail",
			"msg":  "无法连接到服务器",
		})
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}
	defer tcpConn.Close()

	data, _ := json.Marshal(map[string]interface{}{
		"type": "connectSuccess",
		"msg":  "成功连接到服务器",
	})
	conn.WriteMessage(websocket.TextMessage, data)

	// 构造初始数据包，携带用户名
	err = kernal.WritePack(tcpConn, []byte(username))
	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "connectFail",
			"msg":  "登录失败",
		})
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	/* 从 TCP -> WebSocket */
	go func() {
		for {
			// 从TCP连接读取数据包
			body, err := kernal.ReadPack(tcpConn)
			if err != nil {
				fmt.Println("未能从服务器获取数据:", err)
				return
			}

			// 将包体作为一整条 WebSocket 消息发出
			err = conn.WriteMessage(websocket.TextMessage, body)
			if err != nil {
				fmt.Println("向 WebSocket 写入数据失败:", err)
				return
			}
		}
	}()

	/* 从 WebSocket -> TCP */
	for {
		// 接收一条 WebSocket 消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("接收 WebSocket 消息失败:", err)
			break
		}

		// 向TCP连接发送数据包
		err = kernal.WritePack(tcpConn, msg)
		if err != nil {
			fmt.Println("消息未能转发到服务器:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWS)
	fmt.Println("WebSocket 网关启动于 ws://localhost:8080/ws")
	http.ListenAndServe(":8080", nil)
}
