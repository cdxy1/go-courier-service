package worker

import (
	"context"
	"log"
	"time"

	"github.com/cdxy1/go-courier-service/internal/gateway/order"
	"github.com/cdxy1/go-courier-service/internal/model"
)

type OrderAssigner struct {
	orderGateway orderGateway
	deliveryUC   deliveryUsecase
	ticker       *time.Ticker
	cursor       time.Time
}

type orderGateway interface {
	GetOrders(ctx context.Context, from time.Time) ([]*model.Order, error)
}

type deliveryUsecase interface {
	Assign(ctx context.Context, order_id string) (*model.DeliveryModel, *model.CourierModel, error)
}

func NewOrderAssigner(orderGateway *order.OrderGateway, deliveryUC deliveryUsecase) *OrderAssigner {
	return &OrderAssigner{
		orderGateway: orderGateway,
		deliveryUC:   deliveryUC,
		ticker:       time.NewTicker(5 * time.Second),
		cursor:       time.Now().Add(-5 * time.Second),
	}
}

func (w *OrderAssigner) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.ticker.Stop()
			return
		case <-w.ticker.C:
			w.processTick(ctx)
		}
	}
}

func (w *OrderAssigner) processTick(ctx context.Context) {
	now := time.Now()
	currentCursor := now.Add(-5 * time.Second)

	cursor := w.cursor
	if currentCursor.After(w.cursor) {
		cursor = currentCursor
	}

	orders, err := w.orderGateway.GetOrders(ctx, cursor)
	if err != nil {
		log.Printf("Failed to fetch orders: %v", err)
		return
	}

	maxCreatedAt := cursor
	for _, ord := range orders {
		if ord.CreatedAt.After(maxCreatedAt) {
			maxCreatedAt = ord.CreatedAt
		}

		delivery, courier, err := w.deliveryUC.Assign(ctx, ord.ID)
		if err != nil {
			log.Printf("Failed to assign courier to order %s: %v", ord.ID, err)
			continue
		}

		log.Printf("Assigned courier %d (transport: %s) to order %s, deadline: %s",
			courier.ID, courier.TransportType, ord.ID, delivery.Deadline.Format(time.RFC3339))
	}

	if len(orders) > 0 {
		w.cursor = maxCreatedAt
		log.Printf("Updated cursor to %s", w.cursor.Format(time.RFC3339))
	} else {
		w.cursor = currentCursor
	}
}
