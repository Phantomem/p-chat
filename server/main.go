package main

import (
	"fmt"
	"main/chat"
	"main/lib"
	"main/session"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Adjust this in production
		},
	}
)

func main() {
	r := gin.New()

	// Add only the Recovery middleware (to handle panics gracefully)
	r.Use(gin.Recovery())
	err := godotenv.Load()

	lib.InitConfiguration()
	lib.InitMonitor()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	dbHost := lib.GetDotEnv("PSQL_HOST")
	dbUser := lib.GetDotEnv("PSQL_USER")
	dbPassword := lib.GetDotEnv("PSQL_PASSWORD")
	dbName := lib.GetDotEnv("PSQL_DB")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 TimeZone=Europe/Warsaw", dbHost, dbUser, dbPassword, dbName)
	_, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	lib.GetConfig().WP = lib.NewWorkerPool(lib.WorkerPoolConfig{WorkerFn: chat.ChatHandler, NumWorkers: 100000})
	lib.GetConfig().WP.ScaleUp(100)
	defer lib.GetConfig().WP.Shutdown()
	go lib.CollectWorkerPoolMetrics(lib.GetConfig().WP)
	// Public endpoints
	r.GET("/status", statusHandler)
	go r.GET("/ws-upgrade", wsUpgradeHandler)
	// Register session endpoints
	r.POST("/session/authorize", session.AuthorizeHandler)
	r.POST("/session/register", session.RegisterHandler)
	r.GET("/session/emailVerify", session.EmailVerifyHandler)
	r.Use(session.AuthMiddleware(false)).GET("/session", session.SessionHandler)
	// chat endpoints
	authenticated := r.Group("/")
	authenticated.Use(session.AuthMiddleware(false))
	{
		//go authenticated.GET("/ws-upgrade", wsUpgradeHandler)

		authenticated.GET("/chat/messages", messagesHandler)
		authenticated.PUT("/chat/group/:id/join", joinGroupHandler)
		authenticated.DELETE("/chat/group/:id/join", leaveGroupHandler)
	}

	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func statusHandler(c *gin.Context) {
	//lib.GetConfig().WP.EnqueueTask(lib.Task[map[string]any]{Data: map[string]any{"message": "Hello, World!"}})
	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"uptime":  time.Since(time.Now()).String(),
			"version": "1.0.0",
		},
	})
}

// Existing handlers (implement according to your needs)
func messagesHandler(c *gin.Context)   { /* ... */ }
func joinGroupHandler(c *gin.Context)  { /* ... */ }
func leaveGroupHandler(c *gin.Context) { /* ... */ }

func wsUpgradeHandler(c *gin.Context) {
	//accessToken := c.GetHeader("access_token")
	//accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNzZlMGRlZWUtMTI4Yy00ZDdlLWFlMTAtYzJkN2RkMmZlYWEwIiwicm9sZSI6IiIsInRva2VuX3R5cGUiOiIiLCJleHBpcmVzX2F0IjoiMDAwMS0wMS0wMVQwMDowMDowMFoiLCJleHAiOjE3Mzg1Mjg4OTZ9._wKlwEUVwtuJQH4O4DQMKwb1FToiJy05IngEp0hiy1k"
	//var userSession entity.UserSession
	//_ = state.GetByKeyVal[entity.UserSession, string](state.GetConnection(), "access_token", accessToken, &userSession)
	connectionString, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	ids := lib.GetConfig().WP.ScaleUp(1)

	defer func() {
		lib.GetConfig().WP.ScaleDown(ids[0])
		_ = connectionString.Close()
	}()

	connectionString.PingHandler()
	for {
		_, bytes, err := connectionString.ReadMessage()
		if err != nil {
			break
		}
		lib.GetConfig().WP.EnqueueTask(lib.Task[map[string]any]{Data: map[string]any{"message": string(bytes)}})
	}
}

//package main
//
//import (
//"chat-app/config"
//"context"
//"log"
//"net/http"
//"os"
//"os/signal"
//"syscall"
//"time"
//
//"github.com/go-redis/redis/v8"
//"github.com/gorilla/websocket"
//)
//
//var (
//	cfg        *config.Config
//	redisClient *redis.Client
//	upgrader   = websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool { return true },
//	}
//)
//
//func main() {
//	// Load configuration
//	cfg = config.Load()
//
//	// Initialize Redis
//	redisClient = redis.NewClient(&redis.Options{
//		Addr: cfg.RedisAddr,
//	})
//
//	// WebSocket handler
//	http.HandleFunc("/ws", handleWebSocket)
//
//	// Start server
//	server := &http.Server{
//		Addr:         ":" + strconv.Itoa(cfg.ServerPort),
//		ReadTimeout:  10 * time.Second,
//		WriteTimeout: 10 * time.Second,
//	}
//
//	// Graceful shutdown
//	go func() {
//		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//			log.Fatalf("Server error: %v", err)
//		}
//	}()
//
//	// Wait for interrupt signal
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := server.Shutdown(ctx); err != nil {
//		log.Fatal("Server forced to shutdown:", err)
//	}
//
//	if err := redisClient.Close(); err != nil {
//		log.Fatal("Redis connection closure error:", err)
//	}
//}
//
//func handleWebSocket(w http.ResponseWriter, r *http.Request) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Printf("WebSocket upgrade failed: %v", err)
//		return
//	}
//	defer conn.Close()
//
//	// Authenticate user (implement your own logic)
//	userID := authenticateUser(r)
//	if userID == "" {
//		conn.WriteMessage(websocket.CloseMessage, []byte("Unauthorized"))
//		return
//	}
//
//	// Create dedicated PubSub for this connection
//	pubsub := redisClient.Subscribe(r.Context())
//	defer pubsub.Close()
//
//	// Subscribe to user's personal channel
//	err = pubsub.Subscribe(r.Context(), getUserChannel(userID))
//	if err != nil {
//		log.Printf("Subscription failed: %v", err)
//		return
//	}
//
//	go readFromClient(conn, pubsub, userID)
//	writeToClient(conn, pubsub.Channel())
//}
//
//func getUserChannel(userID string) string {
//	return fmt.Sprintf("user:%s", userID)
//}
//
//func readFromClient(conn *websocket.Conn, pubsub *redis.PubSub, userID string) {
//	defer pubsub.Close()
//
//	for {
//		_, msgBytes, err := conn.ReadMessage()
//		if err != nil {
//			break
//		}
//
//		var msg struct {
//			Type    string `json:"type"`
//			To      string `json:"to"`
//			Content string `json:"content"`
//		}
//
//		if err := json.Unmarshal(msgBytes, &msg); err != nil {
//			log.Printf("Invalid message format: %v", err)
//			continue
//		}
//
//		channel := getChannel(msg.Type, userID, msg.To)
//
//		// Add permission checks here (e.g., is user allowed to message this target?)
//		if !validateMessagePermission(userID, msg.Type, msg.To) {
//			log.Printf("Unauthorized message attempt from %s", userID)
//			continue
//		}
//
//		// Publish with sender information
//		envelope := map[string]interface{}{
//			"from":    userID,
//			"to":      msg.To,
//			"content":  msg.Content,
//			"channel":  channel,
//		}
//
//		envelopeBytes, _ := json.Marshal(envelope)
//
//		if err := redisClient.Publish(r.Context(), channel, envelopeBytes).Err(); err != nil {
//			log.Printf("Redis publish error: %v", err)
//		}
//	}
//}
//
//func writeToClient(conn *websocket.Conn, ch <-chan *redis.Message) {
//	for redisMsg := range ch {
//		var envelope map[string]interface{}
//		if err := json.Unmarshal([]byte(redisMsg.Payload), &envelope); err != nil {
//			log.Printf("Invalid message format: %v", err)
//			continue
//		}
//
//		msg := map[string]interface{}{
//			"type":    "message",
//			"from":    envelope["from"],
//			"content": envelope["content"],
//			"channel": envelope["channel"],
//		}
//
//		msgBytes, _ := json.Marshal(msg)
//
//		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
//			break
//		}
//	}
//}
//func handleSubscribe(conn *websocket.Conn, pubsub *redis.PubSub, userID string, channel string) {
//	// Validate group membership before subscribing
//	if isValidGroupMember(userID, channel) {
//		err := pubsub.Subscribe(context.Background(), channel)
//		if err != nil {
//			log.Printf("Subscription failed: %v", err)
//		}
//	}
//}
//
//func handleUnsubscribe(pubsub *redis.PubSub, channel string) {
//	err := pubsub.Unsubscribe(context.Background(), channel)
//	if err != nil {
//		log.Printf("Unsubscription failed: %v", err)
//	}
//}
//func getChannel(msgType, from, to string) string {
//	// For private: always sort IDs to prevent channel duplication
//	ids := []string{from, to}
//	sort.Strings(ids)
//
//	switch msgType {
//	case "private":
//		return fmt.Sprintf("private:%s:%s", ids[0], ids[1])
//	case "group":
//		return fmt.Sprintf("group:%s", to)
//	default:
//		return "global"
//	}
//}
//func validateMessagePermission(senderID, msgType, target string) bool {
//	switch msgType {
//	case "private":
//		// Check if users have a relationship
//		return checkFriendship(senderID, target)
//	case "group":
//		// Check if user is member of group
//		return checkGroupMembership(senderID, target)
//	default:
//		return false
//	}
//}
//
//// Implement these based on your user/group storage
//func checkFriendship(user1, user2 string) bool {
//	// Query your database or relationship store
//	return true
//}
//
//func checkGroupMembership(userID, groupID string) bool {
//	// Query your group membership store
//	return true
//}
//
//# Monitor all messages
//redis-cli psubscribe '*'
//
//# Send test message from console
//redis-cli publish private:user1:user2 '{"from":"user1","to":"user2","content":"Test"}'
