package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"main/lib"
	"main/session"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Adjust this in production
		},
	}
	clients      = make(map[*websocket.Conn]string)
	clientsMutex sync.Mutex
)

func main() {
	r := gin.Default()
	err := godotenv.Load()

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

	if err != nil {
		panic("Error loading .env file")
	}
	// Public endpoints
	r.GET("/status", statusHandler)
	// Register session endpoints
	r.POST("/session/authorize", session.AuthorizeHandler)
	r.POST("/session/register", session.RegisterHandler)
	r.GET("/session/emailVerify", session.EmailVerifyHandler)
	r.Use(session.AuthMiddleware(false)).GET("/session", session.SessionHandler)
	// chat endpoints
	chat := r.Group("/")
	chat.Use(session.AuthMiddleware(false))
	{
		chat.GET("/chat/messages", messagesHandler)
		chat.PUT("/chat/group/:id/join", joinGroupHandler)
		chat.DELETE("/chat/group/:id/join", leaveGroupHandler)
	}

	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func statusHandler(c *gin.Context) {
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
