package ratelimit

import (
	"math"
	"sync"
	"time"
)

type Limiter interface {
	Allow(key string) bool
}

type TokenBucketLimiter struct {
	rate            float64
	capacity        float64
	ttl             time.Duration
	cleanupInterval time.Duration
	now             func() time.Time

	mu          sync.Mutex
	buckets     map[string]*bucket
	lastCleanup time.Time
}

type bucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

func NewTokenBucketLimiter(rate float64, burst int, ttl time.Duration) *TokenBucketLimiter {
	if rate <= 0 {
		rate = 1
	}
	if burst <= 0 {
		burst = 1
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return newTokenBucketLimiter(rate, burst, ttl, time.Now)
}

func newTokenBucketLimiter(rate float64, burst int, ttl time.Duration, now func() time.Time) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:            rate,
		capacity:        float64(burst),
		ttl:             ttl,
		cleanupInterval: ttl,
		now:             now,
		buckets:         make(map[string]*bucket),
	}
}

func (l *TokenBucketLimiter) Allow(key string) bool {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupLocked(now)

	entry := l.buckets[key]
	if entry == nil {
		l.buckets[key] = &bucket{
			tokens:   l.capacity - 1,
			last:     now,
			lastSeen: now,
		}
		return true
	}

	l.refillLocked(entry, now)
	entry.lastSeen = now

	if entry.tokens < 1 {
		return false
	}
	entry.tokens -= 1
	return true
}

func (l *TokenBucketLimiter) refillLocked(entry *bucket, now time.Time) {
	if !now.After(entry.last) {
		return
	}
	elapsed := now.Sub(entry.last).Seconds()
	entry.tokens = math.Min(l.capacity, entry.tokens+(elapsed*l.rate))
	entry.last = now
}

func (l *TokenBucketLimiter) cleanupLocked(now time.Time) {
	if l.lastCleanup.IsZero() {
		l.lastCleanup = now
		return
	}
	if now.Sub(l.lastCleanup) < l.cleanupInterval {
		return
	}
	cutoff := now.Add(-l.ttl)
	for key, entry := range l.buckets {
		if entry.lastSeen.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
	l.lastCleanup = now
}
