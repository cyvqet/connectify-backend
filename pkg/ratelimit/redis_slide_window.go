package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaScript string

type RedisSlideWindowLimiter struct {
	cmd      redis.Cmdable // redis client
	interval time.Duration // time window length
	rate     int           // maximum number of requests allowed within the window
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlideWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (l *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return l.cmd.Eval(ctx, luaScript, []string{key},
		l.interval.Milliseconds(), l.rate, time.Now().UnixMilli()).Bool()
}
