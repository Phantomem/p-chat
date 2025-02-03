package chat

import "github.com/gorilla/websocket"

type ChatRoomConnState map[string]map[string]*websocket.Conn
