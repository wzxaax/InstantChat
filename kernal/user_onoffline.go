package kernal

func (user *User) Online() {
	// 1. 将用户加入本地 OnlineMap
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 2. 同步到 Redis 全局在线名单
	RedisClient.SAdd(ctx, OnlineUsersKey, user.Name)

	// 3. 广播当前	// 用户上线消息广播
	BroadCast(user, "已上线")
}

func (user *User) Offline() {
	// 1. 从本地 OnlineMap 中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	// 2. 从 Redis 全局在线名单中移除
	RedisClient.SRem(ctx, OnlineUsersKey, user.Name)

	// 广播用户下线消息
	BroadCast(user, "下线")
}
