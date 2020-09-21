package tokenbucket

import (
	"math"
	"sync"
	"time"
)

// RateLimiter is a thread-safe wrapper around a map of buckets and an easy to
// use API for generic throttling.
type RateLimiter struct {
	mu      sync.RWMutex
	freq    time.Duration
	buckets map[string]*Bucket
	closing chan struct{}
}

// NewRateLimiter returns a RateLimiter with a single filler go-routine for all
// its Buckets which ticks every freq.
// The number of tokens added on each tick for each bucket is computed
// dynamically to be even accross the duration of a second.
//
// If freq <= 0, the filling go-routine won't be started.
func NewRateLimiter(freq time.Duration) *RateLimiter {
	th := &RateLimiter{
		freq:    freq,
		buckets: map[string]*Bucket{},
		closing: make(chan struct{}),
	}

	if freq > 0 {
		go th.fill(freq)
	}

	return th
}

// Bucket returns a Bucket with rate capacity, keyed by key.
//
// If a Bucket (key, rate) doesn't exist yet, it is created.
//
// You must call Close when you're done with the RateLimiter in order to not leak
// a go-routine and a system-timer.
func (t *RateLimiter) InBucket(key string, rate int64) *Bucket {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.buckets[key]

	if !ok {
		b = NewBucket(rate, -1)
		//get
		b.inc = int64(math.Floor(.5 + (float64(b.capacity) * t.freq.Seconds())))
		b.freq = t.freq
		t.buckets[key] = b
	}

	return b
}

func (t *RateLimiter) OutBucket(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.buckets[key]

	if !ok {
		//
		return false
	}

	b.Close()
	delete(t.buckets, key)
	return true

}

// Wait waits for n amount of tokens to be available.
// If n tokens are immediatelly available it doesn't sleep. Otherwise, it sleeps
// the minimum amount of time required for the remaining tokens to be available.
// It returns the wait duration.
//
// If a Bucket (key, rate) doesn't exist yet, it is created.
// If freq < 1/rate seconds, the effective wait rate won't be correct.
//
// You must call Close when you're done with the RateLimiter in order to not leak
// a go-routine and a system-timer.
func (t *RateLimiter) Wait(key string, n, rate int64) time.Duration {
	return t.InBucket(key, rate).Wait(n)
}

func (t *RateLimiter) CanIThroughIt(key string, n, rate int64) bool {
	b := t.InBucket(key, rate)

	if got := b.Take(n); got != n {
		return true
	}

	return false
}

// Close stops filling the Buckets, closing the filling go-routine.
func (t *RateLimiter) Close() error {
	close(t.closing)

	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, b := range t.buckets {
		b.Close()
	}

	return nil
}

func (t *RateLimiter) fill(freq time.Duration) {
	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	for _ = range ticker.C {
		select {
		case <-t.closing:
			return
		default:
		}
		t.mu.RLock()
		for _, b := range t.buckets {
			b.Put(b.inc)
		}
		t.mu.RUnlock()
	}
}
