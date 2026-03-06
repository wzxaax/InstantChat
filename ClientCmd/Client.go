package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Client struct {
	ServerIp   string
	ServerPort int
	conn       net.Conn
	Name       string
	Mode       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Mode:       -1,
	}

	// 连接 Server, 如果连接失败返回 nil Client
	conn, err := net.Dial("tcp", serverIp+":"+strconv.Itoa(serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	return client
}

// 处理 Server 回应的消息, 直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 不断地从网络连接 client.conn 读取数据，直到 EOF 或 发生错误
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	mode := -1

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&mode)
	if err != nil {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		var discard string
		fmt.Scanln(&discard)
		return false
	}

	if mode >= 0 && mode <= 3 {
		client.Mode = mode
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.Mode != 0 {
		for client.menu() != true {
		}

		// 根据当前模式处理不同的业务
		switch client.Mode {
		case 1:
			// 公聊模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.UpdateName()
			break
		}
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容, exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入消息内容, exit退出:")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发给服务器

		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) UpdateName() bool {

	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败，退出...")
		return
	}

	fmt.Println("连接服务器成功!")
	go client.DealResponse()
	client.Run()
}
