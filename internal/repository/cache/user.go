package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cyvqet/connectify/internal/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
}

type redisUserCache struct {
	client redis.Cmdable
	expire time.Duration
}

func NewUserCache(client redis.Cmdable) UserCache {
	return &redisUserCache{
		client: client,
		expire: time.Minute * 10,
	}
}

func (c *redisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	data, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, ErrKeyNotExist
	}
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (c *redisUserCache) Set(ctx context.Context, user domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(user.Id), data, c.expire).Err()
}

func (c *redisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
