package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	// Use hex-encoded 32-byte key (64 hex chars)
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	plaintext := "Hello, World!"

	// Encrypt
	ciphertext, err := Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.True(t, IsEncrypted(ciphertext))
	assert.Contains(t, ciphertext, "aes-gcm:")

	// Decrypt
	decrypted, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptEmptyInputs(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	plaintext := "Hello"

	// Empty key returns plaintext
	result, err := Encrypt(plaintext, "")
	require.NoError(t, err)
	assert.Equal(t, plaintext, result)

	// Empty plaintext returns plaintext
	result, err = Encrypt("", key)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestDecryptEmptyInputs(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	ciphertext := "some-value"

	// Empty key returns ciphertext unchanged
	result, err := Decrypt(ciphertext, "")
	require.NoError(t, err)
	assert.Equal(t, ciphertext, result)

	// Empty ciphertext returns empty
	result, err = Decrypt("", key)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestDecryptPlainText(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	plaintext := "plain-text-value"

	// Non-encrypted value returns as-is
	result, err := Decrypt(plaintext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, result)
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	// Invalid base64 with prefix should return original (treated as plain text)
	ciphertext := "aes-gcm:!!!invalid-base64!!!"
	result, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, ciphertext, result)
}

func TestDecryptWrongKey(t *testing.T) {
	plaintext := "Secret message"
	key1 := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	key2 := "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	// Encrypt with key1
	ciphertext, err := Encrypt(plaintext, key1)
	require.NoError(t, err)

	// Try to decrypt with key2 - should fail
	_, err = Decrypt(ciphertext, key2)
	assert.Error(t, err)
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "encrypted value",
			input:    "aes-gcm:abc123",
			expected: true,
		},
		{
			name:     "plain text",
			input:    "plain-text",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "similar prefix but not encrypted",
			input:    "aes-gcm", // missing colon
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncrypted(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeriveKey(t *testing.T) {
	// Test hex-encoded key (64 chars = 32 bytes)
	hexKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	key, err := DeriveKey(hexKey)
	require.NoError(t, err)
	assert.Len(t, key, 32)

	// Test raw 32-byte key
	rawKey := "12345678901234567890123456789012"
	key, err = DeriveKey(rawKey)
	require.NoError(t, err)
	assert.Len(t, key, 32)
}

func TestDeriveKeyInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "too short",
			input: "short-key",
		},
		{
			name:  "wrong length hex",
			input: "0123456789abcdef",
		},
		{
			name:  "invalid hex chars",
			input: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeriveKey(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestEncryptDecryptConsistency(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	messages := []string{
		"Hello World",
		"Hello, World!",
		"Unicode: 你好世界 🔐",
		"Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		"Long message: " + string(make([]byte, 1000)),
	}

	for _, msg := range messages {
		name := msg
		if len(name) > 20 {
			name = name[:20]
		}
		t.Run("message: "+name, func(t *testing.T) {
			ciphertext, err := Encrypt(msg, key)
			require.NoError(t, err)

			decrypted, err := Decrypt(ciphertext, key)
			require.NoError(t, err)
			assert.Equal(t, msg, decrypted)
		})
	}
}
