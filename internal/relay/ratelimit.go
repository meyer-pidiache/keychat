package relay

import (
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
)

type RateLimiter struct {
	counters *xsync.MapOf[string, *slidingWindow]
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

type slidingWindow struct {
	entries []int64
	mu      sync.Mutex
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		counters: xsync.NewMapOf[string, *slidingWindow](),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup(10 * time.Minute)
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now().Unix()
	sw, _ := rl.counters.LoadOrStore(key, &slidingWindow{})

	sw.mu.Lock()
	defer sw.mu.Unlock()

	cutoff := now - int64(rl.window.Seconds())
	var valid []int64
	for _, t := range sw.entries {
		if t > cutoff {
			valid = append(valid, t)
		}
	}
	sw.entries = valid

	if len(sw.entries) >= rl.limit {
		return false
	}

	sw.entries = append(sw.entries, now)
	return true
}

func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now().Unix()
		cutoff := now - int64(rl.window.Seconds())
		rl.counters.Range(func(key string, sw *slidingWindow) bool {
			sw.mu.Lock()
			var valid []int64
			for _, t := range sw.entries {
				if t > cutoff {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				rl.counters.Delete(key)
			} else {
				sw.entries = valid
			}
			sw.mu.Unlock()
			return true
		})
	}
}
