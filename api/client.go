// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

// Package api implements shared structures and behaviours of
// the API services
package api // import "github.com/brunomvsouza/ynab.go/api"

import "context"

// ClientReader contract for a read only client
type ClientReader interface {
	GET(url string, responseModel interface{}) error
}

// ClientWriter contract for a write only client
type ClientWriter interface {
	POST(url string, responseModel interface{}, requestBody []byte) error
	PUT(url string, responseModel interface{}, requestBody []byte) error
	PATCH(url string, responseModel interface{}, requestBody []byte) error
	DELETE(url string, responseModel interface{}) error
}

// ClientReaderWriter contract for a read-write client
type ClientReaderWriter interface {
	ClientReader
	ClientWriter
}

// ContextClientReader contract for a context-aware read only client
type ContextClientReader interface {
	GETWithContext(ctx context.Context, url string, responseModel interface{}) error
}

// ContextClientWriter contract for a context-aware write only client
type ContextClientWriter interface {
	POSTWithContext(ctx context.Context, url string, responseModel interface{}, requestBody []byte) error
	PUTWithContext(ctx context.Context, url string, responseModel interface{}, requestBody []byte) error
	PATCHWithContext(ctx context.Context, url string, responseModel interface{}, requestBody []byte) error
	DELETEWithContext(ctx context.Context, url string, responseModel interface{}) error
}

// ContextClientReaderWriter contract for a context-aware read-write client
type ContextClientReaderWriter interface {
	ContextClientReader
	ContextClientWriter
}

// FullClient contract for a client that supports both context-aware and regular methods
type FullClient interface {
	ClientReaderWriter
	ContextClientReaderWriter
}
