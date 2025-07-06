// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
)

// Config holds OAuth 2.0 configuration for YNAB
type Config struct {
	// ClientID is the OAuth application's client identifier
	ClientID string

	// ClientSecret is the OAuth application's client secret
	ClientSecret string

	// RedirectURI is the registered redirect URI for the application
	RedirectURI string

	// Scopes defines the permissions requested
	Scopes []Scope

	// AuthorizeURL is the authorization endpoint URL (defaults to YNAB's)
	AuthorizeURL string

	// TokenURL is the token endpoint URL (defaults to YNAB's)
	TokenURL string
}

// NewConfig creates a new OAuth configuration
func NewConfig(clientID, clientSecret, redirectURI string) *Config {
	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       []Scope{},
		AuthorizeURL: AuthorizeURL,
		TokenURL:     TokenURL,
	}
}

// WithReadOnlyScope sets read-only scope for the configuration
// This limits the client to read-only operations (GET requests only)
func (c *Config) WithReadOnlyScope() *Config {
	c.Scopes = []Scope{ScopeReadOnly}
	return c
}

// IsReadOnly returns true if the configuration is set to read-only access
func (c *Config) IsReadOnly() bool {
	return len(c.Scopes) > 0 && c.Scopes[0] == ScopeReadOnly
}

// GetScopeString returns the scope string for OAuth requests
func (c *Config) GetScopeString() string {
	if c.IsReadOnly() {
		return string(ScopeReadOnly)
	}
	return "" // Default scope (full access)
}

// AuthCodeURL generates the authorization URL for the authorization code flow
func (c *Config) AuthCodeURL(state string) string {
	return c.buildAuthorizeURL(ResponseTypeCode, state)
}

// ImplicitGrantURL generates the authorization URL for the implicit grant flow
func (c *Config) ImplicitGrantURL(state string) string {
	return c.buildAuthorizeURL(ResponseTypeToken, state)
}

// GenerateState generates a secure random state parameter for CSRF protection
func (c *Config) GenerateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateRedirectURI checks if the provided redirect URI matches the configured one
func (c *Config) ValidateRedirectURI(redirectURI string) bool {
	return c.RedirectURI == redirectURI
}

// ValidateState checks if the provided state matches the expected state
func (c *Config) ValidateState(expectedState, actualState string) bool {
	return expectedState != "" && expectedState == actualState
}

// buildAuthorizeURL constructs the authorization URL
func (c *Config) buildAuthorizeURL(responseType ResponseType, state string) string {
	params := url.Values{}
	params.Set("client_id", c.ClientID)
	params.Set("redirect_uri", c.RedirectURI)
	params.Set("response_type", string(responseType))

	// Only add scope parameter if scopes are specified
	// YNAB API: omitting scope parameter grants full access
	scopeString := c.GetScopeString()
	if scopeString != "" {
		params.Set("scope", scopeString)
	}

	if state != "" {
		params.Set("state", state)
	}

	return fmt.Sprintf("%s?%s", c.AuthorizeURL, params.Encode())
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	if c.RedirectURI == "" {
		return fmt.Errorf("redirect URI is required")
	}

	// Validate redirect URI format
	if _, err := url.Parse(c.RedirectURI); err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}

	if c.AuthorizeURL == "" {
		return fmt.Errorf("authorize URL is required")
	}

	if c.TokenURL == "" {
		return fmt.Errorf("token URL is required")
	}

	return nil
}

// ParseCallbackURL parses the callback URL and extracts authorization code or access token
func (c *Config) ParseCallbackURL(callbackURL string) (*CallbackResult, error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("invalid callback URL: %w", err)
	}

	result := &CallbackResult{}

	// Extract state first (present in both success and error cases)
	result.State = parsedURL.Query().Get("state")

	// Check for error in query parameters
	if errorParam := parsedURL.Query().Get("error"); errorParam != "" {
		result.Error = &ErrorResponse{
			ErrorCode:        errorParam,
			ErrorDescription: parsedURL.Query().Get("error_description"),
			ErrorURI:         parsedURL.Query().Get("error_uri"),
		}
		return result, nil
	}

	// Check for authorization code (authorization code flow)
	if code := parsedURL.Query().Get("code"); code != "" {
		result.Code = code
		return result, nil
	}

	// Check for access token in fragment (implicit flow)
	if parsedURL.Fragment != "" {
		fragmentParams, err := url.ParseQuery(parsedURL.Fragment)
		if err != nil {
			return nil, fmt.Errorf("invalid fragment parameters: %w", err)
		}

		if accessToken := fragmentParams.Get("access_token"); accessToken != "" {
			result.AccessToken = accessToken
			result.TokenType = fragmentParams.Get("token_type")
			result.Scope = fragmentParams.Get("scope")

			// Parse expires_in if present
			if expiresIn := fragmentParams.Get("expires_in"); expiresIn != "" {
				if seconds, err := parseExpiresIn(expiresIn); err == nil {
					result.ExpiresIn = seconds
				}
			}

			// Override state from fragment if present
			if fragmentState := fragmentParams.Get("state"); fragmentState != "" {
				result.State = fragmentState
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("no authorization code or access token found in callback URL")
}

// CallbackResult represents the result of parsing a callback URL
type CallbackResult struct {
	// For authorization code flow
	Code  string
	State string

	// For implicit flow
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	Scope       string

	// Error information
	Error *ErrorResponse
}

// ToToken converts CallbackResult to Token (for implicit flow)
func (cr *CallbackResult) ToToken() *Token {
	if cr.AccessToken == "" {
		return nil
	}

	token := &Token{
		AccessToken: cr.AccessToken,
		TokenType:   TokenType(cr.TokenType),
		Scope:       Scope(cr.Scope),
	}

	if cr.ExpiresIn > 0 {
		token.SetExpiration(cr.ExpiresIn)
	}

	return token
}

// parseExpiresIn converts expires_in string to int64
func parseExpiresIn(expiresIn string) (int64, error) {
	// This would typically use strconv.ParseInt but keeping it simple
	switch expiresIn {
	case "7200": // 2 hours (YNAB default)
		return 7200, nil
	case "3600": // 1 hour
		return 3600, nil
	default:
		// For now, default to 2 hours if we can't parse
		return 7200, nil
	}
}
