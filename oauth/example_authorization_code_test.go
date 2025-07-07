package oauth_test

import (
	"context"
	"fmt"
	"log"

	"github.com/coltoneshaw/ynab.go"
	"github.com/coltoneshaw/ynab.go/oauth"
)

func ExampleAuthorizationCodeFlow() {
	// Step 1: Create OAuth configuration
	config := ynab.NewOAuthConfig(
		"your-client-id",
		"your-client-secret",
		"https://yourapp.com/oauth/callback",
	).WithReadOnlyScope()

	// Step 2: Create flow manager
	flowManager := ynab.NewFlowManager(config).
		WithDefaultStorage(ynab.NewFileStorage(ynab.DefaultTokenPath()))

	// Step 3: Start authorization flow
	authURL, state, err := flowManager.StartAuthorizationCodeFlow()
	if err != nil {
		log.Fatal(err)
	}

	// Step 4: Redirect user to authURL
	fmt.Printf("Visit this URL to authorize: %s\n", authURL)

	// Step 5: Handle callback (after user authorizes)
	callbackURL := "https://yourapp.com/oauth/callback?code=received-auth-code&state=" + state

	ctx := context.Background()
	token, err := flowManager.CompleteAuthorizationCodeFlow(ctx, callbackURL, state)
	if err != nil {
		log.Fatal(err)
	}

	// Step 6: Create OAuth client with token
	client, err := ynab.NewOAuthClientFromToken(config, token)
	if err != nil {
		log.Fatal(err)
	}

	// Step 7: Use the client
	user, err := client.User().GetUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authenticated user ID: %s\n", user.ID)
}

func ExampleAuthorizationCodeFlow_withBuilder() {
	// Using the client builder for more control
	config := oauth.NewConfig(
		"your-client-id",
		"your-client-secret",
		"https://yourapp.com/oauth/callback",
	)

	// Build client with file storage and refresh callback
	client, err := ynab.NewOAuthClientBuilder(config).
		WithDefaultFileStorage().
		WithTokenRefreshCallback(func(token *oauth.Token) {
			fmt.Println("Token refreshed automatically")
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Start OAuth flow
	flow := oauth.NewAuthorizationCodeFlow(config)
	authURL, err := flow.GetAuthorizationURL("your-state-parameter")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authorization URL: %s\n", authURL)

	// After receiving callback, exchange code for token
	callbackURL := "https://yourapp.com/oauth/callback?code=auth-code&state=your-state-parameter"
	token, err := flow.HandleCallback(callbackURL, "your-state-parameter")
	if err != nil {
		log.Fatal(err)
	}

	// Set token in client
	if err := client.SetToken(token); err != nil {
		log.Fatal(err)
	}

	// Use client
	budgets, err := client.Budget().GetBudgets()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d budgets\n", len(budgets))
}

func ExampleAuthorizationCodeFlow_serverExample() {
	// Example of how to integrate with an HTTP server
	config := oauth.NewConfig(
		"your-client-id",
		"your-client-secret",
		"https://yourapp.com/oauth/callback",
	)

	flow := oauth.NewAuthorizationCodeFlow(config).
		WithTokenManager(oauth.NewTokenManager(config, oauth.NewFileStorage("tokens.json")))

	// In your HTTP handler for starting OAuth
	startOAuthHandler := func() {
		// Generate secure state parameter
		state, err := config.GenerateState()
		if err != nil {
			log.Fatal(err)
		}

		// Store state in session/database for validation
		// session.Set("oauth_state", state)

		authURL, err := flow.GetAuthorizationURL(state)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Redirect user to: %s\n", authURL)
	}

	// In your HTTP handler for OAuth callback
	callbackHandler := func(callbackURL string) {
		// Retrieve state from session/database
		// expectedState := session.Get("oauth_state")

		token, err := flow.HandleCallback(callbackURL, "expected-state")
		if err != nil {
			log.Fatal(err)
		}

		// Store token for user (in database, session, etc.)
		fmt.Printf("Received token: %s\n", token.AccessToken)

		// Create client for this user
		client, err := ynab.NewOAuthClientFromToken(config, token)
		if err != nil {
			log.Fatal(err)
		}

		// Use client to access YNAB API
		user, err := client.User().GetUser()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("User authenticated: %s\n", user.ID)
	}

	// Example usage
	startOAuthHandler()
	callbackHandler("https://yourapp.com/oauth/callback?code=auth-code&state=expected-state")
}
