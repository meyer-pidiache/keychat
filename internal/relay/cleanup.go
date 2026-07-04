package relay

import (
	"log"
	"time"
)

type TTLStore interface {
	DeleteOlderThan(t time.Time) (int64, error)
	Close() error
}

type TTLCleanup struct {
	store TTLStore
	ttl   time.Duration
	stop  chan struct{}
}

func NewTTLCleanup(store TTLStore, ttl time.Duration) *TTLCleanup {
	return &TTLCleanup{
		store: store,
		ttl:   ttl,
		stop:  make(chan struct{}),
	}
}

func (tc *TTLCleanup) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tc.run()
			case <-tc.stop:
				return
			}
		}
	}()
}

func (tc *TTLCleanup) Stop() {
	close(tc.stop)
}

func (tc *TTLCleanup) RunOnce() {
	tc.run()
}

func (tc *TTLCleanup) run() {
	cutoff := time.Now().Add(-tc.ttl)
	deleted, err := tc.store.DeleteOlderThan(cutoff)
	if err != nil {
		log.Printf("ttl cleanup error: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("ttl cleanup: deleted %d events older than %v", deleted, cutoff)
	}
}
