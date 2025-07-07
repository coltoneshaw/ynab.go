# YNAB API Go Library

[![Go Report Card](https://goreportcard.com/badge/github.com/brunomvsouza/ynab.go)](https://goreportcard.com/report/github.com/brunomvsouza/ynab.go) [![GoDoc Reference](https://godoc.org/github.com/brunomvsouza/ynab.go?status.svg)](https://godoc.org/github.com/brunomvsouza/ynab.go)

This is an UNOFFICIAL Go client for the YNAB API. It covers 100% of the resources made available by the [YNAB API](https://api.youneedabudget.com).

## 📚 Quick Navigation

- [🚀 **Quick Start**](#quick-start) - Get up and running fast
- [🔐 **Authentication**](#authentication-methods) - OAuth 2.0 and Personal Access Tokens  
- [📋 **Usage Examples**](#usage-examples) - Working with budgets, accounts, transactions
- [🚨 **Error Handling**](#error-handling) - Type-safe error constants and patterns
- [⚙️ **Advanced Usage**](#advanced-usage) - Custom HTTP clients, production patterns
- [📊 **Rate Limiting**](#rate-limiting) - Built-in rate tracking and management

## Features

- ✅ **Complete API Coverage** - All 36 YNAB API endpoints implemented
- 🔐 **OAuth 2.0 Support** - Authorization Code and Implicit Grant flows
- 🔄 **Automatic Token Refresh** - Seamless token renewal
- 💾 **Flexible Token Storage** - Memory, file, encrypted, or custom storage
- 🧪 **Comprehensive Testing** - 90%+ test coverage
- 🛡️ **Security Features** - CSRF protection, secure token handling
- 🏗️ **Clean Architecture** - Interface-based design for easy testing
- 🚨 **Enhanced Error Handling** - Type-safe error constants and helper methods for all YNAB API errors

## Installation

```bash
go get github.com/brunomvsouza/ynab.go
```

## Quick Start

### Option 1: OAuth 2.0 Authentication (Recommended)

OAuth provides the most secure and user-friendly authentication method, with automatic token refresh and proper scope management.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/brunomvsouza/ynab.go"
)

func main() {
    // 1. Create OAuth configuration
    config := ynab.NewOAuthConfig(
        "your-client-id",
        "your-client-secret",
        "https://yourapp.com/oauth/callback",
    ).WithReadOnlyScope() // Optional: limits to read-only access

    // 2. Start OAuth flow
    flowManager := ynab.NewFlowManager(config).
        WithDefaultStorage(ynab.NewFileStorage(ynab.DefaultTokenPath()))

    authURL, state, err := flowManager.StartAuthorizationCodeFlow()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Visit this URL to authorize: %s\n", authURL)

    // 3. Handle callback (after user authorizes)
    // In a real app, this would be your HTTP callback handler
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

    // 5. Use the client - tokens refresh automatically!
    user, err := client.User().GetUser()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Authenticated as: %s\n", user.ID)

    // Get budgets
    budgets, err := client.Budget().GetBudgets()
    if err != nil {
        log.Fatal(err)
    }

    for _, budget := range budgets {
        fmt.Printf("Budget: %s\n", budget.Name)
    }
}
```

### Option 2: Personal Access Token (Simple)

For personal scripts or development, you can use a static access token from your [YNAB account settings](https://app.youneedabudget.com/settings).

```go
package main

import (
    "fmt"
    "log"

    "github.com/brunomvsouza/ynab.go"
)

func main() {
    // Get your personal access token from https://app.youneedabudget.com/settings
    const accessToken = "your-personal-access-token"

    client := ynab.NewClient(accessToken)
    budgets, err := client.Budget().GetBudgets()
    if err != nil {
        log.Fatal(err)
    }

    for _, budget := range budgets {
        fmt.Printf("Budget: %s\n", budget.Name)
    }
}
```

## Authentication Methods

### OAuth 2.0 (Recommended for Production Apps)

OAuth 2.0 is the recommended authentication method for production applications. It provides:

- **Secure token management** with automatic refresh
- **User consent** - users explicitly authorize your application
- **Scope control** - request only the permissions you need
- **Better UX** - no need for users to generate personal tokens

#### OAuth Scopes

YNAB supports the following OAuth scopes:

| Scope | Permissions | Usage |
|-------|-------------|-------|
| **Default (no scope)** | Full read-write access | Can read and modify all YNAB data |
| **`read-only`** | Read-only access | Can only read data, cannot create/update/delete |

```go
// Full access (default) - can read and modify data
config := ynab.NewOAuthConfig("client-id", "client-secret", "redirect-uri")

// Read-only access - cannot modify data
config := ynab.NewOAuthConfig("client-id", "client-secret", "redirect-uri").
    WithReadOnlyScope()

// Check current scope configuration
if config.IsReadOnly() {
    fmt.Println("Client has read-only access")
}
```

**Important:** When using `read-only` scope, any attempt to call modification endpoints (POST, PATCH, PUT, DELETE) will result in a `403 Forbidden` error from the YNAB API.

#### Authorization Code Flow (Server-Side Apps)

Best for web applications where you can securely store client secrets:

```go
// Step 1: Setup OAuth configuration
config := ynab.NewOAuthConfig(
    "your-client-id",           // From YNAB OAuth app registration
    "your-client-secret",       // Keep this secure!
    "https://yourapp.com/callback",
).WithReadOnlyScope() // or WithScope() for custom scopes

// Step 2: Create client with persistent storage
client, err := ynab.NewOAuthClientBuilder(config).
    WithDefaultFileStorage().                    // Saves tokens to ~/.config/ynab/token.json
    WithTokenRefreshCallback(func(token *oauth.Token) {
        log.Println("Token refreshed automatically")
    }).
    Build()

// Step 3: Start OAuth flow (typically in HTTP handler)
flow := ynab.NewAuthorizationCodeFlow(config)
authURL, err := flow.GetAuthorizationURL("secure-state-parameter")
// Redirect user to authURL...

// Step 4: Handle callback (in your callback HTTP handler)
token, err := flow.HandleCallback(callbackURL, "secure-state-parameter")
if err != nil {
    // Handle OAuth errors (user denied, invalid request, etc.)
}

// Step 5: Set token and use client
client.SetToken(token)
user, err := client.User().GetUser() // Automatic token refresh if needed!
```

#### Implicit Grant Flow (Client-Side Apps)

Best for SPAs, mobile apps, or anywhere you can't securely store client secrets:

```go
// No client secret needed for implicit grant
config := ynab.NewOAuthConfig(
    "your-client-id",
    "", // No secret for client-side apps
    "https://yourapp.com/callback",
)

flow := ynab.NewImplicitGrantFlow(config)
authURL, err := flow.GetAuthorizationURL("state-123")

// User visits authURL, token returned in URL fragment
// Example: https://yourapp.com/callback#access_token=abc123&token_type=Bearer&expires_in=7200

token, err := flow.HandleCallback(callbackURL, "state-123")
client, err := ynab.NewOAuthClientFromToken(config, token)
```

### Personal Access Tokens (Simple)

For personal scripts, development, or simple integrations:

```go
// Get token from https://app.youneedabudget.com/settings
client := ynab.NewClient("your-personal-access-token")
```

## Usage Examples

### Working with Budgets

```go
// List all budgets
budgets, err := client.Budget().GetBudgets()
if err != nil {
    log.Fatal(err)
}

// Get detailed budget with all data
budget, err := client.Budget().GetBudget("budget-id", nil)
if err != nil {
    log.Fatal(err)
}

// Get budget settings
settings, err := client.Budget().GetBudgetSettings("budget-id")
```

### Working with Accounts

```go
// Get all accounts
accounts, err := client.Account().GetAccounts("budget-id", nil)

// Get specific account
account, err := client.Account().GetAccount("budget-id", "account-id")

// Create new account
newAccount := account.PayloadAccount{
    Name: "Emergency Fund",
    Type: account.TypeSavings,
}
createdAccount, err := client.Account().CreateAccount("budget-id", newAccount)
```

### Working with Transactions

```go
// Get all transactions
transactions, err := client.Transaction().GetTransactions("budget-id", nil)

// Get transactions with filtering
filter := &api.Filter{
    SinceDate: api.Date(time.Now().AddDate(0, -1, 0)), // Last month
}
transactions, err := client.Transaction().GetTransactions("budget-id", filter)

// Create new transaction
newTransaction := transaction.PayloadTransaction{
    AccountID:  "account-id",
    CategoryID: "category-id",
    PayeeID:    "payee-id",
    Amount:     -25000, // $250.00 (milliunits)
    Memo:       "Coffee shop",
}
created, err := client.Transaction().CreateTransaction("budget-id", newTransaction)

// Update transaction
updated, err := client.Transaction().UpdateTransaction("budget-id", "transaction-id", transaction.PayloadTransaction{
    Amount: -30000, // $300.00
})

// Delete transaction
err = client.Transaction().DeleteTransaction("budget-id", "transaction-id")
```

### Working with Categories

```go
// Get all categories
categories, err := client.Category().GetCategories("budget-id", nil)

// Update category budget
err = client.Category().UpdateCategoryForCurrentMonth("budget-id", "category-id", category.PayloadMonthCategory{
    Budgeted: 50000, // $500.00
})
```

## Advanced Usage

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    // Add custom transport, proxy, etc.
}

// For OAuth clients
client := ynab.NewOAuthClientBuilder(config).
    WithHTTPClient(httpClient).
    Build()

// For token-based clients
client := ynab.NewClient("token")
// Note: Custom HTTP clients for token-based clients require modifying the client after creation
```

## Error Handling

The library provides **enhanced error handling** with type-safe constants and helper methods for all documented YNAB API errors. This makes it easy to build robust applications that can gracefully handle different error scenarios.

### Basic Error Handling

```go
import "github.com/brunomvsouza/ynab.go/api"

user, err := client.User().GetUser()
if err != nil {
    if apiErr, ok := err.(*api.Error); ok {
        switch {
        case apiErr.IsAccountError():
            // Subscription lapsed or trial expired
            log.Println("Account issue - redirect to billing")
            redirectToBilling()
        case apiErr.IsAuthenticationError():
            // Invalid token or insufficient permissions
            log.Println("Authentication failed - redirect to login")
            redirectToLogin()
        case apiErr.IsRateLimit():
            // Too many requests
            log.Println("Rate limited - waiting before retry")
            time.Sleep(client.TimeUntilReset())
            // Retry request...
        case apiErr.IsRetryable():
            // Server errors that might resolve
            log.Println("Server error - retrying with backoff")
            // Implement retry logic...
        default:
            log.Printf("API error: %s", apiErr.Detail)
        }
    } else {
        log.Printf("Network error: %v", err)
    }
}
```

### All Available Error Constants

All YNAB API error codes are available as type-safe constants:

```go
import "github.com/brunomvsouza/ynab.go/api"

// 4xx Client Errors
api.ErrorBadRequest         // "400" - Validation/malformed request
api.ErrorUnauthorized       // "401" - Authentication failure
api.ErrorSubscriptionLapsed // "403.1" - Subscription has lapsed
api.ErrorTrialExpired       // "403.2" - Trial has expired
api.ErrorUnauthorizedScope  // "403.3" - Insufficient permissions
api.ErrorDataLimitReached   // "403.4" - Data limits exceeded
api.ErrorNotFound           // "404.1" - URI not found
api.ErrorResourceNotFound   // "404.2" - Resource not found
api.ErrorConflict           // "409" - Resource conflict
api.ErrorRateLimit          // "429" - Too many requests

// 5xx Server Errors
api.ErrorInternalServer     // "500" - Internal server error
api.ErrorServiceUnavailable // "503" - Service unavailable
```

### Error Categorization Helper Methods

#### Account/Subscription Errors
```go
if apiErr.IsSubscriptionLapsed() {
    // Redirect user to billing page
    redirectToBilling()
} else if apiErr.IsTrialExpired() {
    // Show upgrade prompt
    showUpgradePrompt()
} else if apiErr.IsAccountError() {
    // General account issue (covers both above)
    showAccountNotification(apiErr.Detail)
}
```

#### Authentication/Authorization Errors
```go
if apiErr.IsUnauthorized() {
    // Token is invalid, expired, or missing
    redirectToLogin()
} else if apiErr.IsUnauthorizedScope() {
    // Insufficient permissions for requested operation
    showPermissionError()
} else if apiErr.IsAuthenticationError() {
    // General auth issue (covers both above)
    handleAuthFailure(apiErr)
}
```

#### Resource Errors
```go
if apiErr.IsNotFound() {
    log.Println("Resource not found")
} else if apiErr.IsConflict() {
    // Usually means duplicate import_id for transactions
    log.Println("Resource conflict - may be duplicate")
} else if apiErr.IsDataLimitReached() {
    log.Println("Data limit reached - request too large")
}
```

#### General Error Categories
```go
if apiErr.IsRetryable() {
    // Safe to retry (rate limits, server errors)
    implementRetryLogic()
} else if apiErr.RequiresUserAction() {
    // User needs to do something (billing, auth, etc.)
    showUserNotification(apiErr.Detail)
} else if apiErr.IsClientError() {
    // 4xx errors - client issue
    handleClientError()
} else if apiErr.IsServerError() {
    // 5xx errors - server issue
    handleServerError()
}
```

### Production-Ready Retry Logic

```go
func makeRequestWithRetry(client ynab.ClientServicer, budgetID string) ([]*budget.Budget, error) {
    maxRetries := 3
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        budgets, err := client.Budget().GetBudgets()
        if err == nil {
            return budgets, nil
        }
        
        if apiErr, ok := err.(*api.Error); ok {
            if !apiErr.IsRetryable() {
                // Not retryable - return immediately
                return nil, err
            }
            
            if apiErr.IsRateLimit() {
                // Wait for rate limit reset
                waitTime := client.TimeUntilReset()
                log.Printf("Rate limited, waiting %v", waitTime)
                time.Sleep(waitTime)
                continue
            }
            
            if apiErr.IsServerError() {
                // Exponential backoff for server errors
                waitTime := time.Duration(attempt+1) * time.Second
                log.Printf("Server error, retrying in %v", waitTime)
                time.Sleep(waitTime)
                continue
            }
        }
        
        // Non-API error or non-retryable error
        return nil, err
    }
    
    return nil, fmt.Errorf("max retries exceeded")
}
```

### Complete Error Handling Example

```go
func comprehensiveErrorHandling(client ynab.ClientServicer, budgetID string) error {
    budget, err := client.Budget().GetBudget(budgetID, nil)
    if err != nil {
        if apiErr, ok := err.(*api.Error); ok {
            switch {
            case apiErr.IsAccountError():
                return handleAccountIssue(apiErr)
            case apiErr.IsAuthenticationError():
                return handleAuthIssue(apiErr)
            case apiErr.IsRateLimit():
                return handleRateLimit(client, apiErr)
            case apiErr.IsNotFound():
                return fmt.Errorf("budget %s not found", budgetID)
            case apiErr.IsValidationError():
                return fmt.Errorf("invalid request: %s", apiErr.Detail)
            case apiErr.IsRetryable():
                return fmt.Errorf("temporary server issue: %s", apiErr.Detail)
            default:
                return fmt.Errorf("unexpected API error: %s", apiErr.Detail)
            }
        }
        return fmt.Errorf("request failed: %v", err)
    }
    
    log.Printf("Successfully retrieved budget: %s", budget.Name)
    return nil
}
```


### OAuth Error Handling

```go
token, err := flow.HandleCallback(callbackURL, state)
if err != nil {
    if oauthErr, ok := err.(*oauth.ErrorResponse); ok {
        switch oauthErr.ErrorCode {
        case "access_denied":
            log.Println("User denied authorization")
        case "invalid_request":
            log.Println("Invalid OAuth request")
        default:
            log.Printf("OAuth error: %s", oauthErr.Error())
        }
    }
}
```

### Token Storage Options

```go
// File storage (default location: ~/.config/ynab/token.json)
storage := ynab.NewFileStorage(ynab.DefaultTokenPath())

// Custom file location
storage := ynab.NewFileStorage("/secure/path/ynab-tokens.json")

// Memory storage (not persistent)
storage := ynab.NewMemoryStorage()

// Encrypted storage
key := []byte("your-encryption-key")
storage := oauth.NewEncryptedFileStorage("tokens.json", key)

// Custom storage (implement oauth.TokenStorage interface)
type DatabaseStorage struct { /* your implementation */ }

// Use with OAuth client
client := ynab.NewOAuthClientBuilder(config).
    WithStorage(storage).
    Build()
```

### Production Considerations

#### Web Application Integration

```go
// main.go
func main() {
    config := ynab.NewOAuthConfig(
        os.Getenv("YNAB_CLIENT_ID"),
        os.Getenv("YNAB_CLIENT_SECRET"),
        os.Getenv("YNAB_REDIRECT_URI"),
    )

    http.HandleFunc("/login", handleLogin)
    http.HandleFunc("/oauth/callback", handleCallback)
    http.HandleFunc("/dashboard", handleDashboard)

    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    flow := ynab.NewAuthorizationCodeFlow(config)
    
    state, _ := config.GenerateState()
    // Store state in session for CSRF protection
    session.Set("oauth_state", state)
    
    authURL, _ := flow.GetAuthorizationURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
    expectedState := session.Get("oauth_state")
    
    flow := ynab.NewAuthorizationCodeFlow(config)
    token, err := flow.HandleCallback(r.URL.String(), expectedState)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusBadRequest)
        return
    }
    
    // Store token for user (database, session, etc.)
    userID := getCurrentUserID(r)
    saveUserToken(userID, token)
    
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
    userID := getCurrentUserID(r)
    token := getUserToken(userID)
    
    client, _ := ynab.NewOAuthClientFromToken(config, token)
    budgets, _ := client.Budget().GetBudgets()
    
    // Render dashboard with budgets...
}
```

#### Environment Configuration

```bash
# .env file
YNAB_CLIENT_ID=your_oauth_client_id
YNAB_CLIENT_SECRET=your_oauth_client_secret
YNAB_REDIRECT_URI=https://yourapp.com/oauth/callback

# For development
YNAB_REDIRECT_URI=http://localhost:8080/oauth/callback
```

#### Security Best Practices

1. **Always use HTTPS** in production for redirect URIs
2. **Validate state parameters** to prevent CSRF attacks
3. **Store client secrets securely** (environment variables, secrets management)
4. **Use appropriate token storage** with proper file permissions
5. **Implement proper error handling** without exposing sensitive information
6. **Monitor token refresh patterns** for security anomalies

```go
// Secure token storage with restricted permissions
storage := ynab.NewFileStorage("tokens.json").WithFileMode(0600)

// Production client with monitoring
client := ynab.NewOAuthClientBuilder(config).
    WithStorage(storage).
    WithTokenRefreshCallback(func(token *oauth.Token) {
        // Log token refresh for monitoring
        log.Printf("Token refreshed for user, expires: %v", token.ExpiresAt)
        
        // Optional: Send metrics to monitoring system
        metrics.Counter("oauth_token_refresh").Inc()
    }).
    Build()
```

## API Reference

See the [godoc](https://godoc.org/github.com/brunomvsouza/ynab.go) for complete API documentation with examples.

## Rate Limiting

YNAB enforces **200 requests per hour per access token** using a rolling window. When exceeded, you'll get a `429 Too Many Requests` error.

### Handling 429 Errors

```go
budgets, err := client.Budget().GetBudgets()
if err != nil {
    if apiErr, ok := err.(*api.Error); ok && apiErr.ID == "429" {
        log.Println("Rate limited! Try again later or use delta requests")
    }
}
```

### Automatic Rate Tracking

Rate limiting is now built into all YNAB clients - no manual tracking needed:

```go
// Create client (automatically includes rate limiting)
client := ynab.NewClient("your-token")

// Make API calls - rate limiting is automatic!
budgets, err := client.Budget().GetBudgets()
if err != nil {
    if apiErr, ok := err.(*api.Error); ok && apiErr.ID == "429" {
        // Wait for rate limit to reset
        waitTime := client.TimeUntilReset()
        fmt.Printf("Rate limited! Waiting %v before retry\n", waitTime)
        time.Sleep(waitTime)
        // Retry the request...
    }
}

// Check your current usage anytime
fmt.Printf("Used %d/200 requests, %d remaining\n", 
    client.RequestsInWindow(), client.RequestsRemaining())

// Check if you should wait before making more requests
if client.RequestsRemaining() == 0 {
    waitTime := client.TimeUntilReset()
    fmt.Printf("At rate limit, next request available in %v\n", waitTime)
}
```

#### Planning Batch Operations

Before making many requests, check your remaining quota:

```go
transactions := []transaction.PayloadTransaction{ /* ... */ }

if client.RequestsRemaining() < len(transactions) {
    fmt.Printf("Need %d requests, only %d remaining\n", 
        len(transactions), client.RequestsRemaining())
    
    // Wait for rate limit to reset
    waitTime := client.TimeUntilReset()
    fmt.Printf("Waiting %v for rate limit reset\n", waitTime)
    time.Sleep(waitTime)
}

// Now proceed with batch operation
for _, tx := range transactions {
    _, err := client.Transaction().CreateTransaction(budgetID, tx)
    // Rate limiting is tracked automatically
}
```

### Reduce API Usage

**Use delta requests** to fetch only changes:
```go
// Get full data first
snapshot, _ := client.Budget().GetBudget(budgetID, nil)

// Later, get only changes
filter := &api.Filter{LastKnowledgeOfServer: snapshot.ServerKnowledge}
changes, _ := client.Budget().GetBudget(budgetID, filter)
```

**Use batch operations** instead of individual calls:
```go
// Good: 1 request
client.Transaction().CreateTransactions(budgetID, transactions)

// Bad: N requests  
for _, tx := range transactions {
    client.Transaction().CreateTransaction(budgetID, tx)
}
```

## Development

- Make sure you have Go 1.19 or later installed
- Run tests with `go test -race ./...`

## License

BSD-2-Clause
