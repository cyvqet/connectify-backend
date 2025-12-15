package ratelimit

import (
	"context"
	"errors"

	"github.com/cyvqet/connectify/internal/service/sms"
	"github.com/cyvqet/connectify/pkg/ratelimit"
)

type Service struct {
	smsSvc  sms.Service
	limiter ratelimit.Limiter
	key     string
}

func NewService(smsSvc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{
		smsSvc:  smsSvc,
		limiter: limiter,
		key:     "sms-rate-limit",
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ok, err := s.limiter.Limit(ctx, s.key)
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
