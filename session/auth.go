package session

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"main/lib"
	"net/http"
	"time"
)

type TokenType string

const TokenTypeAuth TokenType = "AUTH"
const TokenTypeRefresh TokenType = "REFRESH"

type User struct {
	ID                string
	Email             string
	Password          string
	Verified          bool
	VerificationToken string
	CreatedAt         time.Time
}

type RefreshToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

var users = make(map[string]*User)
var refreshTokens = make(map[string]*RefreshToken)

type Claims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	jwt.RegisteredClaims
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return lib.GetDotEnv("JWT_SECRET"), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token.Claims.(*Claims), nil
}

func GenerateToken(userID string, expires time.Time) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	fmt.Println(lib.GetDotEnv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(lib.GetDotEnv("JWT_SECRET"))
}

func AuthMiddleware(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("token")
		if tokenString == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := ParseToken(tokenString)
		if err != nil || (strict && time.Now().After(claims.ExpiresAt)) {
			// TODO use refresh token to generate new access token instead
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func AuthorizeHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		panic("Failed to hash password")
	}
	user, exists := users[req.Email]
	if !exists || user.Password != string(hashedPassword) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	accessTokenSessionTime := lib.GetIntDotEnv("ACCESS_TOKEN_SESSION_MINUTES")
	refreshTokenSessionTime := lib.GetIntDotEnv("REFRESH_TOKEN_SESSION_HOURS")
	accessToken, _ := GenerateToken(user.ID, time.Now().Add(time.Duration(accessTokenSessionTime)*time.Minute))
	refreshToken, _ := GenerateToken(user.ID, time.Now().Add(time.Duration(refreshTokenSessionTime)*time.Hour))

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"token":                  accessToken,
			"refresh_token":          refreshToken,
			"token_valid_to":         time.Now().Add(15 * time.Minute).Format(time.RFC3339),
			"refresh_token_valid_to": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
	})
}

func SessionHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	refreshToken := c.GetHeader("refresh_token")

	storedToken, exists := refreshTokens[refreshToken]
	if !exists || storedToken.UserID != userID.(string) || time.Now().After(storedToken.ExpiresAt) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"token_valid_to":         time.Now().Add(15 * time.Minute).Format(time.RFC3339),
			"refresh_token_valid_to": storedToken.ExpiresAt.Format(time.RFC3339),
		},
	})
}

func EmailVerifyHandler(c *gin.Context) {
	//verifyToken := c.Param("verifyToken")
	// TODO check if verifyToken is correct and get user by it
	// TODO set up user as verified
	user := User{}
	accessTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("ACCESS_TOKEN_SESSION_MINUTES")) * time.Minute)
	refreshTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("REFRESH_TOKEN_SESSION_HOURS")) * time.Hour)
	accessToken, _ := GenerateToken(user.ID, accessTokenSessionTime)
	refreshToken, _ := GenerateToken(user.ID, refreshTokenSessionTime)

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"access_token":           accessToken,
			"refresh_token":          refreshToken,
			"access_token_valid_to":  accessTokenSessionTime,
			"refresh_token_valid_to": refreshTokenSessionTime,
		},
	})
}

func RegisterHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		panic("Failed to hash password")
	}
	user, exists := users[req.Email]
	if !exists || user.Password != string(hashedPassword) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	verificationToken := uuid.New().String()

	user = &User{
		ID:                uuid.New().String(),
		Email:             req.Email,
		Password:          string(hashedPassword),
		Verified:          false,
		VerificationToken: verificationToken,
		CreatedAt:         time.Now(),
	}
	// TODO add user to database
	// TODO send verification email
	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"VerificationToken": verificationToken,
			"message":           "User registered successfully",
		},
	})
}
