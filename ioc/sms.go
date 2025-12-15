package ioc

import (
	"connectify/internal/service/sms"
	"connectify/internal/service/sms/aliyun"
	"connectify/internal/service/sms/failover"
	"connectify/internal/service/sms/ratelimit"
	"connectify/internal/service/sms/tencent"
	limiter "connectify/pkg/ratelimit"
	"time"

	"github.com/redis/go-redis/v9"
)

// It directly returns a single provider (Tencent Cloud SMS) without
// any additional protection such as rate limiting or failover.
func InitSmsService(redisClient redis.Cmdable) sms.Service {
	// Simple and direct provider usage.
	// Suitable for demos or scenarios where high availability is not required.
	return tencent.NewService("appId", "signName")
}

// Before sending an SMS, the rate limiter is checked.
// If the rate limit is exceeded, the request is rejected immediately.
func InitSmsRatelimitService(redisClient redis.Cmdable) sms.Service {
	return ratelimit.NewService(
		// The actual SMS provider implementation
		tencent.NewService("appId", "signName"),

		// Redis-based sliding window rate limiter
		// Allows up to 100 requests per minute (globally)
		limiter.NewRedisSlideWindowLimiter(redisClient, time.Minute, 100),
	)
}

// This function wraps multiple SMS providers with a failover strategy.
// Providers are tried one by one until one succeeds or all fail.
func InitSmsFailoverService(redisClient redis.Cmdable) sms.Service {
	return failover.NewService(
		[]sms.Service{
			tencent.NewService("appId", "signName"),
			aliyun.NewService("appId", "signName"),
		},
	)
}

// This function wraps multiple SMS providers with a timeout-based failover strategy.
func InitSmsFailoverTimeoutService(redisClient redis.Cmdable) sms.Service {
	return failover.NewTimeoutService(
		[]sms.Service{
			tencent.NewService("appId", "signName"),
			aliyun.NewService("appId", "signName"),
		},
	)
}
