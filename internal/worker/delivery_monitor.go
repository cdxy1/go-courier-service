package worker

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	deliveryRepo "github.com/cdxy1/go-courier-service/internal/repository/delivery"
)

type DeliveryMonitorUsecase interface {
	ProcessExpiredDeliveries(ctx context.Context) (int, error)
}

type DeliveryMonitor struct {
	uc       DeliveryMonitorUsecase
	interval time.Duration
	logger   *log.Logger
}

func NewDeliveryMonitor(uc DeliveryMonitorUsecase, interval time.Duration, logger *log.Logger) *DeliveryMonitor {
	if logger == nil {
		logger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	}
	return &DeliveryMonitor{uc: uc, interval: interval, logger: logger}
}

func (m *DeliveryMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.logger.Printf("starting delivery deadline monitor, interval=%s", m.interval)
	for {
		select {
		case <-ctx.Done():
			m.logger.Println("stopping delivery deadline monitor")
			return
		case <-ticker.C:
			updated, err := m.uc.ProcessExpiredDeliveries(ctx)
			if err != nil {
				if errors.Is(err, deliveryRepo.ErrDeliveryTableMissing) {
					m.logger.Println("warning: delivery table does not exist, please run migrations: make migrate")
					continue
				}
				m.logger.Printf("error processing expired deliveries: %v", err)
				continue
			}
			if updated > 0 {
				m.logger.Printf("delivery monitor: %d couriers marked available", updated)
			}
		}
	}
}
