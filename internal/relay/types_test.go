package relay

import (
	"encoding/json"
	"testing"
)

func makeTestEvent() *Event {
	return &Event{
		ID:        "0000000000000000000000000000000000000000000000000000000000000000",
		Pubkey:    "1111111111111111111111111111111111111111111111111111111111111111",
		CreatedAt: 1700000000,
		Kind:      1,
		Tags:      []Tag{},
		Content:   "hello world",
		Sig:       "22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222",
	}
}

func TestEventSerialize(t *testing.T) {
	e := makeTestEvent()
	data := e.Serialize()
	if len(data) == 0 {
		t.Fatal("empty serialization")
	}
}

func TestEventValidateValid(t *testing.T) {
	e := makeTestEvent()
	e.ID = e.ComputeID()
	err := e.Validate()
	if err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestEventValidateBadID(t *testing.T) {
	e := makeTestEvent()
	e.ID = "short"
	err := e.Validate()
	if err == nil {
		t.Fatal("expected error for short id")
	}
}

func TestEventValidateBadPubkey(t *testing.T) {
	e := makeTestEvent()
	e.Pubkey = "zz"
	err := e.Validate()
	if err == nil {
		t.Fatal("expected error for bad pubkey")
	}
}

func TestEventValidateFutureTimestamp(t *testing.T) {
	e := makeTestEvent()
	e.CreatedAt = 999999999999
	e.ID = e.ComputeID()
	err := e.Validate()
	if err == nil {
		t.Fatal("expected error for future timestamp")
	}
}

func TestEventValidatePreNostr(t *testing.T) {
	e := makeTestEvent()
	e.CreatedAt = 1000000
	e.ID = e.ComputeID()
	err := e.Validate()
	if err == nil {
		t.Fatal("expected error for pre-Nostr timestamp")
	}
}

func TestEventJSONRoundtrip(t *testing.T) {
	e := makeTestEvent()
	e.ID = e.ComputeID()
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var e2 Event
	if err := json.Unmarshal(data, &e2); err != nil {
		t.Fatal(err)
	}
	if e2.ID != e.ID {
		t.Fatalf("id mismatch: %s != %s", e2.ID, e.ID)
	}
	if e2.Content != e.Content {
		t.Fatalf("content mismatch: %s != %s", e2.Content, e.Content)
	}
}
