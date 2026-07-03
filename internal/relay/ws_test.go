package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
)

type mockStore struct {
	events []*Event
}

func (m *mockStore) SaveEvent(e *Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockStore) QueryEvents(filters []Filter) ([]*Event, error) {
	var result []*Event
	for _, e := range m.events {
		match := true
		for _, f := range filters {
			if !f.Match(e) {
				match = false
				break
			}
		}
		if match {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockStore) Close() error { return nil }

type mockRL struct{}

func (m *mockRL) Allow(string) bool { return true }

func TestWebSocketEventOK(t *testing.T) {
	store := &mockStore{}
	rl := &mockRL{}
	relay := NewRelay(store, rl)

	server := httptest.NewServer(http.HandlerFunc(relay.HandleWebSocket))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(t.Context(), url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	conn.Read(t.Context())

	e := makeTestEvent()
	e.ID = e.ComputeID()
	msg := []interface{}{"EVENT", e}
	data, _ := json.Marshal(msg)
	conn.Write(t.Context(), websocket.MessageText, data)

	_, resp, err := conn.Read(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var respArr []interface{}
	json.Unmarshal(resp, &respArr)
	if len(respArr) < 3 {
		t.Fatalf("unexpected response: %s", string(resp))
	}
	ok, _ := respArr[2].(bool)
	if !ok {
		t.Fatalf("expected OK true, got: %s", string(resp))
	}
}

func TestWebSocketREQEOSE(t *testing.T) {
	store := &mockStore{}
	rl := &mockRL{}
	relay := NewRelay(store, rl)

	e := makeTestEvent()
	e.ID = e.ComputeID()
	store.SaveEvent(e)

	server := httptest.NewServer(http.HandlerFunc(relay.HandleWebSocket))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(t.Context(), url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	conn.Read(t.Context())

	req := []interface{}{"REQ", "test", map[string]interface{}{"kinds": []int{1}}}
	data, _ := json.Marshal(req)
	conn.Write(t.Context(), websocket.MessageText, data)

	foundEvent := false
	foundEOSE := false
	for {
		_, resp, err := conn.Read(t.Context())
		if err != nil {
			break
		}
		var respArr []interface{}
		json.Unmarshal(resp, &respArr)
		if len(respArr) == 0 {
			continue
		}
		switch respArr[0].(string) {
		case "EVENT":
			foundEvent = true
		case "EOSE":
			foundEOSE = true
		}
		if foundEOSE {
			break
		}
	}
	if !foundEOSE {
		t.Fatal("expected EOSE")
	}
	if !foundEvent {
		t.Fatal("expected EVENT")
	}
}

func TestWebSocketInvalidEvent(t *testing.T) {
	store := &mockStore{}
	rl := &mockRL{}
	relay := NewRelay(store, rl)

	server := httptest.NewServer(http.HandlerFunc(relay.HandleWebSocket))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(t.Context(), url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	conn.Read(t.Context())

	badEvent := map[string]interface{}{
		"id":     "bad",
		"pubkey": "bad",
		"kind":   1,
	}
	msg := []interface{}{"EVENT", badEvent}
	data, _ := json.Marshal(msg)
	conn.Write(t.Context(), websocket.MessageText, data)

	_, resp, err := conn.Read(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var respArr []interface{}
	json.Unmarshal(resp, &respArr)
	if len(respArr) < 3 {
		t.Fatalf("unexpected response: %s", string(resp))
	}
	ok, _ := respArr[2].(bool)
	if ok {
		t.Fatalf("expected OK false for invalid event, got: %s", string(resp))
	}
}
