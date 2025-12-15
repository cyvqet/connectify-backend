package ratelimit

import (
	"connectify/internal/service/sms"
	"connectify/pkg/ratelimit"
	"context"
	"errors"
	"fmt"
)

type Service struct {
	smsSvc  sms.Service
	limiter ratelimit.Limiter
}

func NewService(smsSvc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{
		smsSvc:  smsSvc,
		limiter: limiter,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	key := fmt.Sprintf("sms:send:%s", numbers[0])
	ok, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("rate limit exceeded")
	}

	err = s.smsSvc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		return err
	}

	return nil
}
