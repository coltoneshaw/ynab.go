package oauth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/coltoneshaw/ynab.go/oauth"
)

func TestNewOAuthClient(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()
	tokenManager := oauth.NewTokenManager(config, storage)

	client := oauth.NewOAuthClient(config, tokenManager)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
	assert.Equal(t, tokenManager, client.TokenManager())
}

func TestNewOAuthClientFromToken(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	token := &oauth.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	client, err := oauth.NewOAuthClientFromToken(config, token)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
	assert.NotNil(t, client.TokenManager())
}

func TestNewOAuthClientFromStorage(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()

	client, err := oauth.NewOAuthClientFromStorage(config, storage)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
	assert.NotNil(t, client.TokenManager())
}

func TestOAuthClient_ServiceGetters(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()
	tokenManager := oauth.NewTokenManager(config, storage)
	client := oauth.NewOAuthClient(config, tokenManager)

	// Test all service getters
	assert.NotNil(t, client.User())
	assert.NotNil(t, client.Budget())
	assert.NotNil(t, client.Account())
	assert.NotNil(t, client.Category())
	assert.NotNil(t, client.Month())
	assert.NotNil(t, client.Payee())
	assert.NotNil(t, client.Transaction())
}

func TestOAuthClient_RateLimiting(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()
	tokenManager := oauth.NewTokenManager(config, storage)
	client := oauth.NewOAuthClient(config, tokenManager)

	// Test rate limiting methods return expected default values
	assert.Equal(t, 200, client.RequestsRemaining())
	assert.False(t, client.IsAtLimit())
	assert.Equal(t, 0, client.RequestsInWindow())
	assert.Equal(t, time.Duration(0), client.TimeUntilReset())
}

func TestNewTokenManager(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()

	tm := oauth.NewTokenManager(config, storage)

	assert.NotNil(t, tm)
}

func TestClientBuilder(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})

	builder := oauth.NewClientBuilder(config)

	assert.NotNil(t, builder)
}

func TestClientBuilder_WithMethods(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})
	storage := oauth.NewMemoryStorage()
	token := &oauth.Token{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	builder := oauth.NewClientBuilder(config).
		WithStorage(storage).
		WithToken(token).
		WithMemoryStorage()

	assert.NotNil(t, builder)

	// Build the client to verify it works
	client, err := builder.Build()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
}

func TestClientBuilder_WithFileStorage(t *testing.T) {
	config := oauth.NewOAuthConfig(oauth.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "https://example.com/callback",
	})

	builder := oauth.NewClientBuilder(config).
		WithFileStorage("test-token.json")

	assert.NotNil(t, builder)

	// Build the client to verify it works
	client, err := builder.Build()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
}
