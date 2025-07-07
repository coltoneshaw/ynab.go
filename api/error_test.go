// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	err := &Error{
		ID:     "403.1",
		Name:   "subscription_lapsed",
		Detail: "Subscription for account has lapsed",
	}

	expected := "api: error id=403.1 name=subscription_lapsed detail=Subscription for account has lapsed"
	assert.Equal(t, expected, err.Error())
}

func TestError_IsSubscriptionLapsed(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"subscription lapsed", ErrorSubscriptionLapsed, true},
		{"trial expired", ErrorTrialExpired, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsSubscriptionLapsed())
		})
	}
}

func TestError_IsTrialExpired(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"trial expired", ErrorTrialExpired, true},
		{"subscription lapsed", ErrorSubscriptionLapsed, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsTrialExpired())
		})
	}
}

func TestError_IsAccountError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"subscription lapsed", ErrorSubscriptionLapsed, true},
		{"trial expired", ErrorTrialExpired, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
		{"not found", ErrorNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsAccountError())
		})
	}
}

func TestError_IsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"unauthorized", ErrorUnauthorized, true},
		{"unauthorized scope", ErrorUnauthorizedScope, false},
		{"subscription lapsed", ErrorSubscriptionLapsed, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsUnauthorized())
		})
	}
}

func TestError_IsUnauthorizedScope(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"unauthorized scope", ErrorUnauthorizedScope, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"subscription lapsed", ErrorSubscriptionLapsed, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsUnauthorizedScope())
		})
	}
}

func TestError_IsAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"unauthorized", ErrorUnauthorized, true},
		{"unauthorized scope", ErrorUnauthorizedScope, true},
		{"subscription lapsed", ErrorSubscriptionLapsed, false},
		{"rate limit", ErrorRateLimit, false},
		{"not found", ErrorNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsAuthenticationError())
		})
	}
}

func TestError_IsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"not found", ErrorNotFound, true},
		{"resource not found", ErrorResourceNotFound, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
		{"conflict", ErrorConflict, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsNotFound())
		})
	}
}

func TestError_IsConflict(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"conflict", ErrorConflict, true},
		{"not found", ErrorNotFound, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsConflict())
		})
	}
}

func TestError_IsDataLimitReached(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"data limit reached", ErrorDataLimitReached, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
		{"not found", ErrorNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsDataLimitReached())
		})
	}
}

func TestError_IsRateLimit(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"rate limit", ErrorRateLimit, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"not found", ErrorNotFound, false},
		{"conflict", ErrorConflict, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsRateLimit())
		})
	}
}

func TestError_IsInternalServerError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"internal server error", ErrorInternalServer, true},
		{"service unavailable", ErrorServiceUnavailable, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsInternalServerError())
		})
	}
}

func TestError_IsServiceUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"service unavailable", ErrorServiceUnavailable, true},
		{"internal server error", ErrorInternalServer, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsServiceUnavailable())
		})
	}
}

func TestError_IsClientError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"bad request", ErrorBadRequest, true},
		{"unauthorized", ErrorUnauthorized, true},
		{"subscription lapsed", ErrorSubscriptionLapsed, true},
		{"not found", ErrorNotFound, true},
		{"conflict", ErrorConflict, true},
		{"rate limit", ErrorRateLimit, true},
		{"internal server error", ErrorInternalServer, false},
		{"service unavailable", ErrorServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsClientError())
		})
	}
}

func TestError_IsServerError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"internal server error", ErrorInternalServer, true},
		{"service unavailable", ErrorServiceUnavailable, true},
		{"bad request", ErrorBadRequest, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"not found", ErrorNotFound, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsServerError())
		})
	}
}

func TestError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"rate limit", ErrorRateLimit, true},
		{"internal server error", ErrorInternalServer, true},
		{"service unavailable", ErrorServiceUnavailable, true},
		{"bad request", ErrorBadRequest, false},
		{"unauthorized", ErrorUnauthorized, false},
		{"not found", ErrorNotFound, false},
		{"conflict", ErrorConflict, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsRetryable())
		})
	}
}

func TestError_IsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"bad request", ErrorBadRequest, true},
		{"unauthorized", ErrorUnauthorized, false},
		{"not found", ErrorNotFound, false},
		{"rate limit", ErrorRateLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.IsValidationError())
		})
	}
}

func TestError_RequiresUserAction(t *testing.T) {
	tests := []struct {
		name     string
		errorID  string
		expected bool
	}{
		{"subscription lapsed", ErrorSubscriptionLapsed, true},
		{"trial expired", ErrorTrialExpired, true},
		{"unauthorized", ErrorUnauthorized, true},
		{"unauthorized scope", ErrorUnauthorizedScope, true},
		{"data limit reached", ErrorDataLimitReached, true},
		{"bad request", ErrorBadRequest, false},
		{"not found", ErrorNotFound, false},
		{"rate limit", ErrorRateLimit, false},
		{"internal server error", ErrorInternalServer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{ID: tt.errorID}
			assert.Equal(t, tt.expected, err.RequiresUserAction())
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined correctly
	assert.Equal(t, "400", ErrorBadRequest)
	assert.Equal(t, "401", ErrorUnauthorized)
	assert.Equal(t, "403.1", ErrorSubscriptionLapsed)
	assert.Equal(t, "403.2", ErrorTrialExpired)
	assert.Equal(t, "403.3", ErrorUnauthorizedScope)
	assert.Equal(t, "403.4", ErrorDataLimitReached)
	assert.Equal(t, "404.1", ErrorNotFound)
	assert.Equal(t, "404.2", ErrorResourceNotFound)
	assert.Equal(t, "409", ErrorConflict)
	assert.Equal(t, "429", ErrorRateLimit)
	assert.Equal(t, "500", ErrorInternalServer)
	assert.Equal(t, "503", ErrorServiceUnavailable)
}

// Test example usage scenarios
func TestError_UsageScenarios(t *testing.T) {
	t.Run("subscription lapsed scenario", func(t *testing.T) {
		err := &Error{
			ID:     ErrorSubscriptionLapsed,
			Name:   "subscription_lapsed",
			Detail: "Subscription for account has lapsed",
		}

		assert.True(t, err.IsSubscriptionLapsed())
		assert.True(t, err.IsAccountError())
		assert.True(t, err.RequiresUserAction())
		assert.False(t, err.IsRetryable())
		assert.True(t, err.IsClientError())
		assert.False(t, err.IsServerError())
	})

	t.Run("rate limit scenario", func(t *testing.T) {
		err := &Error{
			ID:     ErrorRateLimit,
			Name:   "too_many_requests",
			Detail: "Too many requests",
		}

		assert.True(t, err.IsRateLimit())
		assert.True(t, err.IsRetryable())
		assert.False(t, err.RequiresUserAction())
		assert.True(t, err.IsClientError())
		assert.False(t, err.IsServerError())
	})

	t.Run("authentication error scenario", func(t *testing.T) {
		err := &Error{
			ID:     ErrorUnauthorized,
			Name:   "not_authorized",
			Detail: "Invalid access token",
		}

		assert.True(t, err.IsUnauthorized())
		assert.True(t, err.IsAuthenticationError())
		assert.True(t, err.RequiresUserAction())
		assert.False(t, err.IsRetryable())
		assert.True(t, err.IsClientError())
		assert.False(t, err.IsServerError())
	})

	t.Run("server error scenario", func(t *testing.T) {
		err := &Error{
			ID:     ErrorInternalServer,
			Name:   "internal_server_error",
			Detail: "Unexpected API error occurred",
		}

		assert.True(t, err.IsInternalServerError())
		assert.True(t, err.IsRetryable())
		assert.False(t, err.RequiresUserAction())
		assert.False(t, err.IsClientError())
		assert.True(t, err.IsServerError())
	})
}
