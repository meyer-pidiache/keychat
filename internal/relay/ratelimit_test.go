package relay

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.Allow("test-ip") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiterBlocksExcess(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-ip") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if rl.Allow("test-ip") {
		t.Fatal("4th request should be blocked")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	if !rl.Allow("ip-a") {
		t.Fatal("ip-a should be allowed")
	}
	if rl.Allow("ip-a") {
		t.Fatal("ip-a second should be blocked")
	}
	if !rl.Allow("ip-b") {
		t.Fatal("ip-b should be allowed (different key)")
	}
}

func TestRateLimiterWindowSlides(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Second)
	rl.Allow("ip")
	rl.Allow("ip")
	if rl.Allow("ip") {
		t.Fatal("should be blocked")
	}
	time.Sleep(1100 * time.Millisecond)
	if !rl.Allow("ip") {
		t.Fatal("should be allowed after window passes")
	}
}
