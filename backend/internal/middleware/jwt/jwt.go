package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"feedsystem_video_go/internal/account"
	"feedsystem_video_go/internal/auth"
	rediscache "feedsystem_video_go/internal/middleware/redis"

	"github.com/gin-gonic/gin"
)

// JWTAuth 使用 JWT + Redis 黑名单 的现代验证方式
func JWTAuth(accountRepo *account.AccountRepository, cache *rediscache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		tokenString := parts[1]

		// 第一步：解析 JWT（仅验证签名和过期时间）
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// 第二步：Redis 黑名单检查
		if cache != nil {
			blacklistKey := fmt.Sprintf("jwt:blacklist:%d", claims.AccountID)

			cacheCtx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
			defer cancel()

			// 如果 Token 在黑名单中，则拒绝访问
			if exists, _ := cache.Exists(cacheCtx, blacklistKey); exists {
				// 可选：进一步验证是否是当前 Token 被拉黑（更严格）
				if val, err := cache.GetBytes(cacheCtx, blacklistKey); err == nil && string(val) == tokenString {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
					return
				}
			}
		}

		// 验证通过，设置用户信息到上下文
		c.Set("accountID", claims.AccountID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// SoftJWTAuth 软认证（可选登录），不强制要求 Token
func SoftJWTAuth(accountRepo *account.AccountRepository, cache *rediscache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		tokenString := parts[1]

		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Redis 黑名单检查（Soft 模式下也建议检查）
		if cache != nil {
			blacklistKey := fmt.Sprintf("jwt:blacklist:%d", claims.AccountID)

			cacheCtx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
			defer cancel()

			if exists, _ := cache.Exists(cacheCtx, blacklistKey); exists {
				if val, err := cache.GetBytes(cacheCtx, blacklistKey); err == nil && string(val) == tokenString {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
					return
				}
			}
		}

		c.Set("accountID", claims.AccountID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetAccountID 从上下文获取用户ID
func GetAccountID(c *gin.Context) (uint, error) {
	uidValue, exists := c.Get("accountID")
	if !exists {
		return 0, errors.New("accountID not found")
	}

	accountID, ok := uidValue.(uint)
	if !ok {
		return 0, errors.New("accountID has invalid type")
	}

	return accountID, nil
}