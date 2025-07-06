// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")
	
	assert.Equal(t, "client-id", config.ClientID)
	assert.Equal(t, "client-secret", config.ClientSecret)
	assert.Equal(t, "https://example.com/callback", config.RedirectURI)
	assert.Equal(t, AuthorizeURL, config.AuthorizeURL)
	assert.Equal(t, TokenURL, config.TokenURL)
	assert.Empty(t, config.Scopes)
}


func TestConfig_WithReadOnlyScope(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")
	
	result := config.WithReadOnlyScope()
	
	assert.Same(t, config, result)
	assert.True(t, config.IsReadOnly())
}




func TestConfig_GetScopeString(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Config)
		expected string
	}{
		{
			name:     "No scopes (default full access)",
			setup:    func(c *Config) {},
			expected: "",
		},
		{
			name:     "Read-only scope",
			setup:    func(c *Config) { c.WithReadOnlyScope() },
			expected: "read-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig("client-id", "client-secret", "https://example.com/callback")
			tt.setup(config)
			assert.Equal(t, tt.expected, config.GetScopeString())
		})
	}
}

func TestConfig_AuthCodeURL(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	config.WithReadOnlyScope()
	
	authURL := config.AuthCodeURL("test-state")
	
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)
	
	assert.Equal(t, "app.ynab.com", parsedURL.Host)
	assert.Equal(t, "/oauth/authorize", parsedURL.Path)
	
	params := parsedURL.Query()
	assert.Equal(t, "test-client", params.Get("client_id"))
	assert.Equal(t, "https://example.com/callback", params.Get("redirect_uri"))
	assert.Equal(t, "code", params.Get("response_type"))
	assert.Equal(t, "read-only", params.Get("scope"))
	assert.Equal(t, "test-state", params.Get("state"))
}

func TestConfig_ImplicitGrantURL(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	
	authURL := config.ImplicitGrantURL("test-state")
	
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)
	
	params := parsedURL.Query()
	assert.Equal(t, "token", params.Get("response_type"))
}

func TestConfig_GenerateState(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")
	
	state1, err1 := config.GenerateState()
	state2, err2 := config.GenerateState()
	
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, state1)
	assert.NotEmpty(t, state2)
	assert.NotEqual(t, state1, state2) // Should generate different states
	assert.Len(t, state1, 32) // 16 bytes hex-encoded = 32 characters
}

func TestConfig_ValidateRedirectURI(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")
	
	assert.True(t, config.ValidateRedirectURI("https://example.com/callback"))
	assert.False(t, config.ValidateRedirectURI("https://different.com/callback"))
}

func TestConfig_ValidateState(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")
	
	assert.True(t, config.ValidateState("test-state", "test-state"))
	assert.False(t, config.ValidateState("test-state", "different-state"))
	assert.False(t, config.ValidateState("", "any-state")) // Empty expected state should fail
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: &Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURI:  "https://example.com/callback",
				AuthorizeURL: AuthorizeURL,
				TokenURL:     TokenURL,
			},
			shouldError: false,
		},
		{
			name: "Missing client ID",
			config: &Config{
				ClientSecret: "client-secret",
				RedirectURI:  "https://example.com/callback",
				AuthorizeURL: AuthorizeURL,
				TokenURL:     TokenURL,
			},
			shouldError: true,
			errorMsg:    "client ID is required",
		},
		{
			name: "Missing client secret",
			config: &Config{
				ClientID:     "client-id",
				RedirectURI:  "https://example.com/callback",
				AuthorizeURL: AuthorizeURL,
				TokenURL:     TokenURL,
			},
			shouldError: true,
			errorMsg:    "client secret is required",
		},
		{
			name: "Missing redirect URI",
			config: &Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				AuthorizeURL: AuthorizeURL,
				TokenURL:     TokenURL,
			},
			shouldError: true,
			errorMsg:    "redirect URI is required",
		},
		{
			name: "Invalid redirect URI",
			config: &Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURI:  ":",
				AuthorizeURL: AuthorizeURL,
				TokenURL:     TokenURL,
			},
			shouldError: true,
			errorMsg:    "invalid redirect URI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ParseCallbackURL(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "https://example.com/callback")

	tests := []struct {
		name        string
		callbackURL string
		expectError bool
		checkResult func(t *testing.T, result *CallbackResult, err error)
	}{
		{
			name:        "Authorization code callback",
			callbackURL: "https://example.com/callback?code=auth-code&state=test-state",
			expectError: false,
			checkResult: func(t *testing.T, result *CallbackResult, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "auth-code", result.Code)
				assert.Equal(t, "test-state", result.State)
				assert.Nil(t, result.Error)
			},
		},
		{
			name:        "Error callback",
			callbackURL: "https://example.com/callback?error=access_denied&error_description=User%20denied%20access&state=test-state",
			expectError: false,
			checkResult: func(t *testing.T, result *CallbackResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result.Error)
				assert.Equal(t, "access_denied", result.Error.ErrorCode)
				assert.Equal(t, "User denied access", result.Error.ErrorDescription)
				assert.Equal(t, "test-state", result.State)
			},
		},
		{
			name:        "Implicit grant callback",
			callbackURL: "https://example.com/callback#access_token=token123&token_type=Bearer&expires_in=7200&state=test-state",
			expectError: false,
			checkResult: func(t *testing.T, result *CallbackResult, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "token123", result.AccessToken)
				assert.Equal(t, "Bearer", result.TokenType)
				assert.Equal(t, int64(7200), result.ExpiresIn)
				assert.Equal(t, "test-state", result.State)
			},
		},
		{
			name:        "Invalid URL",
			callbackURL: ":",
			expectError: true,
			checkResult: func(t *testing.T, result *CallbackResult, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid callback URL")
			},
		},
		{
			name:        "No code or token",
			callbackURL: "https://example.com/callback?state=test-state",
			expectError: true,
			checkResult: func(t *testing.T, result *CallbackResult, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no authorization code or access token found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := config.ParseCallbackURL(tt.callbackURL)
			tt.checkResult(t, result, err)
		})
	}
}

func TestCallbackResult_ToToken(t *testing.T) {
	tests := []struct {
		name     string
		result   *CallbackResult
		expected *Token
	}{
		{
			name: "Valid callback result",
			result: &CallbackResult{
				AccessToken: "token123",
				TokenType:   "Bearer",
				ExpiresIn:   7200,
				Scope:       "read-only",
			},
			expected: &Token{
				AccessToken: "token123",
				TokenType:   TokenTypeBearer,
				Scope:       ScopeReadOnly,
				ExpiresIn:   7200,
			},
		},
		{
			name: "No access token",
			result: &CallbackResult{
				Code: "auth-code",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.result.ToToken()
			
			if tt.expected == nil {
				assert.Nil(t, token)
			} else {
				assert.Equal(t, tt.expected.AccessToken, token.AccessToken)
				assert.Equal(t, tt.expected.TokenType, token.TokenType)
				assert.Equal(t, tt.expected.Scope, token.Scope)
				assert.Equal(t, tt.expected.ExpiresIn, token.ExpiresIn)
				if tt.expected.ExpiresIn > 0 {
					assert.False(t, token.ExpiresAt.IsZero())
				}
			}
		})
	}
}

func TestConfig_buildAuthorizeURL(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	config.WithReadOnlyScope()
	
	url := config.buildAuthorizeURL(ResponseTypeCode, "test-state")
	
	assert.Contains(t, url, "client_id=test-client")
	assert.Contains(t, url, "redirect_uri=https%3A%2F%2Fexample.com%2Fcallback")
	assert.Contains(t, url, "response_type=code")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "scope=read-only")
	
	// Test without scope (full access)
	config2 := NewConfig("test-client", "test-secret", "https://example.com/callback")
	url2 := config2.buildAuthorizeURL(ResponseTypeCode, "test-state")
	assert.NotContains(t, url2, "scope=")
}