package relay

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type EventStore interface {
	SaveEvent(event *Event) error
	QueryEvents(filters []Filter) ([]*Event, error)
	Close() error
}

type RateLimiter interface {
	Allow(key string) bool
}

type Relay struct {
	store      EventStore
	subManager *SubscriptionManager
	rateLim    RateLimiter
}

func NewRelay(store EventStore, rateLim RateLimiter) *Relay {
	return &Relay{
		store:      store,
		subManager: NewSubscriptionManager(),
		rateLim:    rateLim,
	}
}

func (r *Relay) HandleWebSocket(w http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("ws accept error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	connID := fmt.Sprintf("%p", conn)

	writeMu := sync.Mutex{}
	write := func(v []byte) {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.Write(req.Context(), websocket.MessageText, v)
	}

	write(NewNotice("connected to KeyChat relay"))

	msgChan := make(chan []byte, 64)
	go func() {
		for {
			_, msg, err := conn.Read(req.Context())
			if err != nil {
				close(msgChan)
				return
			}
			msgChan <- msg
		}
	}()

	for msg := range msgChan {
		r.handleMessage(connID, msg, write, req)
	}

	r.subManager.RemoveAllForConn(connID)
}

func (r *Relay) handleMessage(connID string, msg []byte, write func([]byte), req *http.Request) {
	var raw []interface{}
	if err := json.Unmarshal(msg, &raw); err != nil {
		write(NewNotice("invalid message format"))
		return
	}
	if len(raw) < 2 {
		write(NewNotice("message too short"))
		return
	}
	msgType, ok := raw[0].(string)
	if !ok {
		write(NewNotice("first element must be a string type"))
		return
	}

	switch msgType {
	case "EVENT":
		r.handleEvent(raw, write)
	case "REQ":
		r.handleREQ(connID, raw, write)
	case "CLOSE":
		r.handleClose(raw, write)
	default:
		write(NewNotice(fmt.Sprintf("unknown message type: %s", msgType)))
	}
}

func (r *Relay) handleEvent(raw []interface{}, write func([]byte)) {
	if len(raw) < 2 {
		write(NewNotice("EVENT missing data"))
		return
	}
	eventData, err := json.Marshal(raw[1])
	if err != nil {
		write(NewNotice("invalid event JSON"))
		return
	}
	event, err := EventFromJSON(eventData)
	if err != nil {
		write(NewOK("", "", false, "error: invalid event format"))
		return
	}
	if err := event.Validate(); err != nil {
		write(NewOK("", event.ID, false, fmt.Sprintf("error: %v", err)))
		return
	}
	if r.rateLim != nil && !r.rateLim.Allow(event.Pubkey) {
		write(NewOK("", event.ID, false, "rate-limited: too many events"))
		return
	}
	if err := r.store.SaveEvent(event); err != nil {
		write(NewOK("", event.ID, false, fmt.Sprintf("error: %v", err)))
		return
	}
	write(NewOK("", event.ID, true, ""))
}

func (r *Relay) handleREQ(connID string, raw []interface{}, write func([]byte)) {
	if len(raw) < 3 {
		write(NewNotice("REQ missing subscription ID or filters"))
		return
	}
	subID, ok := raw[1].(string)
	if !ok {
		write(NewNotice("REQ subscription ID must be a string"))
		return
	}
	if r.subManager.ConnSubscriptionCount(connID) >= 32 {
		write(NewClosed(subID, "too many subscriptions"))
		return
	}
	filters := make([]Filter, 0)
	for _, fraw := range raw[2:] {
		fdata, err := json.Marshal(fraw)
		if err != nil {
			continue
		}
		f, err := FilterFromJSON(fdata)
		if err != nil {
			continue
		}
		filters = append(filters, *f)
	}
	if len(filters) == 0 {
		write(NewNotice("REQ has no valid filters"))
		return
	}
	if !r.subManager.Add(connID, subID, filters) {
		write(NewNotice(fmt.Sprintf("subscription %s already exists", subID)))
		return
	}
	events, err := r.store.QueryEvents(filters)
	if err != nil {
		write(NewNotice(fmt.Sprintf("query error: %v", err)))
		r.subManager.Remove(subID)
		return
	}
	for _, event := range events {
		write(NewEventMessage(subID, event))
	}
	write(NewEOSE(subID))
}

func (r *Relay) handleClose(raw []interface{}, write func([]byte)) {
	if len(raw) < 2 {
		write(NewNotice("CLOSE missing subscription ID"))
		return
	}
	subID, ok := raw[1].(string)
	if !ok {
		write(NewNotice("CLOSE subscription ID must be a string"))
		return
	}
	r.subManager.Remove(subID)
}
