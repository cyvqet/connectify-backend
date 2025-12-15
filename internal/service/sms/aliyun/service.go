package aliyun

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	appId    string
	signName string
}

func NewService(appId, signName string) *Service {
	return &Service{
		appId:    appId,
		signName: signName,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	zap.L().Info("aliyun sms send",
		zap.String("appId", s.appId),
		zap.String("signName", s.signName),
		zap.String("tplId", tplId),
		zap.Strings("args", args),
		zap.Strings("numbers", numbers),
	)
	return nil
}
