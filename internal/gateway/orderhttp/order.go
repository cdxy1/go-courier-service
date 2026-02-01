package orderhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cdxy1/go-courier-service/internal/observability"
)

type OrderGateway struct {
	baseURL     string
	client      *http.Client
	maxAttempts int
	baseDelay   time.Duration
	wait        func(context.Context, time.Duration) error
	logger      *log.Logger
}

func NewOrderGateway(baseURL string) *OrderGateway {
	return &OrderGateway{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		maxAttempts: 3,
		baseDelay:   150 * time.Millisecond,
		wait:        waitWithContext,
		logger:      log.New(os.Stdout, "[WARN] ", log.LstdFlags),
	}
}

func (g *OrderGateway) GetOrderStatus(ctx context.Context, orderID string) (string, error) {
	if orderID == "" {
		return "", fmt.Errorf("order id is empty")
	}
	url := fmt.Sprintf("%s/public/api/v1/order/%s", g.baseURL, orderID)

	var lastErr error
	for attempt := 1; attempt <= g.maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		status, code, err := g.fetchStatus(ctx, url)
		if err == nil {
			return status, nil
		}
		lastErr = err

		if !shouldRetry(err, code) || attempt == g.maxAttempts {
			return "", fmt.Errorf("request order status: %w", err)
		}

		observability.IncGatewayRetries()
		delay := g.baseDelay * time.Duration(1<<uint(attempt-1))
		g.logger.Printf("order status retry attempt=%d delay=%s err=%v", attempt, delay, err)
		if waitErr := g.wait(ctx, delay); waitErr != nil {
			return "", waitErr
		}
	}

	return "", fmt.Errorf("request order status: %w", lastErr)
}

type orderStatusResponse struct {
	ID        string `json:"id"`
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (g *OrderGateway) fetchStatus(ctx context.Context, url string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			g.logger.Printf("error while close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("order service status %d", resp.StatusCode)
	}

	var payload orderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", resp.StatusCode, fmt.Errorf("decode order response: %w", err)
	}
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		return "", resp.StatusCode, fmt.Errorf("order status is empty")
	}
	return status, resp.StatusCode, nil
}

func shouldRetry(err error, code int) bool {
	if code != 0 {
		return shouldRetryStatus(code)
	}
	return shouldRetryError(err)
}

func shouldRetryStatus(code int) bool {
	if code == http.StatusTooManyRequests || code == http.StatusRequestTimeout {
		return true
	}
	return code >= http.StatusInternalServerError
}

func shouldRetryError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func waitWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
