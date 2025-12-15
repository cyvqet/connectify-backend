package service

//go:generate mockgen -source=code.go -destination=mocks/code_mock.go -package=svcmocks

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"connectify/internal/repository"
	"connectify/internal/repository/cache"
	"connectify/internal/service/sms"
)

type CodeService interface {
	Send(ctx context.Context, bizType, phone string) (string, error)
	Verify(ctx context.Context, bizType, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

const smsTemplateID = "SMS_VERIFICATION_CODE"

func (svc *codeService) Send(ctx context.Context, bizType, phone string) (string, error) {
	verificationCode, err := svc.generate()
	if err != nil {
		return "", fmt.Errorf("generate verification code failed: %w", err)
	}
	if err := svc.repo.Set(ctx, bizType, phone, verificationCode); err != nil {
		return "", fmt.Errorf("set verification code failed: %w", err)
	}

	if err := svc.smsSvc.Send(ctx, smsTemplateID, []string{verificationCode}, phone); err != nil {
		return "", fmt.Errorf("send sms failed: %w", err)
	}

	return verificationCode, nil
}

func (svc *codeService) Verify(ctx context.Context, bizType, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, bizType, phone, inputCode)

	// If the verification frequency limit is triggered, only tell the upper layer "verification failed"
	if err == cache.ErrVerificationCodeCheckRateLimited {
		return false, nil
	}
	return ok, err
}

// generate generate 6-digit random verification code (using crypto/rand to ensure unpredictability)
func (svc *codeService) generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
