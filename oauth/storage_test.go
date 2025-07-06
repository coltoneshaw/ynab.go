// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package oauth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	// Initially no token
	assert.False(t, storage.HasToken())

	token, err := storage.LoadToken()
	assert.Error(t, err)
	assert.Nil(t, token)

	// Save a token
	testToken := &Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    TokenTypeBearer,
		ExpiresIn:    3600,
		Scope:        ScopeReadOnly,
	}
	testToken.SetExpiration(3600)

	err = storage.SaveToken(testToken)
	assert.NoError(t, err)
	assert.True(t, storage.HasToken())

	// Load the token
	loadedToken, err := storage.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)
	assert.Equal(t, testToken.RefreshToken, loadedToken.RefreshToken)
	assert.Equal(t, testToken.TokenType, loadedToken.TokenType)
	assert.Equal(t, testToken.Scope, loadedToken.Scope)

	// Clear the token
	err = storage.ClearToken()
	assert.NoError(t, err)
	assert.False(t, storage.HasToken())
}

func TestFileStorage(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_token.json")

	storage := NewFileStorage(filePath)

	// Initially no token
	assert.False(t, storage.HasToken())

	// Save a token
	testToken := &Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    TokenTypeBearer,
		ExpiresIn:    3600,
		Scope:        ScopeReadOnly,
	}
	testToken.SetExpiration(3600)

	err := storage.SaveToken(testToken)
	assert.NoError(t, err)
	assert.True(t, storage.HasToken())

	// Check file exists and has secure permissions
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode())

	// Load the token
	loadedToken, err := storage.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)
	assert.Equal(t, testToken.RefreshToken, loadedToken.RefreshToken)

	// Clear the token
	err = storage.ClearToken()
	assert.NoError(t, err)
	assert.False(t, storage.HasToken())

	// File should be removed
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileStorage_WithFileMode(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_token.json")

	storage := NewFileStorage(filePath).WithFileMode(0644)

	testToken := &Token{AccessToken: "test"}
	err := storage.SaveToken(testToken)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), fileInfo.Mode())
}

func TestFileStorage_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "dir", "token.json")

	storage := NewFileStorage(nestedPath)

	testToken := &Token{AccessToken: "test"}
	err := storage.SaveToken(testToken)
	assert.NoError(t, err)

	// Check that nested directories were created
	assert.True(t, storage.HasToken())
}

func TestFileStorage_GetFilePath(t *testing.T) {
	filePath := "/path/to/token.json"
	storage := NewFileStorage(filePath)

	assert.Equal(t, filePath, storage.GetFilePath())
}

func TestEncryptedFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "encrypted_token.json")
	key := []byte("test-encryption-key")

	storage := NewEncryptedFileStorage(filePath, key)

	testToken := &Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
	}

	// Save encrypted token
	err := storage.SaveToken(testToken)
	assert.NoError(t, err)
	assert.True(t, storage.HasToken())

	// Load and decrypt token
	loadedToken, err := storage.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)
	assert.Equal(t, testToken.RefreshToken, loadedToken.RefreshToken)

	// Check that file content is actually encrypted (not plain JSON)
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.NotContains(t, string(fileContent), "test-access-token")
}

func TestChainedStorage(t *testing.T) {
	memory1 := NewMemoryStorage()
	memory2 := NewMemoryStorage()

	chained := NewChainedStorage(memory1, memory2)

	// Initially no token in any storage
	assert.False(t, chained.HasToken())

	testToken := &Token{AccessToken: "test-token"}

	// Save token - should save to all storages
	err := chained.SaveToken(testToken)
	assert.NoError(t, err)

	assert.True(t, memory1.HasToken())
	assert.True(t, memory2.HasToken())
	assert.True(t, chained.HasToken())

	// Load token - should load from first available storage
	loadedToken, err := chained.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)

	// Clear from first storage, should still load from second
	err = memory1.ClearToken()
	assert.NoError(t, err)

	loadedToken, err = chained.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)

	// Clear all storages
	err = chained.ClearToken()
	assert.NoError(t, err)

	assert.False(t, memory1.HasToken())
	assert.False(t, memory2.HasToken())
	assert.False(t, chained.HasToken())
}

func TestDefaultTokenPath(t *testing.T) {
	path := DefaultTokenPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".config")
	assert.Contains(t, path, "ynab")
	assert.Contains(t, path, "token.json")
}

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name        string
		opts        StorageOptions
		expectError bool
		checkType   func(storage TokenStorage) bool
	}{
		{
			name: "Memory storage",
			opts: StorageOptions{Type: "memory"},
			checkType: func(storage TokenStorage) bool {
				_, ok := storage.(*MemoryStorage)
				return ok
			},
		},
		{
			name: "File storage with default path",
			opts: StorageOptions{Type: "file"},
			checkType: func(storage TokenStorage) bool {
				_, ok := storage.(*FileStorage)
				return ok
			},
		},
		{
			name: "File storage with custom path",
			opts: StorageOptions{
				Type:     "file",
				FilePath: "/custom/path/token.json",
			},
			checkType: func(storage TokenStorage) bool {
				fileStorage, ok := storage.(*FileStorage)
				return ok && fileStorage.GetFilePath() == "/custom/path/token.json"
			},
		},
		{
			name: "Encrypted storage",
			opts: StorageOptions{
				Type:       "encrypted",
				EncryptKey: []byte("test-key"),
			},
			checkType: func(storage TokenStorage) bool {
				_, ok := storage.(*EncryptedFileStorage)
				return ok
			},
		},
		{
			name:        "Encrypted storage without key",
			opts:        StorageOptions{Type: "encrypted"},
			expectError: true,
		},
		{
			name:        "Unknown storage type",
			opts:        StorageOptions{Type: "unknown"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.opts)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				if tt.checkType != nil {
					assert.True(t, tt.checkType(storage))
				}
			}
		})
	}
}

func TestFileStorage_ErrorCases(t *testing.T) {
	t.Run("Save nil token", func(t *testing.T) {
		storage := NewFileStorage("/tmp/test.json")
		err := storage.SaveToken(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be nil")
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		storage := NewFileStorage("/non/existent/path/token.json")
		token, err := storage.LoadToken()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("Clear non-existent token", func(t *testing.T) {
		storage := NewFileStorage("/non/existent/path/token.json")
		err := storage.ClearToken()
		assert.NoError(t, err) // Should not error for non-existent file
	})
}

func TestEncryptedFileStorage_ErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "token.json")

	t.Run("Save nil token", func(t *testing.T) {
		storage := NewEncryptedFileStorage(filePath, []byte("key"))
		err := storage.SaveToken(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be nil")
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		storage := NewEncryptedFileStorage("/non/existent/token.json", []byte("key"))
		token, err := storage.LoadToken()
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("test-encryption-key")
	storage := &EncryptedFileStorage{key: key}

	original := []byte("sensitive token data")
	encrypted := storage.encrypt(original)
	decrypted := storage.decrypt(encrypted)

	assert.NotEqual(t, original, encrypted) // Should be different when encrypted
	assert.Equal(t, original, decrypted)    // Should be same when decrypted
}

func TestEncryptWithEmptyKey(t *testing.T) {
	storage := &EncryptedFileStorage{key: []byte{}}

	data := []byte("test data")
	encrypted := storage.encrypt(data)

	assert.Equal(t, data, encrypted) // Should return original data with empty key
}
