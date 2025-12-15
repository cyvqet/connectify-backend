package ioc

import (
	"time"

	"github.com/cyvqet/connectify/internal/web"
	"github.com/cyvqet/connectify/internal/web/middleware"
	"github.com/cyvqet/connectify/pkg/middleware/ratelimit"
	limiter "github.com/cyvqet/connectify/pkg/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRouter(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			// List of allowed origins for CORS
			AllowOrigins: []string{"https://foo.com"},

			// Allowed HTTP methods for CORS requests
			AllowMethods: []string{"GET", "POST"},

			// Request headers that the browser is allowed to send
			AllowHeaders: []string{"Origin", "Authorization", "Content-Type"},

			// Response headers that can be accessed by frontend JavaScript
			ExposeHeaders: []string{"Jwt-Token"},

			// Whether to allow credentials such as cookies or Authorization headers
			// Note: when enabled, AllowOrigins cannot be "*"
			AllowCredentials: true,

			// Custom logic to dynamically allow origins
			// Here: allow the request if Origin == "https://github.com"
			// Priority: this function has higher priority than AllowOrigins
			AllowOriginFunc: func(origin string) bool {
				return origin == "https://github.com"
			},

			// Cache duration for preflight (OPTIONS) requests in the browser
			MaxAge: 12 * time.Hour,
		}),

		// Rate limiting: allow up to 100 requests per minute
		ratelimit.NewBuilder(limiter.NewRedisSlideWindowLimiter(redisClient, time.Minute, 100)).Build(),

		// JWT login middleware
		// Ignore authentication for the following paths
		middleware.NewLoginJwtMiddlewareBuilder().
			IgnorePath("/user/login_jwt").
			IgnorePath("/user/signup").
			IgnorePath("/user/send_sms_code").
			IgnorePath("/user/login_sms").
			Build(),
	}
}
