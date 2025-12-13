package repository

import (
	"context"

	"connectify/internal/repository/cache"
)

var (
	ErrVerificationCodeSendRateLimited  = cache.ErrVerificationCodeSendRateLimited
	ErrVerificationCodeCheckRateLimited = cache.ErrVerificationCodeCheckRateLimited
)

type CodeRepository interface {
	Set(ctx context.Context, bizType, phone, verificationCode string) error
	Verify(ctx context.Context, bizType, phone, verificationCode string) (bool, error)
}

type codeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(codeCache cache.CodeCache) CodeRepository {
	return &codeRepository{
		cache: codeCache,
	}
}

func (r *codeRepository) Set(ctx context.Context, bizType, phone, verificationCode string) error {
	return r.cache.Set(ctx, bizType, phone, verificationCode)
}

func (r *codeRepository) Verify(ctx context.Context, bizType, phone, verificationCode string) (bool, error) {
	return r.cache.Verify(ctx, bizType, phone, verificationCode)
}
