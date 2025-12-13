package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"connectify/internal/repository"
	"connectify/internal/repository/cache"
)

type CodeService struct {
	repo repository.CodeRepository
}

func NewCodeService(repo repository.CodeRepository) *CodeService {
	return &CodeService{
		repo: repo,
	}
}

func (svc *CodeService) Send(ctx context.Context, bizType, phone string) (string, error) {
	verificationCode, err := svc.generate()
	if err != nil {
		return "", fmt.Errorf("generate verification code failed: %w", err)
	}
	if err := svc.repo.Set(ctx, bizType, phone, verificationCode); err != nil {
		return "", fmt.Errorf("set verification code failed: %w", err)
	}

	// TODO: Call SMS service to send verification code here
	// const smsTemplateID = "your_sms_template_id"
	// return svc.sms.Send(ctx, smsTemplateID, []string{verificationCode}, phone)

	return verificationCode, nil
}

func (svc *CodeService) Verify(ctx context.Context, bizType, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, bizType, phone, inputCode)

	// If the verification frequency limit is triggered, only tell the upper layer "verification failed"
	if err == cache.ErrVerificationCodeCheckRateLimited {
		return false, nil
	}
	return ok, err
}

// generate generate 6-digit random verification code (using crypto/rand to ensure unpredictability)
func (svc *CodeService) generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
