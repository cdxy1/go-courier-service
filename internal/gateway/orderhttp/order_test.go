package orderhttp

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrderGatewayRetriesOn429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"delivered"}`)
	}))
	defer server.Close()

	gateway := NewOrderGateway(server.URL)
	gateway.maxAttempts = 3
	gateway.baseDelay = 0
	gateway.wait = func(ctx context.Context, delay time.Duration) error { return nil }
	gateway.logger = log.New(io.Discard, "", 0)

	status, err := gateway.GetOrderStatus(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "delivered" {
		t.Fatalf("expected status delivered, got %s", status)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
