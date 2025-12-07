// examples/token_bucket_demo.go
package main

import (
	"fmt"
	"time"

	"github.com/dyubeaton/distributed-rate-limiter/internal/algorithms"
)

func main() {
	fmt.Println("=== Token Bucket Algorithm Demo ===\n")

	// Create a bucket: 10 tokens capacity, refills at 2 tokens/second
	bucket := algorithms.NewTokenBucket(10, 2.0)

	// Test 1: Initial request should be allowed (bucket starts full)
	fmt.Println("Test 1: Initial request (bucket full)")
	allowed, remaining := bucket.Allow("user:123", 3)
	fmt.Printf("  Requested: 3 tokens\n")
	fmt.Printf("  Allowed: %v\n", allowed)
	fmt.Printf("  Remaining: %d tokens\n\n", remaining)

	// Test 2: Immediate second request
	fmt.Println("Test 2: Immediate second request")
	allowed, remaining = bucket.Allow("user:123", 5)
	fmt.Printf("  Requested: 5 tokens\n")
	fmt.Printf("  Allowed: %v\n", allowed)
	fmt.Printf("  Remaining: %d tokens\n\n", remaining)

	// Test 3: Request too many tokens (should be denied)
	fmt.Println("Test 3: Request more than remaining")
	allowed, remaining = bucket.Allow("user:123", 10)
	fmt.Printf("  Requested: 10 tokens\n")
	fmt.Printf("  Allowed: %v\n", allowed)
	fmt.Printf("  Remaining: %d tokens\n\n", remaining)

	// Test 4: Wait for refill and try again
	fmt.Println("Test 4: Wait 3 seconds for refill (should add 6 tokens)")
	fmt.Println("  Waiting...")
	time.Sleep(3 * time.Second)
	allowed, remaining = bucket.Allow("user:123", 5)
	fmt.Printf("  Requested: 5 tokens\n")
	fmt.Printf("  Allowed: %v (should be true now)\n", allowed)
	fmt.Printf("  Remaining: %d tokens\n\n", remaining)

	// Test 5: Different user (separate bucket)
	fmt.Println("Test 5: Different user gets their own bucket")
	allowed, remaining = bucket.Allow("user:456", 8)
	fmt.Printf("  Requested: 8 tokens\n")
	fmt.Printf("  Allowed: %v\n", allowed)
	fmt.Printf("  Remaining: %d tokens\n\n", remaining)

	fmt.Println("=== Demo Complete ===")
}
