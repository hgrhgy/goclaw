package sessions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		channel  string
		kind     PeerKind
		chatID   string
		expected string
	}{
		{
			name:     "DM session",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerDirect,
			chatID:   "123456",
			expected: "agent:default:telegram:direct:123456",
		},
		{
			name:     "Group session",
			agentID:  "default",
			channel:  "discord",
			kind:     PeerGroup,
			chatID:   "-100123456",
			expected: "agent:default:discord:group:-100123456",
		},
		{
			name:     "Slack DM",
			agentID:  "myagent",
			channel:  "slack",
			kind:     PeerDirect,
			chatID:   "U123456",
			expected: "agent:myagent:slack:direct:U123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSessionKey(tt.agentID, tt.channel, tt.kind, tt.chatID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildGroupTopicSessionKey(t *testing.T) {
	result := BuildGroupTopicSessionKey("default", "telegram", "-100123456", 99)
	assert.Equal(t, "agent:default:telegram:group:-100123456:topic:99", result)
}

func TestBuildDMThreadSessionKey(t *testing.T) {
	result := BuildDMThreadSessionKey("default", "telegram", "123456", 789)
	assert.Equal(t, "agent:default:telegram:direct:123456:thread:789", result)
}

func TestBuildSubagentSessionKey(t *testing.T) {
	result := BuildSubagentSessionKey("default", "my-task")
	assert.Equal(t, "agent:default:subagent:my-task", result)
}

func TestBuildCronSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		jobID    string
		expected string
	}{
		{
			name:     "simple job ID",
			agentID:  "default",
			jobID:    "reminder-job",
			expected: "agent:default:cron:reminder-job",
		},
		{
			name:     "already has prefix - strips first two parts",
			agentID:  "default",
			jobID:    "agent:foo:cron:reminder-job",
			expected: "agent:default:cron:cron:reminder-job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCronSessionKey(tt.agentID, tt.jobID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildAgentMainSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		mainKey  string
		expected string
	}{
		{
			name:     "default main key",
			agentID:  "default",
			mainKey:  "",
			expected: "agent:default:main",
		},
		{
			name:     "custom main key",
			agentID:  "default",
			mainKey:  "shared",
			expected: "agent:default:shared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAgentMainSessionKey(tt.agentID, tt.mainKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildScopedSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		channel  string
		kind     PeerKind
		chatID   string
		scope    string
		dmScope  string
		mainKey  string
		expected string
	}{
		{
			name:     "global scope",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerDirect,
			chatID:   "123",
			scope:    "global",
			expected: "global",
		},
		{
			name:     "group always uses full key",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerGroup,
			chatID:   "-100123",
			scope:    "per-sender",
			expected: "agent:default:telegram:group:-100123",
		},
		{
			name:     "DM per-channel-peer (default)",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerDirect,
			chatID:   "123",
			scope:    "per-sender",
			dmScope:  "",
			expected: "agent:default:telegram:direct:123",
		},
		{
			name:     "DM main scope",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerDirect,
			chatID:   "123",
			scope:    "per-sender",
			dmScope:  "main",
			mainKey:  "",
			expected: "agent:default:main",
		},
		{
			name:     "DM per-peer scope",
			agentID:  "default",
			channel:  "telegram",
			kind:     PeerDirect,
			chatID:   "123",
			scope:    "per-sender",
			dmScope:  "per-peer",
			expected: "agent:default:direct:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildScopedSessionKey(tt.agentID, tt.channel, tt.kind, tt.chatID, tt.scope, tt.dmScope, tt.mainKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSessionKey(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectedID   string
		expectedRest string
	}{
		{
			name:         "valid DM key",
			key:          "agent:default:telegram:direct:123",
			expectedID:   "default",
			expectedRest: "telegram:direct:123",
		},
		{
			name:         "valid group key",
			key:          "agent:myagent:discord:group:-100123",
			expectedID:   "myagent",
			expectedRest: "discord:group:-100123",
		},
		{
			name:         "subagent key",
			key:          "agent:default:subagent:task1",
			expectedID:   "default",
			expectedRest: "subagent:task1",
		},
		{
			name:         "invalid - no agent prefix",
			key:          "foo:default:telegram",
			expectedID:   "",
			expectedRest: "",
		},
		{
			name:         "invalid - too few parts",
			key:          "agent:default",
			expectedID:   "",
			expectedRest: "",
		},
		{
			name:         "empty key",
			key:          "",
			expectedID:   "",
			expectedRest: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentID, rest := ParseSessionKey(tt.key)
			assert.Equal(t, tt.expectedID, agentID)
			assert.Equal(t, tt.expectedRest, rest)
		})
	}
}

func TestIsSubagentSession(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"subagent key", "agent:default:subagent:task1", true},
		{"Subagent uppercase", "agent:default:SubAgent:task1", true},
		{"cron key", "agent:default:cron:job1", false},
		{"DM key", "agent:default:telegram:direct:123", false},
		{"global key", "global", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubagentSession(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCronSession(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"cron key", "agent:default:cron:job1", true},
		{"CRON uppercase", "agent:default:CRON:job1", true},
		{"subagent key", "agent:default:subagent:task1", false},
		{"DM key", "agent:default:telegram:direct:123", false},
		{"global key", "global", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCronSession(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPeerKindFromGroup(t *testing.T) {
	assert.Equal(t, PeerGroup, PeerKindFromGroup(true))
	assert.Equal(t, PeerDirect, PeerKindFromGroup(false))
}
