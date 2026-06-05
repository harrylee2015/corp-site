package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	window   time.Duration
	max      int
}

func NewRateLimiter(window time.Duration, max int) *RateLimiter {
	r := &RateLimiter{
		attempts: make(map[string][]time.Time),
		window:   window,
		max:      max,
	}
	go r.cleanup()
	return r
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	filtered := make([]time.Time, 0)
	for _, t := range r.attempts[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.max {
		r.attempts[key] = filtered
		return false
	}

	r.attempts[key] = append(filtered, now)
	return true
}

func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		r.mu.Lock()
		cutoff := time.Now().Add(-r.window)
		for k, times := range r.attempts {
			filtered := make([]time.Time, 0)
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(r.attempts, k)
			} else {
				r.attempts[k] = filtered
			}
		}
		r.mu.Unlock()
	}
}

func LoginRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(429, gin.H{"error": "登录尝试过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		c.Next()
	}
}
