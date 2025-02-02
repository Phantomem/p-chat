package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generate token", func(*testing.T) {
		userID := "test-user-id"
		expires := time.Now().Add(15 * time.Minute)

		token, err := GenerateToken(userID, expires)
		fmt.Println(token)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestParseToken(t *testing.T) {
	t.Run("generate token", func(*testing.T) {

		userID := "test-user-id"
		expires := time.Now().Add(15 * time.Minute)

		token, err := GenerateToken(userID, expires)
		assert.NoError(t, err)

		claims, err := ParseToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

}

//func TestAuthMiddleware(t *testing.T) {
//	t.Run("generate token", func(*testing.T) {
//
//		gin.SetMode(gin.TestMode)
//		r := gin.New()
//		r.Use(AuthMiddleware(false))
//
//		r.GET("/test", func(c *gin.Context) {
//			c.String(200, "Success")
//		})
//
//		w := performRequest(r, "GET", "/test", "")
//
//		assert.Equal(t, 401, w.Code)
//	})
//
//}

func performRequest(r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("token", token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
