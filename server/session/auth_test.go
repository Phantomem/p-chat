package session

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"main/state"
	"main/state/entity"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	t.Run("TestGenerateToken", func(*testing.T) {
		userID := "test-user-id"
		expires := time.Now().Add(15 * time.Minute)

		token, err := GenerateToken(userID, expires)
		fmt.Println(token)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestParseToken(t *testing.T) {
	t.Run("TestParseToken", func(*testing.T) {

		userID := "test-user-id"
		expires := time.Now().Add(15 * time.Minute)

		token, err := GenerateToken(userID, expires)
		assert.NoError(t, err)

		claims, err := ParseToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Successful Registration", func(t *testing.T) {
		// Mock the state functions
		state.GetByKeyVal = func(db *gorm.DB, key string, val string, entity *entity.User) error {
			return errors.New("record not found")
		}
		state.Create = func(db *gorm.DB, entity *entity.User) error {
			return nil
		}

		// Create a request body
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		// Create a request
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		w := httptest.NewRecorder()

		// Create a gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call the handler
		RegisterHandler(c)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success", response["status"])
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create a request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		w := httptest.NewRecorder()

		// Create a gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call the handler
		RegisterHandler(c)

		// Assert the response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User Already Exists", func(t *testing.T) {
		// Mock the state functions
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		state.GetByKeyVal = func(db *gorm.DB, key string, val string, entity *entity.User) error {
			*entity = entity.User{
				Email:    "test@example.com",
				Password: string(hashedPassword),
			}
			return nil
		}

		// Create a request body
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		// Create a request
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		w := httptest.NewRecorder()

		// Create a gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call the handler
		RegisterHandler(c)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
