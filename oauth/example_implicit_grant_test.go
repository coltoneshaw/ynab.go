// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth_test

import (
	"fmt"
	"log"

	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/oauth"
)

func ExampleImplicitGrantFlow() {
	// Step 1: Create OAuth configuration (no client secret needed for implicit flow)
	config := ynab.NewOAuthConfig(
		"your-client-id",
		"", // No client secret for implicit grant
		"https://yourapp.com/oauth/callback",
	).WithReadOnlyScope()

	// Step 2: Create implicit grant flow
	flow := ynab.NewImplicitGrantFlow(config)

	// Step 3: Generate authorization URL
	state, err := config.GenerateState()
	if err != nil {
		log.Fatal(err)
	}

	authURL, err := flow.GetAuthorizationURL(state)
	if err != nil {
		log.Fatal(err)
	}

	// Step 4: Redirect user to authURL (in browser/mobile app)
	fmt.Printf("Visit this URL to authorize: %s\n", authURL)

	// Step 5: Handle callback (URL fragment contains access token)
	// Note: The access token is in the URL fragment, not query parameters
	callbackURL := "https://yourapp.com/oauth/callback#access_token=ya29.access_token&token_type=Bearer&expires_in=7200&state=" + state

	token, err := flow.HandleCallback(callbackURL, state)
	if err != nil {
		log.Fatal(err)
	}

	// Step 6: Create OAuth client with token
	client, err := ynab.NewOAuthClientFromToken(config, token)
	if err != nil {
		log.Fatal(err)
	}

	// Step 7: Use the client (note: no refresh token with implicit grant)
	user, err := client.User().GetUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authenticated user ID: %s\n", user.ID)
	fmt.Printf("Token expires: %v\n", token.ExpiresAt)
	fmt.Printf("Can refresh: %v\n", token.CanRefresh()) // Will be false for implicit grant
}

func ExampleImplicitGrantFlow_jsApp() {
	// Example for JavaScript/SPA applications
	config := oauth.NewConfig(
		"your-client-id",
		"", // No secret for client-side apps
		"https://yourapp.com/callback",
	)

	flow := oauth.NewImplicitGrantFlow(config)

	// Generate authorization URL for JavaScript redirect
	authURL, err := flow.GetAuthorizationURL("random-state-123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("JavaScript should redirect to: %s\n", authURL)

	// In JavaScript, after redirect, you'd get the token from URL fragment:
	// const urlFragment = window.location.hash;
	// const params = new URLSearchParams(urlFragment.substring(1));
	// const accessToken = params.get('access_token');

	// Back in Go, when you receive the token from the frontend:
	callbackURL := "https://yourapp.com/callback#access_token=received_token&token_type=Bearer&expires_in=7200&state=random-state-123"

	token, err := flow.HandleCallback(callbackURL, "random-state-123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received access token: %s\n", token.AccessToken)
	fmt.Printf("Token type: %s\n", token.TokenType)
	fmt.Printf("Expires in: %d seconds\n", token.ExpiresIn)
}

func ExampleImplicitGrantFlow_mobileApp() {
	// Example for mobile applications using implicit grant
	config := oauth.NewConfig(
		"your-mobile-client-id",
		"",
		"https://yourapp.com/mobile/callback", // Or custom URL scheme like "yourapp://oauth"
	)

	flow := oauth.NewImplicitGrantFlow(config)

	// Mobile app would open this URL in browser/webview
	authURL, err := flow.GetAuthorizationURL("mobile-state-456")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mobile app should open: %s\n", authURL)

	// After user authorizes, the callback URL with token in fragment is called
	// Mobile app intercepts this URL and extracts the token
	interceptedURL := "https://yourapp.com/mobile/callback#access_token=mobile_token&token_type=Bearer&expires_in=7200&state=mobile-state-456"

	token, err := flow.HandleCallback(interceptedURL, "mobile-state-456")
	if err != nil {
		log.Fatal(err)
	}

	// Create client for mobile app
	client, err := ynab.NewOAuthClientBuilder(config).
		WithMemoryStorage(). // Mobile apps typically use memory storage
		WithToken(token).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Use client in mobile app
	budgets, err := client.Budget().GetBudgets()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mobile app loaded %d budgets\n", len(budgets))
}

func ExampleFlowManager_recommendedFlow() {
	config := oauth.NewConfig("client-id", "client-secret", "redirect-uri")
	
	// Get recommendation for flow type
	isServerSide := true
	needsRefreshToken := true
	
	recommendedFlow := oauth.RecommendFlow(isServerSide, needsRefreshToken)
	
	fmt.Printf("Recommended flow: %s\n", recommendedFlow)
	
	manager := oauth.NewFlowManager(config)
	flow := manager.GetFlow(recommendedFlow)
	
	authURL, err := flow.GetAuthorizationURL("state-123")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Authorization URL: %s\n", authURL)
}