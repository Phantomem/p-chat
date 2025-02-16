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
