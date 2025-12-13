package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string

	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrVerificationCodeSendRateLimited = errors.New("verification code send rate limited")

	ErrVerificationCodeCheckRateLimited = errors.New("verification code check rate limited")
)

type CodeCache interface {
	Set(ctx context.Context, bizType, phone, verificationCode string) error
	Verify(ctx context.Context, bizType, phone, verificationCode string) (bool, error)
}

type redisCodeCache struct {
	redisClient redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &redisCodeCache{
		redisClient: cmd,
	}
}

// -2 → verification code exists but missing TTL (data exception)
// -1 → send rate limited
// >=0 → set successfully
func (c *redisCodeCache) Set(ctx context.Context, bizType, phone, verificationCode string) error {
	key := c.buildKey(bizType, phone)
	result, err := c.redisClient.Eval(
		ctx,
		luaSetCode,
		[]string{key},
		verificationCode,
	).Int()

	if err != nil {
		// Redis execution failed (network exception, script execution failure, etc.)
		return err
	}

	switch result {
	case -2:
		return errors.New("verification code exists but missing valid expiration time (TTL)")

	case -1:
		return ErrVerificationCodeSendRateLimited

	default:
		return nil
	}
}

// -2 → verification code does not exist or does not match
// -1 → verification frequency limited
// >=0 → verification passed
func (c *redisCodeCache) Verify(ctx context.Context, bizType, phone, verificationCode string) (bool, error) {
	result, err := c.redisClient.Eval(
		ctx,
		luaVerifyCode,
		[]string{c.buildKey(bizType, phone)},
		verificationCode,
	).Int()

	if err != nil {
		return false, err
	}

	switch result {
	case -2:
		return false, nil
	case -1:
		return false, ErrVerificationCodeCheckRateLimited
	default:
		return true, nil
	}
}

func (c *redisCodeCache) buildKey(bizType, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", bizType, phone)
}
