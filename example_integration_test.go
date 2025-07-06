// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

//go:build integration
// +build integration

package ynab_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/oauth"
)

// Example_oAuthIntegration demonstrates OAuth 2.0 integration with the YNAB client
func Example_oAuthIntegration() {
	// OAuth configuration
	config := ynab.NewOAuthConfig(
		"demo-client-id",
		"demo-client-secret",
		"https://myapp.com/oauth/callback",
	).WithReadOnlyScope()

	// Start OAuth flow
	flowManager := ynab.NewFlowManager(config).
		WithDefaultStorage(ynab.NewFileStorage(ynab.DefaultTokenPath()))

	authURL, state, err := flowManager.StartAuthorizationCodeFlow()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Visit: %s\n", authURL)

	// Simulate callback after user authorization
	callbackURL := "https://myapp.com/oauth/callback?code=demo_auth_code&state=" + state

	ctx := context.Background()
	token, err := flowManager.CompleteAuthorizationCodeFlow(ctx, callbackURL, state)
	if err != nil {
		log.Fatal(err)
	}

	// Create OAuth client
	client, err := ynab.NewOAuthClientFromToken(config, token)
	if err != nil {
		log.Fatal(err)
	}

	// Use the client - same API as token-based client
	user, err := client.User().GetUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authenticated user: %s\n", user.ID)

	// Output:
	// Visit: https://app.ynab.com/oauth/authorize?client_id=demo-client-id&redirect_uri=https%3A%2F%2Fmyapp.com%2Foauth%2Fcallback&response_type=code&scope=read-only&state=...
	// Authenticated user: demo-user-id
}

// Example_tokenBasedVsOAuth demonstrates the difference between token-based and OAuth authentication
func Example_tokenBasedVsOAuth() {
	// Method 1: Personal Access Token (simple but static)
	tokenClient := ynab.NewClient("your-personal-access-token")

	// Method 2: OAuth (secure with automatic refresh)
	config := ynab.NewOAuthConfig("client-id", "client-secret", "redirect-uri")

	oauthClient, err := ynab.NewOAuthClientBuilder(config).
		WithDefaultFileStorage().
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Both clients implement the same interface
	clients := []ynab.ClientServicer{tokenClient, oauthClient}

	for i, client := range clients {
		budgets, err := client.Budget().GetBudgets()
		if err != nil {
			log.Printf("Client %d error: %v", i+1, err)
			continue
		}

		fmt.Printf("Client %d found %d budgets\n", i+1, len(budgets))
	}

	// Output:
	// Client 1 found 3 budgets
	// Client 2 found 3 budgets
}

// Example_advancedOAuthUsage demonstrates advanced OAuth features
func Example_advancedOAuthUsage() {
	config := ynab.NewOAuthConfig(
		os.Getenv("YNAB_CLIENT_ID"),
		os.Getenv("YNAB_CLIENT_SECRET"),
		os.Getenv("YNAB_REDIRECT_URI"),
	)

	// Custom storage with encryption
	encryptionKey := []byte("your-32-byte-encryption-key-here")
	storage := oauth.NewEncryptedFileStorage("secure-tokens.json", encryptionKey)

	// Build client with advanced features
	client, err := ynab.NewOAuthClientBuilder(config).
		WithStorage(storage).
		WithTokenRefreshCallback(func(token *oauth.Token) {
			log.Printf("Token refreshed, expires: %v", token.ExpiresAt)
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Check authentication status
	if !client.IsAuthenticated() {
		log.Println("User needs to authenticate")
		// Start OAuth flow...
		return
	}

	// Use authenticated client
	user, err := client.User().GetUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authenticated as: %s\n", user.ID)

	// Manual token refresh (usually automatic)
	ctx := context.Background()
	newToken, err := client.RefreshToken(ctx)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
	} else {
		log.Printf("Token manually refreshed, expires: %v", newToken.ExpiresAt)
	}
}

// Example_errorHandling demonstrates comprehensive error handling for both auth methods
func Example_errorHandling() {
	// OAuth error handling
	config := ynab.NewOAuthConfig("client-id", "client-secret", "redirect-uri")
	flow := ynab.NewAuthorizationCodeFlow(config)

	// Simulate error callback
	errorCallbackURL := "https://myapp.com/callback?error=access_denied&error_description=User%20denied%20access"

	_, err := flow.HandleCallback(errorCallbackURL, "expected-state")
	if err != nil {
		if oauthErr, ok := err.(*oauth.ErrorResponse); ok {
			switch oauthErr.ErrorCode {
			case "access_denied":
				fmt.Println("User denied authorization")
			case "invalid_request":
				fmt.Println("Invalid OAuth request")
			default:
				fmt.Printf("OAuth error: %s\n", oauthErr.Error())
			}
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}

	// API error handling (same for both auth methods)
	client := ynab.NewClient("invalid-token")
	_, err = client.User().GetUser()
	if err != nil {
		// This works for both token-based and OAuth clients
		fmt.Printf("API error: %v\n", err)
	}

	// Output:
	// User denied authorization
	// Authentication failed
}

// Example_migration demonstrates migrating from token-based to OAuth authentication
func Example_migration() {
	// Legacy code using personal access token
	legacyClient := ynab.NewClient("personal-access-token")

	// New OAuth-based client with same API
	config := ynab.NewOAuthConfig("client-id", "client-secret", "redirect-uri")
	oauthClient, err := ynab.NewOAuthClientFromStorage(config, ynab.NewFileStorage("tokens.json"))
	if err != nil {
		// Fall back to legacy client if OAuth setup fails
		oauthClient = nil
	}

	// Function that works with either client type
	useClient := func(client ynab.ClientServicer, name string) {
		budgets, err := client.Budget().GetBudgets()
		if err != nil {
			log.Printf("%s client error: %v", name, err)
			return
		}
		fmt.Printf("%s client: %d budgets\n", name, len(budgets))
	}

	// Use legacy client
	useClient(legacyClient, "Legacy")

	// Use OAuth client if available
	if oauthClient != nil {
		useClient(oauthClient, "OAuth")
	}

	// Output:
	// Legacy client: 3 budgets
	// OAuth client: 3 budgets
}
