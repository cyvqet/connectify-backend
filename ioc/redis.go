package ioc

import (
	"github.com/spf13/viper"

	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	type RedisConfig struct {
		Addr string `yaml:"addr"`
	}
	var redisConfig RedisConfig
	err := viper.UnmarshalKey("redis", &redisConfig)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
	})
}
