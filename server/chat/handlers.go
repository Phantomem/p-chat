package chat

import (
	"main/lib"
	"time"
)

func ChatHandler(task lib.Task[map[string]any]) {
	//fmt.Printf("task received %s", lib.PrettyPrint(task))
	// Sends message over all connections in the room
	time.Sleep(2 * time.Second)
}
