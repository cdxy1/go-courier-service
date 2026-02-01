package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketLimiterAllow(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := newTokenBucketLimiter(1, 2, time.Minute, func() time.Time { return now })

	if !limiter.Allow("client") {
		t.Fatalf("expected first request to be allowed")
	}
	if !limiter.Allow("client") {
		t.Fatalf("expected second request to be allowed")
	}
	if limiter.Allow("client") {
		t.Fatalf("expected third request to be rate limited")
	}

	now = now.Add(time.Second)
	if !limiter.Allow("client") {
		t.Fatalf("expected token to refill after a second")
	}
}
