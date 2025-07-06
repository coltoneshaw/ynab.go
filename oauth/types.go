// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

// Package oauth implements OAuth 2.0 authentication for YNAB API
package oauth

import (
	"errors"
	"time"
)

// Scope represents OAuth permission scope
type Scope string

const (
	// ScopeReadOnly limits access to read-only operations (GET requests only)
	// When this scope is used, any attempt to modify data (POST, PATCH, PUT, DELETE)
	// will result in a 403 Forbidden response from the YNAB API
	ScopeReadOnly Scope = "read-only"
)

// GrantType represents OAuth grant type
type GrantType string

const (
	// GrantTypeAuthorizationCode for server-side applications
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	// GrantTypeRefreshToken for refreshing access tokens
	GrantTypeRefreshToken GrantType = "refresh_token"
	// GrantTypeImplicit for client-side applications (implicit flow)
	GrantTypeImplicit GrantType = "token"
)

// ResponseType represents OAuth response type
type ResponseType string

const (
	// ResponseTypeCode for authorization code flow
	ResponseTypeCode ResponseType = "code"
	// ResponseTypeToken for implicit grant flow
	ResponseTypeToken ResponseType = "token"
)

// YNAB OAuth endpoints
const (
	AuthorizeURL = "https://app.ynab.com/oauth/authorize"
	TokenURL     = "https://app.ynab.com/oauth/token"
)

// Common OAuth errors
var (
	ErrInvalidClient       = errors.New("invalid client credentials")
	ErrInvalidGrant        = errors.New("invalid grant")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInvalidScope        = errors.New("invalid scope")
	ErrUnauthorizedClient  = errors.New("unauthorized client")
	ErrUnsupportedGrant    = errors.New("unsupported grant type")
	ErrAccessDenied        = errors.New("access denied")
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenRefreshFailed  = errors.New("token refresh failed")
)

// TokenType represents the type of token
type TokenType string

const (
	// TokenTypeBearer is the standard OAuth 2.0 bearer token
	TokenTypeBearer TokenType = "Bearer"
)

// Token represents an OAuth 2.0 token
type Token struct {
	// AccessToken is the token used to access YNAB API
	AccessToken string `json:"access_token"`
	
	// RefreshToken is used to obtain new access tokens
	RefreshToken string `json:"refresh_token,omitempty"`
	
	// TokenType is typically "Bearer"
	TokenType TokenType `json:"token_type"`
	
	// ExpiresIn is the number of seconds the token is valid
	ExpiresIn int64 `json:"expires_in"`
	
	// Scope is the granted permission scope
	Scope Scope `json:"scope,omitempty"`
	
	// ExpiresAt is the calculated expiration time
	ExpiresAt time.Time `json:"expires_at"`
	
	// CreatedAt is when the token was created/refreshed
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	
	// Add 5 minute buffer to account for clock skew and network delays
	buffer := 5 * time.Minute
	return time.Now().Add(buffer).After(t.ExpiresAt)
}

// IsValid checks if the token is valid and not expired
func (t *Token) IsValid() bool {
	return t.AccessToken != "" && !t.IsExpired()
}

// CanRefresh checks if the token can be refreshed
func (t *Token) CanRefresh() bool {
	return t.RefreshToken != ""
}

// SetExpiration calculates and sets the expiration time
func (t *Token) SetExpiration(expiresIn int64) {
	t.ExpiresIn = expiresIn
	t.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	t.CreatedAt = time.Now()
}

// ErrorResponse represents an OAuth error response
type ErrorResponse struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	return e.ErrorCode
}

// AuthorizeParams holds parameters for authorization URL generation
type AuthorizeParams struct {
	ClientID     string
	RedirectURI  string
	ResponseType ResponseType
	Scope        Scope
	State        string
}

// TokenRequest represents a token exchange request
type TokenRequest struct {
	GrantType    GrantType `json:"grant_type"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	Code         string    `json:"code,omitempty"`          // For authorization_code grant
	RedirectURI  string    `json:"redirect_uri,omitempty"`  // For authorization_code grant
	RefreshToken string    `json:"refresh_token,omitempty"` // For refresh_token grant
}

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// ToToken converts TokenResponse to Token
func (tr *TokenResponse) ToToken() *Token {
	token := &Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    TokenType(tr.TokenType),
		Scope:        Scope(tr.Scope),
	}
	
	if tr.ExpiresIn > 0 {
		token.SetExpiration(tr.ExpiresIn)
	}
	
	return token
}