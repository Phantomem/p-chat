package session

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"main/lib"
	"main/state"
	"main/state/entity"
	"net/http"
	"time"
)

type TokenType string

const TokenTypeAuth TokenType = "AUTH"
const TokenTypeRefresh TokenType = "REFRESH"

type RefreshToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

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
		accessToken := c.GetHeader("access_token")
		refreshToken := c.GetHeader("refresh_token")
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := ParseToken(accessToken)
		if err != nil || (strict && time.Now().After(claims.ExpiresAt)) {
			refreshClaims, err := ParseToken(refreshToken)
			if err != nil || (time.Now().After(refreshClaims.ExpiresAt)) {
				c.AbortWithStatus(http.StatusUnauthorized)
			}

			accessTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("ACCESS_TOKEN_SESSION_MINUTES")) * time.Minute)
			refreshTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("REFRESH_TOKEN_SESSION_HOURS")) * time.Hour)
			newAccessToken, _ := GenerateToken(claims.UserID, accessTokenSessionTime)
			newRefreshToken, _ := GenerateToken(claims.UserID, refreshTokenSessionTime)

			err = state.Update(state.GetConnection(), &entity.UserSession{UserID: claims.UserID, AccessToken: newAccessToken, RefreshToken: newRefreshToken})

			c.Header("access_token", newAccessToken)
			c.Header("refresh_token", newRefreshToken)
			c.Header("access_token_valid_to", accessTokenSessionTime.String())
			c.Header("refresh_token_valid_to", refreshTokenSessionTime.String())
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
	//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var user = entity.User{}
	err := state.GetByKeyVal[entity.User, string](state.GetConnection(), "email", req.Email, &user)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))

	if user.Verified != true || user.Password == "" || err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	accessTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("ACCESS_TOKEN_SESSION_MINUTES")) * time.Minute)
	refreshTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("REFRESH_TOKEN_SESSION_HOURS")) * time.Hour)
	accessToken, _ := GenerateToken(user.ID, accessTokenSessionTime)
	refreshToken, _ := GenerateToken(user.ID, refreshTokenSessionTime)

	err = state.Update(state.GetConnection(), &entity.UserSession{UserID: user.ID, AccessToken: accessToken, RefreshToken: refreshToken})

	c.Header("access_token", accessToken)
	c.Header("refresh_token", refreshToken)
	c.Header("access_token_valid_to", accessTokenSessionTime.String())
	c.Header("refresh_token_valid_to", refreshTokenSessionTime.String())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"access_token":           accessToken,
			"refresh_token":          refreshToken,
			"token_valid_to":         accessTokenSessionTime,
			"refresh_token_valid_to": refreshTokenSessionTime,
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
	verifyToken := c.Query("verifyToken")
	fmt.Println(verifyToken)
	var user = entity.User{}
	err := state.GetByKeyVal[entity.User, string](state.GetConnection(), "verification_token", verifyToken, &user)
	if err != nil {
		fmt.Println(1)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	user.Verified = true
	err = state.Update[entity.User](state.GetConnection(), &user)
	if err != nil {
		fmt.Println(2)

		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	accessTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("ACCESS_TOKEN_SESSION_MINUTES")) * time.Minute)
	refreshTokenSessionTime := time.Now().Add(time.Duration(lib.GetIntDotEnv("REFRESH_TOKEN_SESSION_HOURS")) * time.Hour)
	accessToken, _ := GenerateToken(user.ID, accessTokenSessionTime)
	refreshToken, _ := GenerateToken(user.ID, refreshTokenSessionTime)

	err = state.Create(state.GetConnection(), &entity.UserSession{UserID: user.ID, AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.Header("access_token", accessToken)
	c.Header("refresh_token", refreshToken)
	c.Header("access_token_valid_to", accessTokenSessionTime.String())
	c.Header("refresh_token_valid_to", refreshTokenSessionTime.String())

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
	var user = entity.User{}
	_ = state.GetByKeyVal[entity.User, string](state.GetConnection(), "email", req.Email, &user)
	if user.ID != "" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"status":  "Error",
			"message": "Failed to register user",
		})
		return
	}

	verificationToken := uuid.New().String()

	user = entity.User{
		ID:                uuid.New().String(),
		Email:             req.Email,
		Password:          string(hashedPassword),
		Verified:          false,
		VerificationToken: verificationToken,
		CreatedAt:         time.Now(),
	}
	err = state.Create[entity.User](state.GetConnection(), &user)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"status":  "Error",
			"message": "Failed to register user",
		})
		return
	}
	// TODO send verification email
	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"data": gin.H{
			"VerificationToken": verificationToken,
			"message":           "User registered successfully",
		},
	})
}
