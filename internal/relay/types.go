package relay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

type Event struct {
	ID        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      []Tag      `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

type Tag []string

type Filter struct {
	IDs     []string          `json:"ids,omitempty"`
	Authors []string          `json:"authors,omitempty"`
	Kinds   []int             `json:"kinds,omitempty"`
	Tags    map[string][]string `json:"-"`
	Since   *int64            `json:"since,omitempty"`
	Until   *int64            `json:"until,omitempty"`
	Limit   *int              `json:"limit,omitempty"`
}

type Subscription struct {
	ID      string
	Filters []Filter
}

var subIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidateSubscriptionID(id string) bool {
	return len(id) > 0 && len(id) <= 256 && subIDRegex.MatchString(id)
}

func (e *Event) Serialize() []byte {
	tags, _ := json.Marshal(e.Tags)
	data := fmt.Sprintf(`[0,"%s",%d,%d,%s,"%s"]`,
		e.Pubkey, e.CreatedAt, e.Kind, string(tags), escapeJSONString(e.Content))
	return []byte(data)
}

func (e *Event) Validate() error {
	if len(e.ID) != 64 {
		return fmt.Errorf("invalid id length: expected 64 hex chars")
	}
	if _, err := hex.DecodeString(e.ID); err != nil {
		return fmt.Errorf("invalid id hex: %w", err)
	}
	if len(e.Pubkey) != 64 {
		return fmt.Errorf("invalid pubkey length: expected 64 hex chars")
	}
	if _, err := hex.DecodeString(e.Pubkey); err != nil {
		return fmt.Errorf("invalid pubkey hex: %w", err)
	}
	if len(e.Sig) != 128 {
		return fmt.Errorf("invalid sig length: expected 128 hex chars")
	}
	if _, err := hex.DecodeString(e.Sig); err != nil {
		return fmt.Errorf("invalid sig hex: %w", err)
	}
	now := time.Now().Unix()
	if e.CreatedAt > now+3600 {
		return fmt.Errorf("created_at too far in the future")
	}
	if e.CreatedAt < 1420070400 {
		return fmt.Errorf("created_at before Nostr epoch (2015)")
	}
	if len(e.Content) > 65536 {
		return fmt.Errorf("content too large: max 64KB")
	}
	computedID := e.ComputeID()
	if computedID != e.ID {
		return fmt.Errorf("id does not match computed hash")
	}
	return nil
}

func (e *Event) ComputeID() string {
	hash := sha256.Sum256(e.Serialize())
	return hex.EncodeToString(hash[:])
}

func escapeJSONString(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func EventFromJSON(data []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
