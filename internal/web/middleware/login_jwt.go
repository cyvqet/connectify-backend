package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cyvqet/connectify/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Token TTL: 30 minutes, refresh when remaining 5 minutes
const (
	refreshWhen = 5 * time.Minute  // Refresh when remaining 5 minutes
	newTokenTTL = 30 * time.Minute // New token TTL is still 30 minutes
)

type LoginJwtMiddlewareBuilder struct {
	paths []string
}

func NewLoginJwtMiddlewareBuilder() *LoginJwtMiddlewareBuilder {
	return &LoginJwtMiddlewareBuilder{}
}

func (l *LoginJwtMiddlewareBuilder) IgnorePath(paths string) *LoginJwtMiddlewareBuilder {
	l.paths = append(l.paths, paths)
	return l
}

func (l *LoginJwtMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if slices.Contains(l.paths, ctx.Request.RequestURI) {
			ctx.Next()
			return
		}

		// JWT verification logic
		// Get JWT tokenHeader from request header, validate it
		// If validation fails, return 401 unauthorized error
		// If validation succeeds, call ctx.Next() to continue processing the request
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未授权"})
			return
		}

		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 || segs[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未授权"})
			return
		}
		claim := web.UserClaims{} // Custom Claims structure
		token := segs[1]
		jwtToken, err := jwt.ParseWithClaims(token, &claim, func(t *jwt.Token) (any, error) {
			return []byte("secret"), nil
		})

		if err != nil || !jwtToken.Valid { // Token expired, return false
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未授权"})
			return
		}

		if claim.UserAgent != ctx.Request.UserAgent() { // Check if UserAgent is consistent
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未授权"})
			return
		}

		remaining := time.Until(claim.ExpiresAt.Time)
		fmt.Printf("token remaining time: %v\n", remaining)
		if remaining <= refreshWhen {
			fmt.Printf("token will expire in %v, refresh token\n", remaining)
			// Refresh token expiration time
			claim.ExpiresAt = jwt.NewNumericDate(time.Now().Add(newTokenTTL))

			// Generate new JWT token
			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
			newTokenStr, err := newToken.SignedString([]byte("secret"))
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "system error"})
				return
			}

			// Return new token to client
			ctx.Header("Jwt-Token", newTokenStr)
		}

		// Store claim in context for subsequent processing
		ctx.Set("claim", claim)

		ctx.Next()
	}
}
