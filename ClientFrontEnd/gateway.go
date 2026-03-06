package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	conn, _ := upgrader.Upgrade(w, r, nil)

	// 连接 TCP 聊天服务器
	tcpConn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"type": "connectFail",
			"msg":  "无法连接到服务器",
		})
		conn.WriteMessage(websocket.TextMessage, data)
		conn.Close()
		return
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type": "connectSuccess",
		"msg":  "成功连接到服务器",
	})
	conn.WriteMessage(websocket.TextMessage, data)

	// 直接在 TCP 连接上发送用户名（握手）
	tcpConn.Write([]byte(username + "\n"))

	// 从 TCP -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := tcpConn.Read(buf)
			if err != nil {
				conn.Close()
				return
			}
			conn.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	// 从 WebSocket -> TCP
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		tcpConn.Write(msg)
	}
}

func main() {
	http.HandleFunc("/ws", handleWS)
	fmt.Println("WebSocket 网关启动于 ws://localhost:8080/ws")
	http.ListenAndServe(":8080", nil)
}
