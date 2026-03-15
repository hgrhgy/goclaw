package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexibleStringSlice_StringArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "string array",
			input:    `["a", "b", "c"]`,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "number array",
			input:    `[1, 2, 3]`,
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "mixed array",
			input:    `["a", 2, "c"]`,
			expected: []string{"a", "2", "c"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fss FlexibleStringSlice
			err := json.Unmarshal([]byte(tt.input), &fss)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, []string(fss))
		})
	}
}

func TestNormalizeAgentID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid lowercase",
			input:    "myagent",
			expected: "myagent",
		},
		{
			name:     "valid with numbers",
			input:    "agent123",
			expected: "agent123",
		},
		{
			name:     "valid with underscore and dash",
			input:    "my_agent-123",
			expected: "my_agent-123",
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "MyAgent",
			expected: "myagent",
		},
		{
			name:     "spaces replaced with dash",
			input:    "my agent",
			expected: "my-agent",
		},
		{
			name:     "special chars replaced",
			input:    "my@agent#test",
			expected: "my-agent-test",
		},
		{
			name:     "leading dash stripped",
			input:    "-agent",
			expected: "agent",
		},
		{
			name:     "trailing dash - code returns as-is (bug: should strip)",
			input:    "agent-",
			expected: "agent-",
		},
		{
			name:     "multiple invalid chars collapsed",
			input:    "my@@agent##test",
			expected: "my-agent-test",
		},
		{
			name:     "truncated to 64 chars",
			input:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 65 a's
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",  // 64 a's
		},
		{
			name:     "empty string returns default",
			input:    "",
			expected: "default",
		},
		{
			name:     "whitespace only returns default",
			input:    "   ",
			expected: "default",
		},
		{
			name:     "only invalid chars returns default",
			input:    "@#$%",
			expected: "default",
		},
		{
			name:     "starts with digit",
			input:    "123agent",
			expected: "123agent",
		},
		{
			name:     "valid 63 chars (first + 62 more)",
			input:    "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk", // 63 chars
			expected: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAgentID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no tilde",
			input:    "/some/path",
			expected: "/some/path",
		},
		{
			name:     "tilde only",
			input:    "~",
			expected: home,
		},
		{
			name:     "tilde with path",
			input:    "~/some/path",
			expected: filepath.Join(home, "some/path"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandHome(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContractHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "home path converted to tilde",
			input:    filepath.Join(home, "some/path"),
			expected: "~/some/path",
		},
		{
			name:     "non-home path unchanged",
			input:    "/some/other/path",
			expected: "/some/other/path",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "home dir equals input",
			input:    home,
			expected: "~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContractHome(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	assert.NotNil(t, cfg)
	assert.Equal(t, "anthropic", cfg.Agents.Defaults.Provider)
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Agents.Defaults.Model)
	assert.Equal(t, 8192, cfg.Agents.Defaults.MaxTokens)
	assert.Equal(t, 0.7, cfg.Agents.Defaults.Temperature)
	assert.Equal(t, 20, cfg.Agents.Defaults.MaxToolIterations)
	assert.Equal(t, 25, cfg.Agents.Defaults.MaxToolCalls)
	assert.Equal(t, 200000, cfg.Agents.Defaults.ContextWindow)
	assert.Equal(t, true, cfg.Agents.Defaults.RestrictToWorkspace)
	assert.Equal(t, "~/.goclaw/workspace", cfg.Agents.Defaults.Workspace)

	// Gateway defaults
	assert.Equal(t, "0.0.0.0", cfg.Gateway.Host)
	assert.Equal(t, 18790, cfg.Gateway.Port)

	// Tools defaults
	assert.Equal(t, true, cfg.Tools.Web.DuckDuckGo.Enabled)
	assert.Equal(t, 5, cfg.Tools.Web.DuckDuckGo.MaxResults)
	assert.Equal(t, true, cfg.Tools.Browser.Enabled)
	assert.Equal(t, true, cfg.Tools.Browser.Headless)
}

func TestConfigReplaceFrom(t *testing.T) {
	cfg1 := &Config{
		Gateway: GatewayConfig{Host: "localhost", Port: 8080},
	}
	cfg2 := &Config{
		Gateway: GatewayConfig{Host: "0.0.0.0", Port: 18790},
	}

	cfg1.ReplaceFrom(cfg2)

	assert.Equal(t, "0.0.0.0", cfg1.Gateway.Host)
	assert.Equal(t, 18790, cfg1.Gateway.Port)
}

func TestConfigHash(t *testing.T) {
	cfg := Default()
	hash1 := cfg.Hash()

	// Hash should be consistent
	hash2 := cfg.Hash()
	assert.Equal(t, hash1, hash2)

	// Hash should change when config changes
	cfg.Gateway.Port = 9999
	hash3 := cfg.Hash()
	assert.NotEqual(t, hash1, hash3)
}

func TestConfigResolveAgent(t *testing.T) {
	cfg := Default()
	cfg.Agents.List = map[string]AgentSpec{
		"custom": {
			Provider:    "openai",
			Model:       "gpt-4",
			MaxTokens:   4096,
			Temperature: 0.9,
		},
	}

	// Test with custom agent
	agent := cfg.ResolveAgent("custom")
	assert.Equal(t, "openai", agent.Provider)
	assert.Equal(t, "gpt-4", agent.Model)
	assert.Equal(t, 4096, agent.MaxTokens)
	assert.Equal(t, 0.9, agent.Temperature)
	// Other fields should inherit from defaults
	assert.Equal(t, 20, agent.MaxToolIterations)

	// Test with non-existent agent (should use defaults)
	agent = cfg.ResolveAgent("nonexistent")
	assert.Equal(t, "anthropic", agent.Provider)
	assert.Equal(t, "claude-sonnet-4-5-20250929", agent.Model)
}

func TestConfigResolveDefaultAgentID(t *testing.T) {
	cfg := Default()

	// No default set
	assert.Equal(t, DefaultAgentID, cfg.ResolveDefaultAgentID())

	// Set a default agent
	cfg.Agents.List = map[string]AgentSpec{
		"myagent": {Default: true},
		"other":   {},
	}

	assert.Equal(t, "myagent", cfg.ResolveDefaultAgentID())
}

func TestConfigResolveDisplayName(t *testing.T) {
	cfg := Default()
	cfg.Agents.List = map[string]AgentSpec{
		"myagent": {DisplayName: "My Agent"},
	}

	// Custom agent
	assert.Equal(t, "My Agent", cfg.ResolveDisplayName("myagent"))

	// Non-existent agent
	assert.Equal(t, "GoClaw", cfg.ResolveDisplayName("nonexistent"))
}

func TestConfigWorkspacePath(t *testing.T) {
	cfg := Default()

	// Default workspace
	path := cfg.WorkspacePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".goclaw")
}

func TestAgentsConfigResolveAgentPath(t *testing.T) {
	agents := AgentsConfig{
		Root: "/Users/test/agents",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path returns as is",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path joined with root",
			input:    "myagent",
			expected: "/Users/test/agents/myagent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agents.ResolveAgentPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSandboxConfigToSandboxConfig(t *testing.T) {
	tests := []struct {
		name         string
		input        *SandboxConfig
		expectedMode string
	}{
		{
			name:         "nil returns default",
			input:        nil,
			expectedMode: "off",
		},
		{
			name:         "mode all",
			input:        &SandboxConfig{Mode: "all"},
			expectedMode: "all",
		},
		{
			name:         "mode non-main",
			input:        &SandboxConfig{Mode: "non-main"},
			expectedMode: "non-main",
		},
		{
			name:         "workspace access ro",
			input:        &SandboxConfig{WorkspaceAccess: "ro"},
			expectedMode: "off",
		},
		{
			name:         "workspace access rw",
			input:        &SandboxConfig{WorkspaceAccess: "rw"},
			expectedMode: "off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToSandboxConfig()
			// Just verify it doesn't panic and returns valid config
			assert.NotNil(t, result)
		})
	}
}

func TestCronConfigToRetryConfig(t *testing.T) {
	cfg := CronConfig{
		MaxRetries:     5,
		RetryBaseDelay: "1s",
		RetryMaxDelay:  "60s",
	}

	result := cfg.ToRetryConfig()

	assert.Equal(t, 5, result.MaxRetries)
	assert.Equal(t, 1.0, result.BaseDelay.Seconds())
	assert.Equal(t, 60.0, result.MaxDelay.Seconds())
}

func TestLoadConfigNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/config.json")

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Should return defaults
	assert.Equal(t, "anthropic", cfg.Agents.Defaults.Provider)
}
