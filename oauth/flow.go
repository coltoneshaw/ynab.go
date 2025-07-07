package oauth

import (
	"context"
	"fmt"
	"net/http"
)

// Flow represents an OAuth flow implementation
type Flow interface {
	// GetAuthorizationURL returns the URL users should visit to authorize the application
	GetAuthorizationURL(state string) (string, error)

	// HandleCallback processes the callback from the authorization server
	HandleCallback(callbackURL string, expectedState string) (*Token, error)
}

// AuthorizationCodeFlow implements the OAuth 2.0 Authorization Code Grant flow
type AuthorizationCodeFlow struct {
	config       *Config
	tokenManager *TokenManager
}

// NewAuthorizationCodeFlow creates a new authorization code flow
func NewAuthorizationCodeFlow(config *Config) *AuthorizationCodeFlow {
	return &AuthorizationCodeFlow{
		config:       config,
		tokenManager: NewTokenManager(config, nil),
	}
}

// WithTokenManager sets a custom token manager
func (f *AuthorizationCodeFlow) WithTokenManager(manager *TokenManager) *AuthorizationCodeFlow {
	f.tokenManager = manager
	return f
}

// WithHTTPClient sets a custom HTTP client for token requests
func (f *AuthorizationCodeFlow) WithHTTPClient(client *http.Client) *AuthorizationCodeFlow {
	f.tokenManager.WithHTTPClient(client)
	return f
}

// GetAuthorizationURL returns the authorization URL for the authorization code flow
func (f *AuthorizationCodeFlow) GetAuthorizationURL(state string) (string, error) {
	if err := f.config.Validate(); err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	return f.config.AuthCodeURL(state), nil
}

// HandleCallback processes the authorization callback and exchanges the code for tokens
func (f *AuthorizationCodeFlow) HandleCallback(callbackURL string, expectedState string) (*Token, error) {
	// Parse the callback URL
	result, err := f.config.ParseCallbackURL(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback URL: %w", err)
	}

	// Check for OAuth errors
	if result.Error != nil {
		return nil, result.Error
	}

	// Validate state parameter if provided
	if expectedState != "" && !f.config.ValidateState(expectedState, result.State) {
		return nil, fmt.Errorf("state parameter mismatch")
	}

	// Ensure we have an authorization code
	if result.Code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}

	// Exchange the code for tokens
	ctx := context.Background()
	token, err := f.tokenManager.ExchangeCode(ctx, result.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// HandleCallbackWithContext processes the authorization callback with context
func (f *AuthorizationCodeFlow) HandleCallbackWithContext(ctx context.Context, callbackURL string, expectedState string) (*Token, error) {
	// Parse the callback URL
	result, err := f.config.ParseCallbackURL(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback URL: %w", err)
	}

	// Check for OAuth errors
	if result.Error != nil {
		return nil, result.Error
	}

	// Validate state parameter if provided
	if expectedState != "" && !f.config.ValidateState(expectedState, result.State) {
		return nil, fmt.Errorf("state parameter mismatch")
	}

	// Ensure we have an authorization code
	if result.Code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}

	// Exchange the code for tokens
	token, err := f.tokenManager.ExchangeCode(ctx, result.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// ImplicitGrantFlow implements the OAuth 2.0 Implicit Grant flow
type ImplicitGrantFlow struct {
	config *Config
}

// NewImplicitGrantFlow creates a new implicit grant flow
func NewImplicitGrantFlow(config *Config) *ImplicitGrantFlow {
	return &ImplicitGrantFlow{
		config: config,
	}
}

// GetAuthorizationURL returns the authorization URL for the implicit grant flow
func (f *ImplicitGrantFlow) GetAuthorizationURL(state string) (string, error) {
	if err := f.config.Validate(); err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	return f.config.ImplicitGrantURL(state), nil
}

// HandleCallback processes the authorization callback and extracts the access token
func (f *ImplicitGrantFlow) HandleCallback(callbackURL string, expectedState string) (*Token, error) {
	// Parse the callback URL
	result, err := f.config.ParseCallbackURL(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback URL: %w", err)
	}

	// Check for OAuth errors
	if result.Error != nil {
		return nil, result.Error
	}

	// Validate state parameter if provided
	if expectedState != "" && !f.config.ValidateState(expectedState, result.State) {
		return nil, fmt.Errorf("state parameter mismatch")
	}

	// Convert to token
	token := result.ToToken()
	if token == nil {
		return nil, fmt.Errorf("no access token received")
	}

	return token, nil
}

// FlowManager manages multiple OAuth flows
type FlowManager struct {
	authCodeFlow   *AuthorizationCodeFlow
	implicitFlow   *ImplicitGrantFlow
	config         *Config
	defaultStorage TokenStorage
}

// NewFlowManager creates a new flow manager
func NewFlowManager(config *Config) *FlowManager {
	return &FlowManager{
		config:       config,
		authCodeFlow: NewAuthorizationCodeFlow(config),
		implicitFlow: NewImplicitGrantFlow(config),
	}
}

// WithDefaultStorage sets the default token storage for flows
func (fm *FlowManager) WithDefaultStorage(storage TokenStorage) *FlowManager {
	fm.defaultStorage = storage
	if fm.authCodeFlow != nil && fm.authCodeFlow.tokenManager != nil {
		fm.authCodeFlow.tokenManager.storage = storage
	}
	return fm
}

// WithHTTPClient sets a custom HTTP client for token operations
func (fm *FlowManager) WithHTTPClient(client *http.Client) *FlowManager {
	if fm.authCodeFlow != nil {
		fm.authCodeFlow.WithHTTPClient(client)
	}
	return fm
}

// AuthorizationCode returns the authorization code flow
func (fm *FlowManager) AuthorizationCode() *AuthorizationCodeFlow {
	return fm.authCodeFlow
}

// ImplicitGrant returns the implicit grant flow
func (fm *FlowManager) ImplicitGrant() *ImplicitGrantFlow {
	return fm.implicitFlow
}

// GetFlow returns the appropriate flow based on the response type
func (fm *FlowManager) GetFlow(responseType ResponseType) Flow {
	switch responseType {
	case ResponseTypeCode:
		return fm.authCodeFlow
	case ResponseTypeToken:
		return fm.implicitFlow
	default:
		return fm.authCodeFlow // Default to authorization code flow
	}
}

// StartAuthorizationCodeFlow is a helper method to start the authorization code flow
func (fm *FlowManager) StartAuthorizationCodeFlow() (authURL, state string, err error) {
	state, err = fm.config.GenerateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL, err = fm.authCodeFlow.GetAuthorizationURL(state)
	if err != nil {
		return "", "", fmt.Errorf("failed to get authorization URL: %w", err)
	}

	return authURL, state, nil
}

// StartImplicitGrantFlow is a helper method to start the implicit grant flow
func (fm *FlowManager) StartImplicitGrantFlow() (authURL, state string, err error) {
	state, err = fm.config.GenerateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL, err = fm.implicitFlow.GetAuthorizationURL(state)
	if err != nil {
		return "", "", fmt.Errorf("failed to get authorization URL: %w", err)
	}

	return authURL, state, nil
}

// CompleteAuthorizationCodeFlow completes the authorization code flow
func (fm *FlowManager) CompleteAuthorizationCodeFlow(ctx context.Context, callbackURL, expectedState string) (*Token, error) {
	return fm.authCodeFlow.HandleCallbackWithContext(ctx, callbackURL, expectedState)
}

// CompleteImplicitGrantFlow completes the implicit grant flow
func (fm *FlowManager) CompleteImplicitGrantFlow(callbackURL, expectedState string) (*Token, error) {
	return fm.implicitFlow.HandleCallback(callbackURL, expectedState)
}

// RecommendFlow recommends the best OAuth flow based on application type
func RecommendFlow(isServerSide, needsRefreshToken bool) ResponseType {
	if isServerSide && needsRefreshToken {
		return ResponseTypeCode // Authorization Code flow
	}

	if !isServerSide {
		return ResponseTypeToken // Implicit flow for client-side apps
	}

	return ResponseTypeCode // Default to Authorization Code flow
}
