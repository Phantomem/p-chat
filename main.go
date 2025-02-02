package main

import (
	"github.com/joho/godotenv"
	"main/session"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	err := godotenv.Load()
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
