package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/cyvqet/connectify/internal/service/sms"
)

type Service struct {
	services []sms.Service
	index    uint64
}

func NewService(services []sms.Service) sms.Service {
	return &Service{services: services}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	length := uint64(len(s.services))
	if length == 0 {
		return errors.New("no SMS services configured")
	}

	index := atomic.AddUint64(&s.index, 1)

	for i := index; i < index+length; i++ {
		svc := s.services[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded),
			errors.Is(err, context.Canceled):
			// ctx is invalid, continue retrying is meaningless
			return err
		default:
			// TODO: log error
			continue
		}
	}
	return errors.New("all SMS services failed")
}
