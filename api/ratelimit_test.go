// Copyright (c) 2024, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimitTracker(t *testing.T) {
	tracker := NewRateLimitTracker(100, time.Hour)

	assert.Equal(t, 100, tracker.limit)
	assert.Equal(t, time.Hour, tracker.window)
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 100, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
}

func TestNewYNABRateLimitTracker(t *testing.T) {
	tracker := NewYNABRateLimitTracker()

	assert.Equal(t, 200, tracker.limit)
	assert.Equal(t, time.Hour, tracker.window)
}

func TestRateLimitTracker_RecordRequest(t *testing.T) {
	tracker := NewRateLimitTracker(5, time.Minute)

	// Record some requests
	tracker.RecordRequest()
	assert.Equal(t, 1, tracker.RequestsInWindow())
	assert.Equal(t, 4, tracker.RequestsRemaining())

	tracker.RecordRequest()
	tracker.RecordRequest()
	assert.Equal(t, 3, tracker.RequestsInWindow())
	assert.Equal(t, 2, tracker.RequestsRemaining())
}

func TestRateLimitTracker_IsAtLimit(t *testing.T) {
	tracker := NewRateLimitTracker(3, time.Minute)

	assert.False(t, tracker.IsAtLimit())

	tracker.RecordRequest()
	tracker.RecordRequest()
	assert.False(t, tracker.IsAtLimit())

	tracker.RecordRequest()
	assert.True(t, tracker.IsAtLimit())

	// Even more requests
	tracker.RecordRequest()
	assert.True(t, tracker.IsAtLimit())
}

func TestRateLimitTracker_TimeWindow(t *testing.T) {
	// Use a very short window for testing
	tracker := NewRateLimitTracker(5, 100*time.Millisecond)

	// Record some requests
	tracker.RecordRequest()
	tracker.RecordRequest()
	assert.Equal(t, 2, tracker.RequestsInWindow())

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Requests should be cleaned up
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 5, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
}

func TestRateLimitTracker_TimeUntilReset(t *testing.T) {
	tracker := NewRateLimitTracker(5, time.Minute)

	// No requests recorded
	assert.Equal(t, time.Duration(0), tracker.TimeUntilReset())

	// Record a request
	tracker.RecordRequest()

	// Should have some time until reset
	resetTime := tracker.TimeUntilReset()
	assert.True(t, resetTime > 59*time.Second)
	assert.True(t, resetTime <= time.Minute)
}

func TestRateLimitTracker_Methods(t *testing.T) {
	tracker := NewRateLimitTracker(10, time.Hour)

	// Initial state
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 10, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
	assert.Equal(t, time.Duration(0), tracker.TimeUntilReset())

	// After recording requests
	tracker.RecordRequest()
	tracker.RecordRequest()

	assert.Equal(t, 2, tracker.RequestsInWindow())
	assert.Equal(t, 8, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
	assert.True(t, tracker.TimeUntilReset() > 0)
}

func TestRateLimitTracker_Cleanup(t *testing.T) {
	tracker := NewRateLimitTracker(10, 50*time.Millisecond)

	// Record requests over time
	tracker.RecordRequest()
	time.Sleep(20 * time.Millisecond)
	tracker.RecordRequest()
	time.Sleep(20 * time.Millisecond)
	tracker.RecordRequest()

	assert.Equal(t, 3, tracker.RequestsInWindow())

	// Wait for some to expire
	time.Sleep(40 * time.Millisecond)

	// First request should be cleaned up
	requests := tracker.RequestsInWindow()
	assert.True(t, requests < 3)

	// Wait for all to expire
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, tracker.RequestsInWindow())
}

func TestRateLimitTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewRateLimitTracker(100, time.Minute)

	// Simulate concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 5; j++ {
				tracker.RecordRequest()
				_ = tracker.RequestsInWindow()
				_ = tracker.RequestsRemaining()
				_ = tracker.IsAtLimit()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have recorded 50 requests
	assert.Equal(t, 50, tracker.RequestsInWindow())
	assert.Equal(t, 50, tracker.RequestsRemaining())
}

func TestRateLimitTracker_Reset(t *testing.T) {
	tracker := NewRateLimitTracker(5, time.Minute)

	// Record some requests
	tracker.RecordRequest()
	tracker.RecordRequest()
	tracker.RecordRequest()

	assert.Equal(t, 3, tracker.RequestsInWindow())
	assert.Equal(t, 2, tracker.RequestsRemaining())

	// Reset the tracker
	tracker.Reset()

	// Should be back to initial state
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 5, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
	assert.Equal(t, time.Duration(0), tracker.TimeUntilReset())
}

func TestRateLimitTracker_GetConfiguration(t *testing.T) {
	tracker := NewRateLimitTracker(150, 2*time.Hour)

	assert.Equal(t, 150, tracker.GetLimit())
	assert.Equal(t, 2*time.Hour, tracker.GetWindow())
}

func TestNewCustomYNABRateLimitTracker(t *testing.T) {
	tracker := NewCustomYNABRateLimitTracker(500)

	assert.Equal(t, 500, tracker.GetLimit())
	assert.Equal(t, time.Hour, tracker.GetWindow())
	assert.Equal(t, 500, tracker.RequestsRemaining())
}

func TestRateLimitTracker_ConcurrentReadWrite(t *testing.T) {
	tracker := NewRateLimitTracker(100, time.Minute)

	// Start concurrent readers
	readDone := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = tracker.RequestsInWindow()
				_ = tracker.RequestsRemaining()
				_ = tracker.IsAtLimit()
				_ = tracker.TimeUntilReset()
				time.Sleep(time.Millisecond)
			}
			readDone <- true
		}()
	}

	// Start concurrent writers
	writeDone := make(chan bool, 2)
	for i := 0; i < 2; i++ {
		go func() {
			for j := 0; j < 5; j++ {
				tracker.RecordRequest()
				time.Sleep(time.Millisecond)
			}
			writeDone <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-readDone
	}
	for i := 0; i < 2; i++ {
		<-writeDone
	}

	// Should have recorded 10 requests from writers
	assert.Equal(t, 10, tracker.RequestsInWindow())
	assert.Equal(t, 90, tracker.RequestsRemaining())
}

// Time Precision Tests - Phase 1

func TestRateLimitTracker_MicrosecondPrecision(t *testing.T) {
	// Use a very short window to test precise timing
	tracker := NewRateLimitTracker(5, 10*time.Millisecond)

	// Record a request
	tracker.RecordRequest()
	assert.Equal(t, 1, tracker.RequestsInWindow())

	// Wait just under the window - should still be present
	time.Sleep(8 * time.Millisecond)
	assert.Equal(t, 1, tracker.RequestsInWindow())

	// Wait past the window boundary - should be cleaned up
	time.Sleep(5 * time.Millisecond) // Total: 13ms > 10ms window
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 5, tracker.RequestsRemaining())
}

func TestRateLimitTracker_RapidSequence(t *testing.T) {
	// Test rapid requests in quick succession
	tracker := NewRateLimitTracker(10, 50*time.Millisecond)

	// Record multiple requests rapidly
	for i := 0; i < 5; i++ {
		tracker.RecordRequest()
		if i < 4 { // Don't sleep after the last request
			time.Sleep(1 * time.Millisecond)
		}
	}

	// All requests should be tracked
	assert.Equal(t, 5, tracker.RequestsInWindow())
	assert.Equal(t, 5, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())

	// Wait for window to expire
	time.Sleep(55 * time.Millisecond)

	// All should be cleaned up
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 10, tracker.RequestsRemaining())
}

func TestRateLimitTracker_BoundaryEdge(t *testing.T) {
	// Test requests exactly at window boundaries
	tracker := NewRateLimitTracker(5, 100*time.Millisecond)

	// Record first request
	start := time.Now()
	tracker.RecordRequest()
	assert.Equal(t, 1, tracker.RequestsInWindow())

	// Wait almost to the boundary
	elapsed := time.Since(start)
	waitTime := 100*time.Millisecond - elapsed - 5*time.Millisecond
	if waitTime > 0 {
		time.Sleep(waitTime)
	}

	// Should still be present (just before boundary)
	assert.Equal(t, 1, tracker.RequestsInWindow())

	// Now wait past the boundary
	time.Sleep(10 * time.Millisecond)

	// Should be cleaned up
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 5, tracker.RequestsRemaining())

	// Verify TimeUntilReset behavior
	assert.Equal(t, time.Duration(0), tracker.TimeUntilReset())
}

func TestRateLimitTracker_CleanupTiming(t *testing.T) {
	// Test partial vs full cleanup scenarios
	tracker := NewRateLimitTracker(10, 40*time.Millisecond)

	// Record requests at different times
	tracker.RecordRequest() // Request 1
	time.Sleep(15 * time.Millisecond)

	tracker.RecordRequest() // Request 2
	time.Sleep(15 * time.Millisecond)

	tracker.RecordRequest() // Request 3

	// All 3 should be present
	assert.Equal(t, 3, tracker.RequestsInWindow())

	// Wait for first request to expire (45ms total elapsed)
	time.Sleep(15 * time.Millisecond)

	// Should have partial cleanup - first request expired
	requests := tracker.RequestsInWindow()
	assert.True(t, requests < 3, "Should have fewer than 3 requests after partial cleanup")
	assert.True(t, requests > 0, "Should still have some requests")

	// Wait for all to expire
	time.Sleep(30 * time.Millisecond)

	// Full cleanup - all should be gone
	assert.Equal(t, 0, tracker.RequestsInWindow())
	assert.Equal(t, 10, tracker.RequestsRemaining())
	assert.False(t, tracker.IsAtLimit())
}
