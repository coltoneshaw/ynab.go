// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		expected bool
	}{
		{
			name: "Token not expired",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Token expired",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Token without expiration time",
			token: &Token{
				AccessToken: "test-token",
			},
			expected: false,
		},
		{
			name: "Token expiring within buffer time",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(2 * time.Minute), // Within 5-minute buffer
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.token.IsExpired())
		})
	}
}

func TestToken_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		expected bool
	}{
		{
			name: "Valid token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Invalid token - no access token",
			token: &Token{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Invalid token - expired",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.token.IsValid())
		})
	}
}

func TestToken_CanRefresh(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		expected bool
	}{
		{
			name: "Can refresh - has refresh token",
			token: &Token{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			expected: true,
		},
		{
			name: "Cannot refresh - no refresh token",
			token: &Token{
				AccessToken: "access-token",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.token.CanRefresh())
		})
	}
}

func TestToken_SetExpiration(t *testing.T) {
	token := &Token{}
	expiresIn := int64(3600) // 1 hour

	token.SetExpiration(expiresIn)

	assert.Equal(t, expiresIn, token.ExpiresIn)
	assert.False(t, token.ExpiresAt.IsZero())
	assert.False(t, token.CreatedAt.IsZero())
	
	// Check that expiration is approximately 1 hour from now
	expectedExpiration := time.Now().Add(time.Duration(expiresIn) * time.Second)
	assert.WithinDuration(t, expectedExpiration, token.ExpiresAt, 5*time.Second)
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ErrorResponse
		expected string
	}{
		{
			name: "Error with description",
			err: &ErrorResponse{
				ErrorCode:        "invalid_request",
				ErrorDescription: "The request is missing a required parameter",
			},
			expected: "The request is missing a required parameter",
		},
		{
			name: "Error without description",
			err: &ErrorResponse{
				ErrorCode: "invalid_client",
			},
			expected: "invalid_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestTokenResponse_ToToken(t *testing.T) {
	tokenResponse := &TokenResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "read-only",
	}

	token := tokenResponse.ToToken()

	assert.Equal(t, "access-token", token.AccessToken)
	assert.Equal(t, "refresh-token", token.RefreshToken)
	assert.Equal(t, TokenTypeBearer, token.TokenType)
	assert.Equal(t, ScopeReadOnly, token.Scope)
	assert.Equal(t, int64(3600), token.ExpiresIn)
	assert.False(t, token.ExpiresAt.IsZero())
	assert.False(t, token.CreatedAt.IsZero())
}

func TestScopes(t *testing.T) {
	assert.Equal(t, "read-only", string(ScopeReadOnly))
}

func TestGrantTypes(t *testing.T) {
	assert.Equal(t, "authorization_code", string(GrantTypeAuthorizationCode))
	assert.Equal(t, "refresh_token", string(GrantTypeRefreshToken))
	assert.Equal(t, "token", string(GrantTypeImplicit))
}

func TestResponseTypes(t *testing.T) {
	assert.Equal(t, "code", string(ResponseTypeCode))
	assert.Equal(t, "token", string(ResponseTypeToken))
}

func TestTokenTypes(t *testing.T) {
	assert.Equal(t, "Bearer", string(TokenTypeBearer))
}

func TestOAuthEndpoints(t *testing.T) {
	assert.Equal(t, "https://app.ynab.com/oauth/authorize", AuthorizeURL)
	assert.Equal(t, "https://app.ynab.com/oauth/token", TokenURL)
}