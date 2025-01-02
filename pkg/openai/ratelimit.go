package openai

import (
	"context"
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the given parameters
func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			r.mu.Lock()
			now := time.Now()
			timePassed := now.Sub(r.lastRefillTime).Seconds()
			r.tokens = math.Min(r.maxTokens, r.tokens+timePassed*r.refillRate)
			r.lastRefillTime = now

			if r.tokens >= 1 {
				r.tokens--
				r.mu.Unlock()
				return nil
			}
			r.mu.Unlock()

			// Sleep for a short duration before trying again
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// RetryConfig defines the configuration for retry behavior
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}
}

// WithRetry wraps a function with retry logic using exponential backoff
func WithRetry[T any](ctx context.Context, cfg RetryConfig, fn func(context.Context) (T, error)) (T, error) {
	var lastErr error
	var result T

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, lastErr = fn(ctx)
		if lastErr == nil {
			return result, nil
		}

		if attempt == cfg.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt))
		if delay > float64(cfg.MaxDelay) {
			delay = float64(cfg.MaxDelay)
		}
		// Add jitter (±20%)
		jitter := (rand.Float64()*0.4 - 0.2) * delay
		delay += jitter

		timer := time.NewTimer(time.Duration(delay))
		select {
		case <-ctx.Done():
			timer.Stop()
			var zero T
			return zero, ctx.Err()
		case <-timer.C:
			continue
		}
	}

	var zero T
	return zero, lastErr
}
