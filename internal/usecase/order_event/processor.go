package order_event

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type orderStatusGateway interface {
	GetOrderStatus(ctx context.Context, orderID string) (string, error)
}

type Processor struct {
	factory *HandlerFactory
	gateway orderStatusGateway
}

func NewProcessor(factory *HandlerFactory, gateway orderStatusGateway) *Processor {
	return &Processor{factory: factory, gateway: gateway}
}

func (p *Processor) Handle(ctx context.Context, event model.OrderStatusEvent) error {
	if strings.TrimSpace(event.OrderID) == "" || strings.TrimSpace(event.Status) == "" {
		return fmt.Errorf("invalid order event payload")
	}

	status, err := p.gateway.GetOrderStatus(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("fetch order status: %w", err)
	}

	if !sameStatus(status, event.Status) {
		log.Printf("order event skipped: status changed for order %s (event=%s actual=%s)", event.OrderID, event.Status, status)
		return nil
	}

	handler, ok := p.factory.Handler(event.Status)
	if !ok {
		return nil
	}

	return handler.Handle(ctx, event)
}

func sameStatus(actual string, event string) bool {
	return strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(event))
}
