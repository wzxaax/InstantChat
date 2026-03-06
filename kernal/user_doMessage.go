package kernal

import (
	"encoding/json"
	"sort"
	"strings"
)

// 用户处理发出的消息的业务 <根据消息内容决定行为>
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		user.server.mapLock.RLock()
		var users []string
		for _, onlineUser := range user.server.OnlineMap {
			users = append(users, onlineUser.Name)
		}
		user.server.mapLock.RUnlock()

		sort.Strings(users)
		// 把用户列表序列化成 JSON
		data, _ := json.Marshal(map[string]interface{}{
			"type":  "userList",
			"users": users,
		})
		user.C <- string(data)

		//user.server.mapLock.Lock()
		//for _, onlineUser := range user.server.OnlineMap {
		//	onlineMsg := "[" + onlineUser.Addr + "]" + onlineUser.Name + ":" + "在线..."
		//	user.C <- onlineMsg
		//}
		//user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 修改用户名消息格式: rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "userRename_alreadyUsed",
				"msg":  "此用户名已被使用",
			})
			user.C <- string(data)
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			data, _ := json.Marshal(map[string]interface{}{
				"type": "userRename_success",
				"msg":  "您已经更新用户名:" + user.Name,
			})
			user.C <- string(data)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 私聊消息格式   to|张三|消息内容
		// 1 获取对方的用户名
		parts := strings.Split(msg, "|")
		if len(parts) < 3 {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "privateChat_badRequest",
				"msg":  "消息格式不正确，请使用 \"to|张三|你好啊\" 格式",
			})
			user.C <- string(data)
			return
		}
		remoteName := parts[1]
		// 2 根据用户名 得到对方 User 对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "privateChat_noSuchUser",
				"msg":  "该用户不存在",
			})
			user.C <- string(data)
			return
		}
		// 3 获取消息内容，通过对方的 User 对象将消息内容发送过去
		content := parts[2]
		if content == "" {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "privateChat_noMessageContent",
				"msg":  "无消息内容，请重发",
			})
			user.C <- string(data)
			return
		}
		data, _ := json.Marshal(map[string]interface{}{
			"type": "privateChat_content",
			"msg":  user.Name + "对您说:" + content,
		})
		remoteUser.C <- string(data)
	} else {
		// 广播
		user.server.BroadCast(user, msg)
	}
}
