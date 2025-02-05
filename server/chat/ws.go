package chat

import (
	chat "main/chat/types"
	"main/state/entity"

	"github.com/gorilla/websocket"
)

var wsConnectionsMap chat.ChatRoomConnState

func AssignConnection(chatRoom entity.ChatRoom, connectionString *websocket.Conn, userID string) error {
	wsConnectionsMap[chatRoom.ID][userID] = connectionString
	return nil
}

func CloseConnection(chatRoom entity.ChatRoom, userID string) error {
	wsConnectionsMap[chatRoom.ID][userID] = nil
	return nil
}
