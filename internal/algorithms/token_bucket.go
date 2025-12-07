package algorithms

import (
	"sync"
	"time"
)

/*
Will implement the otken bucket rate limiting algorithm
This is an in-memory version for basic testing, will implement Redis later
*/
type TokenBucket struct {
	capacity   int     //Max tokens the bucket can hold
	refillRate float64 //Tokens added per second

	//basic in memory state, hold a map of strings (users) to their respective buckets
	mu      sync.Mutex
	buckets map[string]*bucketState
}

// state of a single identifier
type bucketState struct {
	tokens     float64   //current tokens
	lastRefill time.Time //last time refill was calculated
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*bucketState),
	}
}

func (tb *TokenBucket) Allow(id string, tokensRequested int) (bool, int) {
	//lock the single instance of tb so that this request can be fully completed
	tb.mu.Lock()
	defer tb.mu.Unlock()

	state, exist := tb.buckets[id]

	if !exist {
		//first time, make a new bucket
		state = &bucketState{
			tokens:     float64(tb.capacity), //starts full
			lastRefill: time.Now(),
		}

		tb.buckets[id] = state
	}

	//We add tokens on operations based on how much time has elapsed since the last operation
	now := time.Now()
	elapsed := now.Sub(state.lastRefill).Seconds()

	tokensToAdd := elapsed * tb.refillRate

	state.tokens = min(tokensToAdd+state.tokens, float64(tb.capacity))
	state.lastRefill = now

	//update or reject
	if state.tokens >= float64(tokensRequested) {
		state.tokens -= float64(tokensRequested)
		return true, int(state.tokens)
	}

	return false, int(state.tokens)

}
