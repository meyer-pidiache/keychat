package templates

import (
	"strings"
	"testing"
)

func TestTruncateShort(t *testing.T) {
	s := truncate("hello", 10)
	if s != "hello" {
		t.Fatalf("expected 'hello', got '%s'", s)
	}
}

func TestTruncateLong(t *testing.T) {
	s := truncate("hello world this is long", 10)
	if !strings.HasSuffix(s, "...") {
		t.Fatal("expected truncation with ...")
	}
}

func TestFormatTime(t *testing.T) {
	s := formatTime(1700000000)
	if s != "2023-11-14 22:13:20" {
		t.Fatalf("unexpected time format: %s", s)
	}
}
