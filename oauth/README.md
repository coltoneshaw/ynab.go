# YNAB Go OAuth Package

This package provides comprehensive OAuth 2.0 support for the YNAB Go library, enabling secure authentication and automatic token management.

## Quick Start

### Authorization Code Flow (Recommended for Server Applications)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/coltoneshaw/ynab.go"
    "github.com/coltoneshaw/ynab.go/oauth"
)

func main() {
    // 1. Create OAuth configuration
    config := ynab.NewOAuthConfig(
        "your-client-id",
        "your-client-secret",
        "https://yourapp.com/oauth/callback",
    ).WithReadOnlyScope()

    // 2. Start OAuth flow
    flowManager := ynab.NewFlowManager(config)
    authURL, state, err := flowManager.StartAuthorizationCodeFlow()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Visit: %s\n", authURL)

    // 3. Handle callback (after user authorization)
    callbackURL := "https://yourapp.com/oauth/callback?code=received-code&state=" + state
    
    ctx := context.Background()
    token, err := flowManager.CompleteAuthorizationCodeFlow(ctx, callbackURL, state)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Create OAuth client
    client, err := ynab.NewOAuthClientFromToken(config, token)
    if err != nil {
        log.Fatal(err)
    }

    // 5. Use the client
    user, err := client.User().GetUser()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Authenticated as: %s\n", user.ID)
}
```

### Implicit Grant Flow (For Client-Side Applications)

```go
// For JavaScript/mobile apps where client secret cannot be secured
config := ynab.NewOAuthConfig(
    "your-client-id",
    "", // No client secret for implicit grant
    "https://yourapp.com/oauth/callback",
)

flow := ynab.NewImplicitGrantFlow(config)
authURL, err := flow.GetAuthorizationURL("state-123")
if err != nil {
    log.Fatal(err)
}

// User visits authURL, gets redirected with token in URL fragment
callbackURL := "https://yourapp.com/oauth/callback#access_token=token&token_type=Bearer"

token, err := flow.HandleCallback(callbackURL, "state-123")
if err != nil {
    log.Fatal(err)
}

client, err := ynab.NewOAuthClientFromToken(config, token)
// ... use client
```

## Features

### 🔐 OAuth 2.0 Flows
- **Authorization Code Grant** - Secure server-side authentication with refresh tokens
- **Implicit Grant** - Client-side authentication for SPAs and mobile apps
- **State parameter** - CSRF protection with automatic state generation
- **Scope support** - Request specific permissions (read-only, full access)

### 🔄 Token Management
- **Automatic refresh** - Transparent token renewal when expired
- **Multiple storage options** - Memory, file-based, encrypted, or custom storage
- **Thread-safe operations** - Concurrent access to tokens
- **Expiration handling** - Built-in token validity checking

### 🏗️ Flexible Architecture
- **Builder pattern** - Fluent API for client configuration
- **Interface-based design** - Easy to mock and test
- **Context support** - Proper cancellation and timeout handling
- **HTTP client customization** - Use your own HTTP transport

### 🛡️ Security Features
- **Secure token storage** - File permissions and optional encryption
- **CSRF protection** - State parameter validation
- **Token refresh callbacks** - Get notified when tokens are renewed
- **Error handling** - Proper OAuth error response handling

## Components

### Configuration

```go
config := oauth.NewOAuthConfig(oauth.Config{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    RedirectURI:  "redirect-uri",
})

// Add scopes
config.WithReadOnlyScope()
config.WithScope(oauth.ScopeDefault)

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

### Token Storage

#### File Storage (Recommended)
```go
// Default location (~/.config/ynab/token.json)
storage := ynab.NewFileStorage(ynab.DefaultTokenPath())

// Custom location
storage := ynab.NewFileStorage("/path/to/token.json")

// Custom permissions
storage := oauth.NewFileStorage("token.json").WithFileMode(0600)
```

#### Memory Storage
```go
storage := ynab.NewMemoryStorage() // Not persistent
```

#### Encrypted Storage
```go
key := []byte("your-encryption-key")
storage := oauth.NewEncryptedFileStorage("token.json", key)
```

#### Chained Storage (Fallback)
```go
primary := oauth.NewFileStorage("primary.json")
backup := oauth.NewMemoryStorage()
storage := oauth.NewChainedStorage(primary, backup)
```

### Client Builder

```go
client, err := ynab.NewOAuthClientBuilder(config).
    WithDefaultFileStorage().
    WithTokenRefreshCallback(func(token *oauth.Token) {
        log.Println("Token refreshed")
    }).
    WithHTTPClient(&http.Client{Timeout: 30 * time.Second}).
    Build()
```

### Token Management

```go
tokenManager := ynab.NewTokenManager(config, storage)

// Get token (refreshes automatically if needed)
ctx := context.Background()
token, err := tokenManager.GetToken(ctx)

// Manual refresh
newToken, err := tokenManager.RefreshToken(ctx)

// Check authentication status
if tokenManager.IsAuthenticated() {
    // Token is valid
}

// Clear stored token
tokenManager.ClearToken()
```

## OAuth Flows

### Authorization Code Flow

Best for server-side applications where the client secret can be kept secure.

**Pros:**
- Most secure OAuth flow
- Supports refresh tokens
- Long-lived authentication

**Cons:**
- Requires server-side implementation
- More complex setup

```go
flow := oauth.NewAuthorizationCodeFlow(config)

// Step 1: Get authorization URL
authURL, err := flow.GetAuthorizationURL("state-parameter")

// Step 2: User visits URL and authorizes

// Step 3: Handle callback
token, err := flow.HandleCallback(callbackURL, "state-parameter")
```

### Implicit Grant Flow

Best for client-side applications (SPAs, mobile apps) where client secret cannot be secured.

**Pros:**
- Simple client-side implementation
- No server-side component needed
- Good for SPAs and mobile apps

**Cons:**
- Less secure than authorization code flow
- No refresh tokens
- Shorter token lifetime

```go
flow := oauth.NewImplicitGrantFlow(config)

// Step 1: Get authorization URL
authURL, err := flow.GetAuthorizationURL("state-parameter")

// Step 2: User visits URL, token returned in URL fragment

// Step 3: Handle callback
token, err := flow.HandleCallback(callbackURL, "state-parameter")
```

### Flow Selection

Use the built-in recommendation system:

```go
// Recommend flow based on app architecture
isServerSide := true
needsRefreshToken := true

recommendedFlow := oauth.RecommendFlow(isServerSide, needsRefreshToken)
// Returns: oauth.ResponseTypeCode for authorization code flow
```

## Error Handling

The package provides detailed error information for OAuth failures:

```go
token, err := flow.HandleCallback(callbackURL, state)
if err != nil {
    if oauthErr, ok := err.(*oauth.ErrorResponse); ok {
        switch oauthErr.ErrorCode {
        case "access_denied":
            // User denied authorization
        case "invalid_request":
            // Invalid OAuth request
        default:
            // Other OAuth error
        }
    } else {
        // Network or parsing error
    }
}
```

Common error scenarios:
- `access_denied` - User denied authorization
- `invalid_request` - Malformed OAuth request
- `invalid_client` - Invalid client credentials
- `invalid_grant` - Invalid authorization code
- `unauthorized_client` - Client not authorized for this grant type

## Integration Examples

### Web Server Integration

```go
func startOAuth(w http.ResponseWriter, r *http.Request) {
    state, _ := config.GenerateState()
    
    // Store state in session for validation
    session.Set("oauth_state", state)
    
    authURL, _ := flow.GetAuthorizationURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
    expectedState := session.Get("oauth_state")
    
    token, err := flow.HandleCallback(r.URL.String(), expectedState)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Store token for user
    userTokens[userID] = token
    
    // Create client and use YNAB API
    client, _ := ynab.NewOAuthClientFromToken(config, token)
    user, _ := client.User().GetUser()
    
    fmt.Fprintf(w, "Authenticated as: %s", user.ID)
}
```

### Command Line Application

```go
func authenticateUser() *oauth.OAuthClient {
    config := oauth.NewOAuthConfig(oauth.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURI:  "http://localhost:8080/callback",
    })
    
    // Start local server for callback
    server := &http.Server{Addr: ":8080"}
    var token *oauth.Token
    
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        var err error
        token, err = flow.HandleCallback(r.URL.String(), state)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        
        fmt.Fprintln(w, "Authentication successful! You can close this window.")
        go server.Shutdown(context.Background())
    })
    
    go server.ListenAndServe()
    
    // Open browser
    authURL, state, _ := flowManager.StartAuthorizationCodeFlow()
    exec.Command("open", authURL).Start() // macOS
    
    // Wait for callback
    <-server.Context().Done()
    
    client, _ := ynab.NewOAuthClientFromToken(config, token)
    return client
}
```

### Mobile App Integration

```go
// Use implicit grant for mobile apps
config := oauth.NewOAuthConfig(oauth.Config{
    ClientID:     clientID,
    ClientSecret: "", // No secret for mobile apps
    RedirectURI:  "yourapp://oauth/callback",
})
flow := oauth.NewImplicitGrantFlow(config)

// Open authorization URL in system browser
authURL, _ := flow.GetAuthorizationURL("mobile-state")
openURL(authURL) // Platform-specific browser opening

// Handle custom URL scheme callback
func handleCustomURL(url string) {
    token, err := flow.HandleCallback(url, "mobile-state")
    if err != nil {
        // Handle error
        return
    }
    
    // Create client with memory storage for mobile
    client, _ := ynab.NewOAuthClientBuilder(config).
        WithMemoryStorage().
        WithToken(token).
        Build()
    
    // Use client
    user, _ := client.User().GetUser()
}
```

## Advanced Features

### Custom HTTP Client

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
    },
}

client := ynab.NewOAuthClientBuilder(config).
    WithHTTPClient(httpClient).
    Build()
```

### Token Refresh Callbacks

```go
tokenManager := oauth.NewTokenManager(config, storage).
    WithTokenRefreshCallback(func(token *oauth.Token) {
        log.Printf("Token refreshed, expires at: %v", token.ExpiresAt)
        
        // Optional: notify user, update UI, etc.
        notifyUser("Authentication renewed")
        
        // Optional: log for monitoring
        metrics.Counter("token_refresh").Inc()
    })
```

### Custom Storage Implementation

```go
type DatabaseStorage struct {
    userID string
    db     *sql.DB
}

func (s *DatabaseStorage) SaveToken(token *oauth.Token) error {
    data, _ := json.Marshal(token)
    _, err := s.db.Exec("UPDATE users SET oauth_token = ? WHERE id = ?", data, s.userID)
    return err
}

func (s *DatabaseStorage) LoadToken() (*oauth.Token, error) {
    var data []byte
    err := s.db.QueryRow("SELECT oauth_token FROM users WHERE id = ?", s.userID).Scan(&data)
    if err != nil {
        return nil, err
    }
    
    var token oauth.Token
    err = json.Unmarshal(data, &token)
    return &token, err
}

func (s *DatabaseStorage) ClearToken() error {
    _, err := s.db.Exec("UPDATE users SET oauth_token = NULL WHERE id = ?", s.userID)
    return err
}

func (s *DatabaseStorage) HasToken() bool {
    var count int
    s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ? AND oauth_token IS NOT NULL", s.userID).Scan(&count)
    return count > 0
}
```

## Security Considerations

### Production Best Practices

1. **Always use HTTPS** for redirect URIs in production
2. **Validate state parameters** to prevent CSRF attacks
3. **Store client secrets securely** (environment variables, secret management)
4. **Use secure token storage** with appropriate file permissions
5. **Implement proper error handling** without exposing sensitive information
6. **Monitor token refresh patterns** for unusual activity
7. **Set appropriate token lifetimes** based on your security requirements

### Token Storage Security

```go
// Secure file storage with restricted permissions
storage := oauth.NewFileStorage("tokens.json").WithFileMode(0600)

// Encrypted storage for sensitive environments
key := generateSecureKey() // Use proper key derivation
storage := oauth.NewEncryptedFileStorage("tokens.json", key)

// Environment-specific configuration
if isProduction() {
    storage = oauth.NewEncryptedFileStorage(getSecureTokenPath(), getEncryptionKey())
} else {
    storage = oauth.NewFileStorage("dev-tokens.json")
}
```

### CSRF Protection

```go
// Always validate state parameters
state, err := config.GenerateState()
if err != nil {
    log.Fatal("Failed to generate secure state")
}

// Store state securely (session, database, etc.)
session.Set("oauth_state", state)

// Validate on callback
expectedState := session.Get("oauth_state")
token, err := flow.HandleCallback(callbackURL, expectedState)
if err != nil {
    // Handle validation failure
}
```

## Testing

The package provides comprehensive test coverage and examples for testing OAuth flows:

```go
func TestOAuthFlow(t *testing.T) {
    // Mock HTTP responses
    httpmock.Activate()
    defer httpmock.DeactivateAndReset()
    
    httpmock.RegisterResponder("POST", oauth.TokenURL,
        httpmock.NewStringResponder(200, `{
            "access_token": "test-token",
            "token_type": "Bearer",
            "expires_in": 7200
        }`))
    
    config := oauth.NewOAuthConfig(oauth.Config{
        ClientID:     "test-client",
        ClientSecret: "test-secret",
        RedirectURI:  "http://test.com/callback",
    })
    flow := oauth.NewAuthorizationCodeFlow(config)
    
    // Test authorization URL generation
    authURL, err := flow.GetAuthorizationURL("test-state")
    assert.NoError(t, err)
    assert.Contains(t, authURL, "client_id=test-client")
    
    // Test callback handling
    callbackURL := "http://test.com/callback?code=test-code&state=test-state"
    token, err := flow.HandleCallback(callbackURL, "test-state")
    assert.NoError(t, err)
    assert.Equal(t, "test-token", token.AccessToken)
}
```

## Performance Considerations

### Token Caching

Tokens are automatically cached in memory for the duration of the application. For multi-instance deployments, consider shared storage:

```go
// Use database storage for shared token access
storage := &DatabaseStorage{userID: userID, db: db}
tokenManager := oauth.NewTokenManager(config, storage)
```

### HTTP Client Reuse

The package reuses HTTP clients by default. For high-throughput applications, consider connection pooling:

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

httpClient := &http.Client{Transport: transport}
client := ynab.NewOAuthClientBuilder(config).
    WithHTTPClient(httpClient).
    Build()
```

### Monitoring

Monitor OAuth operations for performance and security:

```go
tokenManager := oauth.NewTokenManager(config, storage).
    WithTokenRefreshCallback(func(token *oauth.Token) {
        // Monitor refresh frequency
        metrics.Counter("oauth_token_refresh").Inc()
        
        // Alert on excessive refresh rates
        if metrics.GetRate("oauth_token_refresh") > threshold {
            alerting.Send("High OAuth refresh rate detected")
        }
    })
```

## Troubleshooting

### Common Issues

**1. "invalid_client" error**
- Check client ID and secret
- Ensure client is registered with YNAB
- Verify redirect URI matches registration

**2. "access_denied" error**
- User denied authorization
- Check OAuth scope requirements
- Verify application permissions

**3. "invalid_grant" error**
- Authorization code expired or already used
- Code doesn't match redirect URI
- Check system clock synchronization

**4. Token refresh failures**
- Refresh token expired
- Invalid client credentials
- Network connectivity issues

### Debug Mode

Enable detailed logging for OAuth operations:

```go
import "log"

config := oauth.NewConfig(clientID, clientSecret, redirectURI)
client := ynab.NewOAuthClientBuilder(config).
    WithTokenRefreshCallback(func(token *oauth.Token) {
        log.Printf("DEBUG: Token refreshed - expires: %v", token.ExpiresAt)
    }).
    Build()
```

### Environment Variables

Configure OAuth settings via environment:

```go
config := oauth.NewOAuthConfig(oauth.Config{
    ClientID:     os.Getenv("YNAB_CLIENT_ID"),
    ClientSecret: os.Getenv("YNAB_CLIENT_SECRET"),
    RedirectURI:  os.Getenv("YNAB_REDIRECT_URI"),
})

if os.Getenv("YNAB_DEBUG") == "true" {
    // Enable debug logging
}
```

## Migration Guide

### From Token-Based Authentication

If you're currently using the library with static access tokens:

```go
// Old approach
client := ynab.NewClient("static-access-token")

// New OAuth approach
config := oauth.NewOAuthConfig(oauth.Config{
    ClientID:     clientID,
    ClientSecret: clientSecret,
    RedirectURI:  redirectURI,
})
client, err := ynab.NewOAuthClientFromStorage(config, storage)
```

### Backwards Compatibility

The OAuth package is fully compatible with existing code. You can migrate gradually:

```go
// Existing token-based client still works
legacyClient := ynab.NewClient("access-token")

// New OAuth client with same API
oauthClient, _ := ynab.NewOAuthClientFromToken(config, token)

// Both implement the same interfaces
func useClient(client ynab.ClientServicer) {
    user, _ := client.User().GetUser()
    // ... same API for both clients
}
```

## Contributing

Contributions are welcome! Please ensure:

1. **Add tests** for new functionality
2. **Update documentation** for API changes
3. **Follow Go conventions** for naming and structure
4. **Include examples** for new features
5. **Maintain backwards compatibility** when possible

## License

This package is licensed under the same BSD-2-Clause license as the main YNAB Go library.