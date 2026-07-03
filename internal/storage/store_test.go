package storage

import (
	"os"
	"testing"
	"time"

	"github.com/opc/keychat/internal/relay"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	f, err := os.CreateTemp("", "keychat-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	s, err := NewSQLiteStore(f.Name())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}
	t.Cleanup(func() {
		s.Close()
		os.Remove(f.Name())
	})
	return s
}

func makeEvent(id string, kind int, pubkey string, ts int64) *relay.Event {
	return &relay.Event{
		ID:        id,
		Pubkey:    pubkey,
		CreatedAt: ts,
		Kind:      kind,
		Tags:      []relay.Tag{},
		Content:   "test",
		Sig:       "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	}
}

func TestSaveAndQueryByID(t *testing.T) {
	s := newTestStore(t)
	e := makeEvent("abc123", 1, "pubkey1", 1700000000)
	if err := s.SaveEvent(e); err != nil {
		t.Fatal(err)
	}
	events, err := s.QueryEvents([]relay.Filter{{IDs: []string{"abc123"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "abc123" {
		t.Fatalf("wrong event: %s", events[0].ID)
	}
}

func TestFilterByKind(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		s.SaveEvent(makeEvent(string(rune('a'+i)), 1, "pubkey1", 1700000000))
	}
	for i := 0; i < 3; i++ {
		s.SaveEvent(makeEvent(string(rune('f'+i)), 4, "pubkey1", 1700000000))
	}
	events, err := s.QueryEvents([]relay.Filter{{Kinds: []int{1}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Fatalf("expected 5 kind-1 events, got %d", len(events))
	}
	events, err = s.QueryEvents([]relay.Filter{{Kinds: []int{4}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 kind-4 events, got %d", len(events))
	}
}

func TestFilterByPTag(t *testing.T) {
	s := newTestStore(t)
	alice := makeEvent("e1", 1, "alice", 1700000000)
	alice.Tags = []relay.Tag{{"p", "bob", ""}}
	s.SaveEvent(alice)
	bob := makeEvent("e2", 1, "bob", 1700000000)
	bob.Tags = []relay.Tag{{"p", "alice", ""}}
	s.SaveEvent(bob)

	events, err := s.QueryEvents([]relay.Filter{{Tags: map[string][]string{"p": {"bob"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != "e1" {
		t.Fatalf("expected 1 event with p=bob, got %d", len(events))
	}
}

func TestTTLDelete(t *testing.T) {
	s := newTestStore(t)
	old := makeEvent("old", 1, "pubkey1", 1000000)
	s.SaveEvent(old)
	new := makeEvent("new", 1, "pubkey1", 2000000000)
	s.SaveEvent(new)

	deleted, err := s.DeleteOlderThan(time.Unix(1000000000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}
	events, _ := s.QueryEvents([]relay.Filter{{IDs: []string{"old"}}})
	if len(events) != 0 {
		t.Fatal("old event should have been deleted")
	}
	events, _ = s.QueryEvents([]relay.Filter{{IDs: []string{"new"}}})
	if len(events) != 1 {
		t.Fatal("new event should remain")
	}
}
