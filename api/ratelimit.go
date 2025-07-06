// Copyright (c) 2024, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package api

import (
	"sync"
	"time"
)

// RateLimitTracker tracks API requests in a rolling time window
// to help users stay within YNAB's 200 requests/hour limit.
// This is completely optional - users can choose whether to use it.
type RateLimitTracker struct {
	requests []time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimitTracker creates a new rate limit tracker.
// For YNAB API, use: NewRateLimitTracker(200, time.Hour)
func NewRateLimitTracker(limit int, window time.Duration) *RateLimitTracker {
	return &RateLimitTracker{
		requests: make([]time.Time, 0),
		limit:    limit,
		window:   window,
	}
}

// NewYNABRateLimitTracker creates a tracker configured for YNAB's limits:
// 200 requests per hour in a rolling window
func NewYNABRateLimitTracker() *RateLimitTracker {
	return NewRateLimitTracker(200, time.Hour)
}

// NewCustomYNABRateLimitTracker creates a tracker with custom requests per hour for YNAB.
// Useful if YNAB changes their rate limits or for testing with different limits.
func NewCustomYNABRateLimitTracker(requestsPerHour int) *RateLimitTracker {
	return NewRateLimitTracker(requestsPerHour, time.Hour)
}

// RecordRequest records that an API request was made at the current time.
// Call this after making any YNAB API request.
func (r *RateLimitTracker) RecordRequest() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.requests = append(r.requests, time.Now())
	r.cleanup()
}

// RequestsInWindow returns the number of requests made in the current rolling window
func (r *RateLimitTracker) RequestsInWindow() int {
	r.mutex.RLock()

	// Quick check if cleanup needed
	if r.needsCleanup() {
		r.mutex.RUnlock()
		r.mutex.Lock()
		r.cleanup()
		count := len(r.requests)
		r.mutex.Unlock()
		return count
	}

	count := len(r.requests)
	r.mutex.RUnlock()
	return count
}

// RequestsRemaining returns how many requests can be made before hitting the limit
func (r *RateLimitTracker) RequestsRemaining() int {
	remaining := r.limit - r.RequestsInWindow()
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeUntilReset returns the duration until the oldest request falls out of the rolling window,
// which would free up one request slot. Returns 0 if no requests are recorded.
//
// Example: If you made 200 API calls over the last 50 minutes, this returns ~10 minutes
// (the time until the oldest request will be 1 hour old and fall off the rolling window).
func (r *RateLimitTracker) TimeUntilReset() time.Duration {
	r.mutex.RLock()

	// Quick check if cleanup needed
	if r.needsCleanup() {
		r.mutex.RUnlock()
		r.mutex.Lock()
		r.cleanup()
		defer r.mutex.Unlock()
	} else {
		defer r.mutex.RUnlock()
	}

	if len(r.requests) == 0 {
		return 0
	}

	oldest := r.requests[0]
	resetTime := oldest.Add(r.window)

	if resetTime.Before(time.Now()) {
		return 0
	}

	return time.Until(resetTime)
}

// IsAtLimit returns true if the rate limit has been reached
func (r *RateLimitTracker) IsAtLimit() bool {
	return r.RequestsInWindow() >= r.limit
}

// Reset clears all recorded requests from the tracker.
// Useful for testing or when you want to start fresh.
func (r *RateLimitTracker) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.requests = r.requests[:0]
}

// GetLimit returns the configured rate limit (requests per window)
func (r *RateLimitTracker) GetLimit() int {
	return r.limit
}

// GetWindow returns the configured time window duration
func (r *RateLimitTracker) GetWindow() time.Duration {
	return r.window
}

// needsCleanup checks if cleanup is needed without modifying state.
// Must be called with at least a read lock held.
func (r *RateLimitTracker) needsCleanup() bool {
	if len(r.requests) == 0 {
		return false
	}

	cutoff := time.Now().Add(-r.window)
	return r.requests[0].Before(cutoff) || r.requests[0].Equal(cutoff)
}

// cleanup removes requests that are outside the rolling window
// Must be called with a write lock held.
func (r *RateLimitTracker) cleanup() {
	cutoff := time.Now().Add(-r.window)

	// Find the first request that's still within the window
	for i, reqTime := range r.requests {
		if reqTime.After(cutoff) {
			// Keep requests from index i onwards
			r.requests = r.requests[i:]
			return
		}
	}

	// All requests are outside the window
	r.requests = r.requests[:0]
}
