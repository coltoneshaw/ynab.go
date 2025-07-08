package api

import (
	"context"
	"fmt"
	"sync"
)

// TokenProvider defines the interface for providing access tokens to the YNAB API client.
// This abstraction allows for different authentication methods including static API keys
// and OAuth tokens with automatic refresh capabilities.
type TokenProvider interface {
	// GetAccessToken returns the current access token for API authentication.
	// It should return an error if no valid token is available.
	GetAccessToken(ctx context.Context) (string, error)

	// IsAuthenticated returns true if the provider currently has a valid token.
	IsAuthenticated() bool

	// SetAccessToken updates the current access token.
	// This enables hot-swapping of tokens during runtime.
	SetAccessToken(token string) error

	// GetAccessTokenString returns the current token as a string without context.
	// This is provided for convenience and backward compatibility.
	GetAccessTokenString() string
}

// StaticTokenProvider implements TokenProvider for static API keys.
// This is suitable for scenarios where the token doesn't change or is managed externally.
type StaticTokenProvider struct {
	mu    sync.RWMutex
	token string
}

// NewStaticTokenProvider creates a new StaticTokenProvider with the given token.
func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{
		token: token,
	}
}

// GetAccessToken returns the current static token.
func (p *StaticTokenProvider) GetAccessToken(ctx context.Context) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.token, nil
}

// IsAuthenticated returns true if a token is set.
func (p *StaticTokenProvider) IsAuthenticated() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.token != ""
}

// SetAccessToken updates the token, enabling hot-swapping at runtime.
func (p *StaticTokenProvider) SetAccessToken(token string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.token = token
	return nil
}

// GetAccessTokenString returns the current token without context.
func (p *StaticTokenProvider) GetAccessTokenString() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.token
}

// OAuthTokenProvider implements TokenProvider for OAuth tokens with automatic refresh.
// This wraps the existing TokenManager to provide the TokenProvider interface.
type OAuthTokenProvider struct {
	manager OAuthTokenManager
}

// OAuthTokenManager interface defines the methods we need from a concrete oauth.TokenManager
// We use a concrete interface to avoid type assertion issues
type OAuthTokenManager interface {
	GetAccessToken(ctx context.Context) (string, error)
	IsAuthenticated() bool
}

// NewOAuthTokenProvider creates a new OAuthTokenProvider wrapping a TokenManager.
func NewOAuthTokenProvider(manager OAuthTokenManager) *OAuthTokenProvider {
	return &OAuthTokenProvider{
		manager: manager,
	}
}

// GetAccessToken returns the current OAuth access token, refreshing if necessary.
func (p *OAuthTokenProvider) GetAccessToken(ctx context.Context) (string, error) {
	return p.manager.GetAccessToken(ctx)
}

// IsAuthenticated returns true if the OAuth provider has a valid token.
func (p *OAuthTokenProvider) IsAuthenticated() bool {
	return p.manager.IsAuthenticated()
}

// SetAccessToken is not supported for OAuth tokens as they are managed by the TokenManager.
// OAuth tokens should be managed through the OAuth flow or TokenManager directly.
func (p *OAuthTokenProvider) SetAccessToken(token string) error {
	return fmt.Errorf("SetAccessToken not supported for OAuth tokens - tokens are managed by OAuth flow")
}

// GetAccessTokenString returns the current token without context.
// This will return empty string if token retrieval fails.
func (p *OAuthTokenProvider) GetAccessTokenString() string {
	token, err := p.manager.GetAccessToken(context.Background())
	if err != nil {
		return ""
	}
	return token
}
