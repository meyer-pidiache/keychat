package relay

import (
	"testing"
	"time"
)

type mockTTLStore struct {
	deleted int64
}

func (m *mockTTLStore) DeleteOlderThan(t time.Time) (int64, error) {
	m.deleted++
	return 1, nil
}

func (m *mockTTLStore) Close() error { return nil }

func TestTTLCleanupRunOnce(t *testing.T) {
	store := &mockTTLStore{}
	tc := NewTTLCleanup(store, 24*time.Hour)
	tc.RunOnce()
	if store.deleted != 1 {
		t.Fatalf("expected 1 delete call, got %d", store.deleted)
	}
}
