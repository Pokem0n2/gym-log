package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// GenerateToken 生成 JWT（有效期 10 年）
func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().AddDate(10, 0, 0).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthRequired 认证中间件：验证 JWT，将用户 DB 注入 context
func AuthRequired(store *repository.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
			c.Abort()
			return
		}

		// 获取用户业务数据库
		userDB, err := store.GetUserDB(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库加载失败"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("userDB", userDB)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// 优先从 Cookie 获取
	if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
		return cookie
	}
	// 其次从 Authorization header 获取
	bearer := c.GetHeader("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	return ""
}

// GetUserDB 从 gin.Context 取出用户 DB
func GetUserDB(c *gin.Context) *repository.DB {
	v, exists := c.Get("userDB")
	if !exists {
		return nil
	}
	return v.(*repository.DB)
}

// GetUserID 从 gin.Context 取出用户 ID
func GetUserID(c *gin.Context) string {
	v, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return v.(string)
}
