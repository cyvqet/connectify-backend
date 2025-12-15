package ratelimit

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/cyvqet/connectify/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	prefix  string
	limiter ratelimit.Limiter
}

// NewBuilder creates a Builder instance
func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		limiter: limiter,
		prefix:  "ip-limiter", // Default prefix
	}
}

// Prefix sets the Redis key prefix
func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build creates a Gin rate limiting middleware
func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			// Conservative approach (rate limiting) vs aggressive approach (allowing through)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

// limit performs rate limiting checks
func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limiter.Limit(ctx, key)
}
