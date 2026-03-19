package kernal

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// 用户处理发出的消息的业务 <根据消息内容决定行为>
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户信息 (从 Redis 获取全局在线名单)
		users, err := RedisClient.SMembers(ctx, OnlineUsersKey).Result()
		if err != nil {
			fmt.Printf("Redis SMembers 失败: %v\n", err)
			return
		}

		// 把在线用户列表序列化成 JSON，发送给客户端
		sort.Strings(users)
		data, _ := json.Marshal(map[string]interface{}{
			"type":  "userList",
			"users": users,
		})
		user.C <- string(data)
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 修改用户名消息格式 --  rename|张三
		newName := strings.Split(msg, "|")[1]
		// 1. 先检查 Redis 全局在线名单
		exists, err := RedisClient.SIsMember(ctx, OnlineUsersKey, newName).Result()
		if err != nil {
			fmt.Printf("Redis 查询失败: %v\n", err)
		}
		// 2. 检查本地 Map (仅作冗余检查)
		_, ok := user.server.OnlineMap[newName]
		if ok || exists {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "userRename_alreadyUsed",
				"msg":  "此用户名已被使用",
			})
			user.C <- string(data)
		} else {
			// 3. 可以改名，更新 Redis 的全局在线名单
			// 创建一个Redis事务管道  批量执行多条Redis命令
			// 原子性删除旧名，添加新名
			pipe := RedisClient.TxPipeline()
			pipe.SRem(ctx, OnlineUsersKey, user.Name)
			pipe.SAdd(ctx, OnlineUsersKey, newName)
			_, err := pipe.Exec(ctx)
			if err != nil {
				fmt.Printf("Redis 更新用户名失败: %v\n", err)
				return
			}
			// 4. 更新本地 Map
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			// 5. 完成改名
			user.Name = newName
			data, _ := json.Marshal(map[string]interface{}{
				"type": "userRename_success",
				"msg":  "您已经更新用户名:" + user.Name,
			})
			user.C <- string(data)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 私聊消息格式 --  to|张三|消息内容
		// 1. 获取对方的用户名
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
		// 2. 获取消息内容
		content := parts[2]
		if content == "" {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "privateChat_noMessageContent",
				"msg":  "消息内容不能为空",
			})
			user.C <- string(data)
			return
		}
		// 3. 检查目标用户是否存在 (从 Redis 全局在线名单查询)
		exists, err := RedisClient.SIsMember(ctx, OnlineUsersKey, remoteName).Result()
		if err != nil {
			fmt.Printf("Redis 查询目标用户失败: %v\n", err)
			return
		}
		if !exists {
			data, _ := json.Marshal(map[string]interface{}{
				"type": "privateChat_noSuchUser",
				"msg":  "该用户不存在或已下线",
			})
			user.C <- string(data)
			return
		}
		// 4. 向统一的 Kafka 私聊频道投递消息
		data, _ := json.Marshal(map[string]interface{}{
			"type":   "privateChat_content",
			"msg":    user.Name + "对您说:" + content,
			"target": remoteName, // 必须附带目标用户名，以便各节点确定转发行为
		})
		// 投递私聊消息
		Private(remoteName, data)
	} else {
		// 广播
		// 向统一的 Kafka 广播频道投递消息
		BroadCast(user, msg)
	}
}
