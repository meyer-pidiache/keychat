package relay

import "testing"

func TestSubscriptionAddAndRemove(t *testing.T) {
	sm := NewSubscriptionManager()
	if !sm.Add("conn1", "sub1", []Filter{{Kinds: []int{1}}}) {
		t.Fatal("expected add to succeed")
	}
	if sm.Add("conn1", "sub1", []Filter{}) {
		t.Fatal("expected duplicate add to fail")
	}
	if !sm.Exists("sub1") {
		t.Fatal("expected sub1 to exist")
	}
	sm.Remove("sub1")
	if sm.Exists("sub1") {
		t.Fatal("expected sub1 to be removed")
	}
}

func TestSubscriptionRemoveAllForConn(t *testing.T) {
	sm := NewSubscriptionManager()
	sm.Add("conn1", "sub1", []Filter{{Kinds: []int{1}}})
	sm.Add("conn1", "sub2", []Filter{{Kinds: []int{4}}})
	sm.Add("conn2", "sub3", []Filter{{Kinds: []int{7}}})
	sm.RemoveAllForConn("conn1")
	if sm.Exists("sub1") {
		t.Fatal("expected sub1 removed")
	}
	if sm.Exists("sub2") {
		t.Fatal("expected sub2 removed")
	}
	if !sm.Exists("sub3") {
		t.Fatal("expected sub3 to remain")
	}
}

func TestSubscriptionCountLimit(t *testing.T) {
	sm := NewSubscriptionManager()
	for i := 0; i < 10; i++ {
		subID := string(rune('a' + i))
		sm.Add("conn1", subID, []Filter{{Kinds: []int{i}}})
	}
	if sm.ConnSubscriptionCount("conn1") != 10 {
		t.Fatalf("expected 10 subs, got %d", sm.ConnSubscriptionCount("conn1"))
	}
}

func TestInvalidSubID(t *testing.T) {
	sm := NewSubscriptionManager()
	if sm.Add("conn1", "invalid space!", []Filter{}) {
		t.Fatal("expected invalid subID to be rejected")
	}
}
