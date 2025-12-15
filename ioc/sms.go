package ioc

import (
	"connectify/internal/service/sms"
	"connectify/internal/service/sms/ratelimit"
	"connectify/internal/service/sms/tencent"
	limiter "connectify/pkg/ratelimit"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitSMSService(redisClient redis.Cmdable) sms.Service {
	return ratelimit.NewService(
		tencent.NewService("appId", "signName"),
		limiter.NewRedisSlideWindowLimiter(redisClient, time.Minute, 100),
	)
}
