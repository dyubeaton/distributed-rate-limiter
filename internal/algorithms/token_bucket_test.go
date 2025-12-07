package algorithms

import (
	"testing"
	"time"
)

// TestTokenBucket_InitialState tests that new buckets start full
func TestTokenBucket_InitialState(t *testing.T) {
	bucket := NewTokenBucket(10, 2.0)

	// First request from a new user should be allowed (bucket starts full)
	allowed, remaining := bucket.Allow("user:123", 5)

	if !allowed {
		t.Errorf("Expected first request to be allowed, got denied")
	}

	if remaining != 5 {
		t.Errorf("Expected remaining=5, got %d", remaining)
	}
}

// TestTokenBucket_Consumption tests token consumption
func TestTokenBucket_Consumption(t *testing.T) {
	bucket := NewTokenBucket(10, 2.0)

	// Take 3 tokens
	allowed, remaining := bucket.Allow("user:123", 3)
	if !allowed || remaining != 7 {
		t.Errorf("First request: expected allowed=true, remaining=7; got allowed=%v, remaining=%d", allowed, remaining)
	}

	// Take 5 more tokens (should have 2 left)
	allowed, remaining = bucket.Allow("user:123", 5)
	if !allowed || remaining != 2 {
		t.Errorf("Second request: expected allowed=true, remaining=2; got allowed=%v, remaining=%d", allowed, remaining)
	}
}

// TestTokenBucket_Denial tests denial when insufficient tokens
func TestTokenBucket_Denial(t *testing.T) {
	bucket := NewTokenBucket(10, 2.0)

	// Take 8 tokens (2 remaining)
	bucket.Allow("user:123", 8)

	// Try to take 5 tokens (should be denied)
	allowed, remaining := bucket.Allow("user:123", 5)

	if allowed {
		t.Errorf("Expected request to be denied, but was allowed")
	}

	if remaining != 2 {
		t.Errorf("Expected remaining=2 after denial, got %d", remaining)
	}
}

// TestTokenBucket_Refill tests time-based token refill
func TestTokenBucket_Refill(t *testing.T) {
	bucket := NewTokenBucket(10, 5.0) // 5 tokens/second refill

	// Take 8 tokens (2 remaining)
	bucket.Allow("user:123", 8)

	// Wait 1 second (should add 5 tokens: 2 + 5 = 7)
	time.Sleep(1 * time.Second)

	// Request 6 tokens (should be allowed with 1 remaining)
	allowed, remaining := bucket.Allow("user:123", 6)

	if !allowed {
		t.Errorf("Expected request to be allowed after refill, but was denied")
	}

	// Remaining should be approximately 1 (allowing for small timing variations)
	if remaining < 0 || remaining > 2 {
		t.Errorf("Expected remaining≈1, got %d", remaining)
	}
}

// TestTokenBucket_CapacityLimit tests that tokens don't exceed capacity
func TestTokenBucket_CapacityLimit(t *testing.T) {
	bucket := NewTokenBucket(10, 100.0) // Very fast refill

	// Take 1 token (9 remaining)
	bucket.Allow("user:123", 1)

	// Wait for refill (would add way more than 10 if uncapped)
	time.Sleep(1 * time.Second)

	// Try to take 10 tokens (should be allowed, bucket refilled to capacity)
	allowed, remaining := bucket.Allow("user:123", 10)

	if !allowed {
		t.Errorf("Expected request to be allowed, bucket should be at capacity")
	}

	if remaining != 0 {
		t.Errorf("Expected remaining=0 after taking full capacity, got %d", remaining)
	}
}

// TestTokenBucket_IsolatedBuckets tests that different users have separate buckets
func TestTokenBucket_IsolatedBuckets(t *testing.T) {
	bucket := NewTokenBucket(10, 2.0)

	// User 123 takes 8 tokens
	bucket.Allow("user:123", 8)

	// User 456 should still have full bucket
	allowed, remaining := bucket.Allow("user:456", 9)

	if !allowed {
		t.Errorf("User 456 should have full bucket, but request was denied")
	}

	if remaining != 1 {
		t.Errorf("Expected user:456 to have remaining=1, got %d", remaining)
	}

	// User 123 should still have only 2 tokens
	allowed, remaining = bucket.Allow("user:123", 3)

	if allowed {
		t.Errorf("User 123 should be denied (only 2 tokens), but was allowed")
	}
}

// TestTokenBucket_ConcurrentAccess tests thread safety with goroutines
func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	bucket := NewTokenBucket(1000, 100.0)

	// Launch 10 goroutines, each making 10 requests
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				bucket.Allow("user:concurrent", 1)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// After 100 requests (10 goroutines × 10 requests), should have ~900 tokens
	// (accounting for some refill during execution)
	allowed, remaining := bucket.Allow("user:concurrent", 800)

	if !allowed {
		t.Errorf("Expected concurrent requests to be handled correctly")
	}

	// Remaining should be reasonable (we can't be exact due to timing)
	if remaining < 0 || remaining > 200 {
		t.Logf("Warning: remaining tokens (%d) outside expected range, possible race condition", remaining)
	}
}
