package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT_English(t *testing.T) {
	// Register test messages
	register("en", map[string]string{
		"greeting": "Hello, %s!",
		"welcome":  "Welcome to GoClaw",
	})

	// Basic message
	msg := T(LocaleEN, "welcome")
	assert.Equal(t, "Welcome to GoClaw", msg)

	// Message with args
	msg = T(LocaleEN, "greeting", "World")
	assert.Equal(t, "Hello, World!", msg)
}

func TestT_Fallback(t *testing.T) {
	// Register English only
	register("en", map[string]string{
		"hello": "Hello",
	})

	// Vietnamese (not registered) falls back to English
	msg := T(LocaleVI, "hello")
	assert.Equal(t, "Hello", msg)

	// Unknown key returns the key itself
	msg = T(LocaleEN, "unknown.key")
	assert.Equal(t, "unknown.key", msg)
}

func TestT_Chinese(t *testing.T) {
	// Register Chinese messages
	register("zh", map[string]string{
		"greeting": "你好, %s!",
		"welcome":  "欢迎使用 GoClaw",
	})

	msg := T(LocaleZH, "welcome")
	assert.Equal(t, "欢迎使用 GoClaw", msg)

	msg = T(LocaleZH, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", msg)
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected bool
	}{
		{"English", LocaleEN, true},
		{"Vietnamese", LocaleVI, true},
		{"Chinese", LocaleZH, true},
		{"Unknown", "fr", false},
		{"Empty", "", false},
		{"Case sensitive", "EN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupported(tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"English code", "en", LocaleEN},
		{"Vietnamese code", "vi", LocaleVI},
		{"Chinese code", "zh", LocaleZH},
		{"English US", "en-US", LocaleEN},
		{"Vietnamese VN", "vi-VN", LocaleVI},
		{"Chinese CN", "zh-CN", LocaleZH},
		{"Chinese TW", "zh-TW", LocaleZH},
		{"Unknown returns default", "fr", DefaultLocale},
		{"Empty returns default", "", DefaultLocale},
		{"Partial match", "eng", DefaultLocale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLookup(t *testing.T) {
	// Register test catalog
	register("test", map[string]string{
		"key1": "Value 1",
		"key2": "Value 2",
	})

	tests := []struct {
		name     string
		locale   string
		key      string
		expected string
	}{
		{"Found in locale", "test", "key1", "Value 1"},
		{"Fallback to English", "test", "unknown", "unknown"},
		{"Not found returns key", "en", "unknown.key", "unknown.key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookup(tt.locale, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
