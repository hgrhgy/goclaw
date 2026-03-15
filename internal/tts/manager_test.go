package tts

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements Provider for testing
type mockProvider struct {
	name       string
	shouldFail bool
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Synthesize(ctx context.Context, text string, opts Options) (*SynthResult, error) {
	if m.shouldFail {
		return nil, errors.New("synthesize failed")
	}
	return &SynthResult{
		Audio:     []byte("fake-audio"),
		Extension: "mp3",
		MimeType:  "audio/mpeg",
	}, nil
}

func TestNewManager(t *testing.T) {
	cfg := ManagerConfig{
		Primary:   "openai",
		Auto:      AutoAlways,
		Mode:      ModeAll,
		MaxLength: 2000,
		TimeoutMs: 5000,
	}

	m := NewManager(cfg)

	assert.Equal(t, "openai", m.primary)
	assert.Equal(t, AutoAlways, m.auto)
	assert.Equal(t, ModeAll, m.mode)
	assert.Equal(t, 2000, m.maxLength)
	assert.Equal(t, 5000, m.timeoutMs)
}

func TestNewManagerDefaults(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "test"})

	assert.Equal(t, AutoOff, m.auto)    // default
	assert.Equal(t, ModeFinal, m.mode)  // default
	assert.Equal(t, 1500, m.maxLength)  // default
	assert.Equal(t, 30000, m.timeoutMs) // default
}

func TestManagerRegisterProvider(t *testing.T) {
	m := NewManager(ManagerConfig{})

	m.RegisterProvider(&mockProvider{name: "openai"})
	assert.True(t, m.HasProviders())

	p, ok := m.GetProvider("openai")
	require.True(t, ok)
	assert.Equal(t, "openai", p.Name())
}

func TestManagerGetProviderNotFound(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "openai"})

	_, ok := m.GetProvider("nonexistent")
	assert.False(t, ok)
}

func TestManagerPrimaryProvider(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "openai"})

	assert.Equal(t, "openai", m.PrimaryProvider())
}

func TestManagerAutoMode(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways})

	assert.Equal(t, AutoAlways, m.AutoMode())
}

func TestManagerSynthesize(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "openai"})
	m.RegisterProvider(&mockProvider{name: "openai"})

	result, err := m.Synthesize(context.Background(), "Hello world", Options{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mp3", result.Extension)
}

func TestManagerSynthesizeProviderNotFound(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "nonexistent"})

	_, err := m.Synthesize(context.Background(), "Hello", Options{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestManagerSynthesizeWithFallback(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "primary"})
	m.RegisterProvider(&mockProvider{name: "primary", shouldFail: true})
	m.RegisterProvider(&mockProvider{name: "fallback"})

	result, err := m.SynthesizeWithFallback(context.Background(), "Hello", Options{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestManagerSynthesizeAllFailed(t *testing.T) {
	m := NewManager(ManagerConfig{Primary: "p1"})
	m.RegisterProvider(&mockProvider{name: "p1", shouldFail: true})
	m.RegisterProvider(&mockProvider{name: "p2", shouldFail: true})

	_, err := m.SynthesizeWithFallback(context.Background(), "Hello", Options{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all tts providers failed")
}

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "code block removed",
			input:    "Hello ```go\ncode\n``` world",
			expected: "Hello  world",
		},
		{
			name:     "inline code",
			input:    "Hello `code` world",
			expected: "Hello code world",
		},
		{
			name:     "bold",
			input:    "Hello **bold** world",
			expected: "Hello bold world",
		},
		{
			name:     "italic",
			input:    "Hello *italic* world",
			expected: "Hello italic world",
		},
		{
			name:     "link",
			input:    "Hello [link](https://example.com) world",
			expected: "Hello link world",
		},
		{
			name:     "header",
			input:    "# Hello\nWorld",
			expected: "Hello\nWorld",
		},
		{
			name:     "no markdown",
			input:    "Hello world",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripTtsDirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tts block",
			input:    "Hello [[tts:text]]world[[/tts:text]]",
			expected: "Hello world",
		},
		{
			name:     "tts tag",
			input:    "Hello [[tts]] world",
			expected: "Hello  world",
		},
		{
			name:     "tts with param",
			input:    "Hello [[tts:voice1]] world",
			expected: "Hello  world",
		},
		{
			name:     "no directive",
			input:    "Hello world",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripTtsDirectives(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaybeApplyAutoOff(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoOff})

	result, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.Nil(t, result)
	assert.False(t, applied)
}

func TestMaybeApplyModeFilter(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways, Mode: ModeFinal})

	// Tool blocks should be skipped in ModeFinal
	result, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "tool")
	assert.Nil(t, result)
	assert.False(t, applied)

	// But "final" should work
	m.RegisterProvider(&mockProvider{name: "openai"})
	result, applied = m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.True(t, applied)
}

func TestMaybeApplyAutoInbound(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoInbound})
	m.RegisterProvider(&mockProvider{name: "openai"})

	// Should not apply without voice inbound
	result, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.Nil(t, result)
	assert.False(t, applied)

	// Should apply with voice inbound
	result, applied = m.MaybeApply(context.Background(), "Hello world", "telegram", true, "final")
	assert.NotNil(t, result)
	assert.True(t, applied)
}

func TestMaybeApplyAutoTagged(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoTagged})
	m.RegisterProvider(&mockProvider{name: "openai"})

	// Should not apply without tag
	result, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.Nil(t, result)
	assert.False(t, applied)

	// Should apply with [[tts]] tag
	result, applied = m.MaybeApply(context.Background(), "Hello [[tts]] world", "telegram", false, "final")
	assert.NotNil(t, result)
	assert.True(t, applied)
}

func TestMaybeApplyAutoAlways(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways})
	m.RegisterProvider(&mockProvider{name: "openai"})

	result, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.NotNil(t, result)
	assert.True(t, applied)
}

func TestMaybeApplyShortContent(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways})
	m.RegisterProvider(&mockProvider{name: "openai"})

	// Content too short
	result, applied := m.MaybeApply(context.Background(), "Hi", "telegram", false, "final")
	assert.Nil(t, result)
	assert.False(t, applied)
}

func TestMaybeApplyMediaContent(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways})
	m.RegisterProvider(&mockProvider{name: "openai"})

	// Contains MEDIA:
	result, applied := m.MaybeApply(context.Background(), "Hello MEDIA: something", "telegram", false, "final")
	assert.Nil(t, result)
	assert.False(t, applied)
}

func TestMaybeApplyTelegramFormat(t *testing.T) {
	m := NewManager(ManagerConfig{Auto: AutoAlways})
	m.RegisterProvider(&mockProvider{name: "openai"})

	// Should use opus for telegram
	_, applied := m.MaybeApply(context.Background(), "Hello world", "telegram", false, "final")
	assert.True(t, applied)
}
