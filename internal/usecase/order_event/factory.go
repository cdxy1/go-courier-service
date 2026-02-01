package order_event

import (
	"context"
	"strings"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type Handler interface {
	Handle(ctx context.Context, event model.OrderStatusEvent) error
}

type HandlerFactory struct {
	handlers map[string]Handler
}

func NewHandlerFactory(uc deliveryUsecase) *HandlerFactory {
	f := &HandlerFactory{
		handlers: map[string]Handler{
			statusCreated:   &createdHandler{uc: uc},
			statusCancelled: &cancelledHandler{uc: uc},
			statusCanceled:  &cancelledHandler{uc: uc},
			statusCompleted: &completedHandler{uc: uc},
			statusDelivered: &completedHandler{uc: uc},
		},
	}
	return f
}

func (f *HandlerFactory) Handler(status string) (Handler, bool) {
	if f == nil {
		return nil, false
	}
	key := strings.ToLower(strings.TrimSpace(status))
	handler, ok := f.handlers[key]
	return handler, ok
}
