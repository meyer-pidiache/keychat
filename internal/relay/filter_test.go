package relay

import "testing"

func TestFilterMatchByKind(t *testing.T) {
	e := makeTestEvent()
	e.Kind = 1
	f := &Filter{Kinds: []int{1}}
	if !f.Match(e) {
		t.Fatal("expected kind 1 to match kind filter [1]")
	}
	f2 := &Filter{Kinds: []int{4}}
	if f2.Match(e) {
		t.Fatal("expected kind 1 to NOT match kind filter [4]")
	}
}

func TestFilterMatchByAuthor(t *testing.T) {
	e := makeTestEvent()
	f := &Filter{Authors: []string{e.Pubkey}}
	if !f.Match(e) {
		t.Fatal("expected author match")
	}
	f2 := &Filter{Authors: []string{"nonexistent"}}
	if f2.Match(e) {
		t.Fatal("expected no match for wrong author")
	}
}

func TestFilterMatchByID(t *testing.T) {
	e := makeTestEvent()
	f := &Filter{IDs: []string{e.ID}}
	if !f.Match(e) {
		t.Fatal("expected id match")
	}
	f2 := &Filter{IDs: []string{"other"}}
	if f2.Match(e) {
		t.Fatal("expected no match for wrong id")
	}
}

func TestFilterMatchByPTag(t *testing.T) {
	e := makeTestEvent()
	e.Tags = []Tag{{"p", "recipient123", ""}}
	f := &Filter{Tags: map[string][]string{"p": {"recipient123"}}}
	if !f.Match(e) {
		t.Fatal("expected p-tag match")
	}
	f2 := &Filter{Tags: map[string][]string{"p": {"other"}}}
	if f2.Match(e) {
		t.Fatal("expected no match for wrong p-tag")
	}
}

func TestFilterMatchBySince(t *testing.T) {
	e := makeTestEvent()
	e.CreatedAt = 1000
	since := int64(500)
	f := &Filter{Since: &since}
	if !f.Match(e) {
		t.Fatal("expected match: event after since")
	}
	since2 := int64(2000)
	f2 := &Filter{Since: &since2}
	if f2.Match(e) {
		t.Fatal("expected no match: event before since")
	}
}

func TestFilterMatchByUntil(t *testing.T) {
	e := makeTestEvent()
	e.CreatedAt = 1000
	until := int64(2000)
	f := &Filter{Until: &until}
	if !f.Match(e) {
		t.Fatal("expected match: event before until")
	}
	until2 := int64(500)
	f2 := &Filter{Until: &until2}
	if f2.Match(e) {
		t.Fatal("expected no match: event after until")
	}
}

func TestFilterJSONRoundtrip(t *testing.T) {
	f := &Filter{
		Kinds:   []int{1, 4},
		Authors: []string{"abc"},
		Tags:    map[string][]string{"p": {"xyz"}},
	}
	data, err := f.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	f2, err := FilterFromJSON(data)
	if err != nil {
		t.Fatalf("parse error: %v, json: %s", err, string(data))
	}
	if len(f2.Kinds) != 2 || f2.Kinds[0] != 1 {
		t.Fatal("kinds mismatch after roundtrip")
	}
	got := f2.Tags["p"]
	if len(got) != 1 || got[0] != "xyz" {
		t.Fatalf("p-tag mismatch after roundtrip: got %v, expected [xyz]", got)
	}
}

func TestFilterJSONFromNostrClient(t *testing.T) {
	raw := `{"kinds":[1059],"#p":["abcdef1234567890abcdef1234567890"],"limit":20}`
	f, err := FilterFromJSON([]byte(raw))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(f.Kinds) != 1 || f.Kinds[0] != 1059 {
		t.Fatalf("kinds mismatch: got %v", f.Kinds)
	}
	if f.Tags["p"][0] != "abcdef1234567890abcdef1234567890" {
		t.Fatalf("p-tag mismatch: got %v", f.Tags["p"])
	}
	if f.Limit == nil || *f.Limit != 20 {
		t.Fatal("limit mismatch")
	}
}
