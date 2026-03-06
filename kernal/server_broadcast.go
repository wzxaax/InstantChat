package kernal

import "encoding/json"

func (s *Server) ListenBroadcastMessage() {
	for {
		msg := <-s.BroadcastMessage

		data, _ := json.Marshal(map[string]interface{}{
			"type": "Broadcast",
			"msg":  msg,
		})

		s.mapLock.RLock()
		for _, client := range s.OnlineMap {
			client.C <- string(data)
		}
		s.mapLock.RUnlock()
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := user.Name + ":" + msg
	s.BroadcastMessage <- sendMsg
}
