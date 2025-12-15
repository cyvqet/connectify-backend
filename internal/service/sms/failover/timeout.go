package failover

import (
	"connectify/internal/service/sms"
	"context"
	"errors"
	"sync/atomic"
)

type TimeoutService struct {
	services  []sms.Service // the services to try
	index     uint32        // the index of the current service to try (0-based)
	count     uint32        // the number of times the service has been tried
	threshold uint32        // the number of times the service has been tried before giving up
}

func NewTimeoutService(services []sms.Service) sms.Service {
	return &TimeoutService{services: services}
}

func (s *TimeoutService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	length := uint32(len(s.services))
	if length == 0 {
		return errors.New("no SMS services configured")
	}

	index := atomic.LoadUint32(&s.index) % length

	if s.threshold > 0 {
		count := atomic.LoadUint32(&s.count)
		if count >= s.threshold {
			cur := atomic.LoadUint32(&s.index) % length
			newIndex := (cur + 1) % length
			if atomic.CompareAndSwapUint32(&s.index, cur, newIndex) {
				atomic.StoreUint32(&s.count, 0)
				index = newIndex
			} else {
				index = atomic.LoadUint32(&s.index) % length
			}
		}
	}

	svc := s.services[index]
	err := svc.Send(ctx, tplId, args, numbers...)

	switch {
	case err == nil:
		atomic.StoreUint32(&s.count, 0)
		return nil
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		atomic.AddUint32(&s.count, 1)
		return err
	default:
		return err
	}
}
