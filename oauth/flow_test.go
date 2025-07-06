// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestAuthorizationCodeFlow_GetAuthorizationURL(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	authURL, err := flow.GetAuthorizationURL("test-state")
	
	assert.NoError(t, err)
	
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)
	
	params := parsedURL.Query()
	assert.Equal(t, "test-client", params.Get("client_id"))
	assert.Equal(t, "code", params.Get("response_type"))
	assert.Equal(t, "test-state", params.Get("state"))
}

func TestAuthorizationCodeFlow_GetAuthorizationURL_InvalidConfig(t *testing.T) {
	config := &Config{} // Invalid config
	flow := NewAuthorizationCodeFlow(config)
	
	_, err := flow.GetAuthorizationURL("test-state")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestAuthorizationCodeFlow_HandleCallback(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	
	// Mock token exchange endpoint
	httpmock.RegisterResponder(http.MethodPost, TokenURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{
				"access_token": "access-token-123",
				"refresh_token": "refresh-token-123",
				"token_type": "Bearer",
				"expires_in": 7200,
				"scope": "read-only"
			}`), nil
		},
	)
	
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	callbackURL := "https://example.com/callback?code=auth-code-123&state=test-state"
	
	token, err := flow.HandleCallback(callbackURL, "test-state")
	
	assert.NoError(t, err)
	assert.Equal(t, "access-token-123", token.AccessToken)
	assert.Equal(t, "refresh-token-123", token.RefreshToken)
	assert.Equal(t, TokenTypeBearer, token.TokenType)
	assert.Equal(t, ScopeReadOnly, token.Scope)
	assert.Equal(t, int64(7200), token.ExpiresIn)
}

func TestAuthorizationCodeFlow_HandleCallback_StateValidation(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	callbackURL := "https://example.com/callback?code=auth-code-123&state=wrong-state"
	
	_, err := flow.HandleCallback(callbackURL, "expected-state")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state parameter mismatch")
}

func TestAuthorizationCodeFlow_HandleCallback_ErrorResponse(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	callbackURL := "https://example.com/callback?error=access_denied&error_description=User%20denied%20access&state=test-state"
	
	_, err := flow.HandleCallback(callbackURL, "test-state")
	
	assert.Error(t, err)
	errorResp, ok := err.(*ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "access_denied", errorResp.ErrorCode)
	assert.Equal(t, "User denied access", errorResp.ErrorDescription)
}

func TestAuthorizationCodeFlow_HandleCallback_NoCode(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	callbackURL := "https://example.com/callback?state=test-state"
	
	_, err := flow.HandleCallback(callbackURL, "test-state")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorization code or access token found")
}

func TestAuthorizationCodeFlow_HandleCallbackWithContext(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	
	httpmock.RegisterResponder(http.MethodPost, TokenURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{
				"access_token": "access-token-123",
				"token_type": "Bearer",
				"expires_in": 7200
			}`), nil
		},
	)
	
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewAuthorizationCodeFlow(config)
	
	ctx := context.Background()
	callbackURL := "https://example.com/callback?code=auth-code-123&state=test-state"
	
	token, err := flow.HandleCallbackWithContext(ctx, callbackURL, "test-state")
	
	assert.NoError(t, err)
	assert.Equal(t, "access-token-123", token.AccessToken)
}

func TestImplicitGrantFlow_GetAuthorizationURL(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewImplicitGrantFlow(config)
	
	authURL, err := flow.GetAuthorizationURL("test-state")
	
	assert.NoError(t, err)
	
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)
	
	params := parsedURL.Query()
	assert.Equal(t, "test-client", params.Get("client_id"))
	assert.Equal(t, "token", params.Get("response_type"))
	assert.Equal(t, "test-state", params.Get("state"))
}

func TestImplicitGrantFlow_HandleCallback(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewImplicitGrantFlow(config)
	
	callbackURL := "https://example.com/callback#access_token=token123&token_type=Bearer&expires_in=7200&state=test-state"
	
	token, err := flow.HandleCallback(callbackURL, "test-state")
	
	assert.NoError(t, err)
	assert.Equal(t, "token123", token.AccessToken)
	assert.Equal(t, TokenTypeBearer, token.TokenType)
	assert.Equal(t, int64(7200), token.ExpiresIn)
}

func TestImplicitGrantFlow_HandleCallback_NoToken(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	flow := NewImplicitGrantFlow(config)
	
	callbackURL := "https://example.com/callback#state=test-state"
	
	_, err := flow.HandleCallback(callbackURL, "test-state")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorization code or access token found")
}

func TestFlowManager(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	manager := NewFlowManager(config)
	
	// Test authorization code flow
	authCodeFlow := manager.AuthorizationCode()
	assert.NotNil(t, authCodeFlow)
	
	// Test implicit grant flow
	implicitFlow := manager.ImplicitGrant()
	assert.NotNil(t, implicitFlow)
	
	// Test GetFlow
	assert.Equal(t, authCodeFlow, manager.GetFlow(ResponseTypeCode))
	assert.Equal(t, implicitFlow, manager.GetFlow(ResponseTypeToken))
	assert.Equal(t, authCodeFlow, manager.GetFlow("unknown")) // Should default to auth code
}

func TestFlowManager_StartAuthorizationCodeFlow(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	manager := NewFlowManager(config)
	
	authURL, state, err := manager.StartAuthorizationCodeFlow()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "state="+state)
}

func TestFlowManager_StartImplicitGrantFlow(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	manager := NewFlowManager(config)
	
	authURL, state, err := manager.StartImplicitGrantFlow()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "response_type=token")
	assert.Contains(t, authURL, "state="+state)
}

func TestFlowManager_CompleteAuthorizationCodeFlow(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	
	httpmock.RegisterResponder(http.MethodPost, TokenURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{
				"access_token": "access-token-123",
				"token_type": "Bearer"
			}`), nil
		},
	)
	
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	manager := NewFlowManager(config)
	
	ctx := context.Background()
	callbackURL := "https://example.com/callback?code=auth-code&state=test-state"
	
	token, err := manager.CompleteAuthorizationCodeFlow(ctx, callbackURL, "test-state")
	
	assert.NoError(t, err)
	assert.Equal(t, "access-token-123", token.AccessToken)
}

func TestFlowManager_CompleteImplicitGrantFlow(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	manager := NewFlowManager(config)
	
	callbackURL := "https://example.com/callback#access_token=token123&token_type=Bearer"
	
	token, err := manager.CompleteImplicitGrantFlow(callbackURL, "")
	
	assert.NoError(t, err)
	assert.Equal(t, "token123", token.AccessToken)
}

func TestFlowManager_WithDefaultStorage(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	storage := NewMemoryStorage()
	
	manager := NewFlowManager(config).WithDefaultStorage(storage)
	
	assert.Same(t, storage, manager.defaultStorage)
	assert.Same(t, storage, manager.authCodeFlow.tokenManager.storage)
}

func TestFlowManager_WithHTTPClient(t *testing.T) {
	config := NewConfig("test-client", "test-secret", "https://example.com/callback")
	httpClient := &http.Client{}
	
	manager := NewFlowManager(config).WithHTTPClient(httpClient)
	
	assert.Same(t, httpClient, manager.authCodeFlow.tokenManager.client)
}

func TestRecommendFlow(t *testing.T) {
	tests := []struct {
		name              string
		isServerSide      bool
		needsRefreshToken bool
		expected          ResponseType
	}{
		{
			name:              "Server-side with refresh token",
			isServerSide:      true,
			needsRefreshToken: true,
			expected:          ResponseTypeCode,
		},
		{
			name:              "Server-side without refresh token",
			isServerSide:      true,
			needsRefreshToken: false,
			expected:          ResponseTypeCode,
		},
		{
			name:              "Client-side",
			isServerSide:      false,
			needsRefreshToken: false,
			expected:          ResponseTypeToken,
		},
		{
			name:              "Client-side with refresh token need",
			isServerSide:      false,
			needsRefreshToken: true,
			expected:          ResponseTypeToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RecommendFlow(tt.isServerSide, tt.needsRefreshToken)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthorizationCodeFlow_WithTokenManager(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "redirect-uri")
	tokenManager := NewTokenManager(config, NewMemoryStorage())
	
	flow := NewAuthorizationCodeFlow(config).WithTokenManager(tokenManager)
	
	assert.Same(t, tokenManager, flow.tokenManager)
}

func TestAuthorizationCodeFlow_WithHTTPClient(t *testing.T) {
	config := NewConfig("client-id", "client-secret", "redirect-uri")
	httpClient := &http.Client{}
	
	flow := NewAuthorizationCodeFlow(config).WithHTTPClient(httpClient)
	
	assert.Same(t, httpClient, flow.tokenManager.client)
}