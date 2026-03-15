package permissions

import (
	"testing"

	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestRoleLevel(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected int
	}{
		{"Admin", RoleAdmin, 3},
		{"Operator", RoleOperator, 2},
		{"Viewer", RoleViewer, 1},
		{"Unknown", "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleLevel(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPolicyEngine(t *testing.T) {
	ownerIDs := []string{"owner1", "owner2", "owner3"}
	pe := NewPolicyEngine(ownerIDs)

	assert.True(t, pe.IsOwner("owner1"))
	assert.True(t, pe.IsOwner("owner2"))
	assert.True(t, pe.IsOwner("owner3"))
	assert.False(t, pe.IsOwner("unknown"))
}

func TestPolicyEngineIsOwner(t *testing.T) {
	pe := NewPolicyEngine([]string{"owner1"})

	// Existing owner
	assert.True(t, pe.IsOwner("owner1"))

	// Non-owner
	assert.False(t, pe.IsOwner("owner2"))
	assert.False(t, pe.IsOwner(""))
}

func TestPolicyEngineCanAccess(t *testing.T) {
	pe := NewPolicyEngine([]string{})

	tests := []struct {
		name     string
		role     Role
		method   string
		expected bool
	}{
		// Admin methods
		{"Admin access admin method", RoleAdmin, protocol.MethodAgentsCreate, true},
		{"Operator access admin method", RoleOperator, protocol.MethodAgentsCreate, false},
		{"Viewer access admin method", RoleViewer, protocol.MethodAgentsCreate, false},

		// Write methods
		{"Admin access write method", RoleAdmin, protocol.MethodChatSend, true},
		{"Operator access write method", RoleOperator, protocol.MethodChatSend, true},
		{"Viewer access write method", RoleViewer, protocol.MethodChatSend, false},

		// Read methods
		{"Admin access read method", RoleAdmin, protocol.MethodSessionsList, true},
		{"Operator access read method", RoleOperator, protocol.MethodSessionsList, true},
		{"Viewer access read method", RoleViewer, protocol.MethodSessionsList, true},

		// Unknown role
		{"Unknown role access", "unknown", protocol.MethodSessionsList, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pe.CanAccess(tt.role, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPolicyEngineCanAccessWithScopes(t *testing.T) {
	pe := NewPolicyEngine([]string{})

	tests := []struct {
		name     string
		scopes   []Scope
		method   string
		expected bool
	}{
		// Admin scope
		{"Admin scope access admin method", []Scope{ScopeAdmin}, protocol.MethodAgentsCreate, true},
		{"Admin scope access write method", []Scope{ScopeAdmin}, protocol.MethodChatSend, true},
		{"Admin scope access read method", []Scope{ScopeAdmin}, protocol.MethodSessionsList, true},

		// Write scope
		{"Write scope access admin method", []Scope{ScopeWrite}, protocol.MethodAgentsCreate, false},
		{"Write scope access write method", []Scope{ScopeWrite}, protocol.MethodChatSend, true},
		{"Write scope access read method", []Scope{ScopeWrite}, protocol.MethodSessionsList, true},

		// Read scope
		{"Read scope access admin method", []Scope{ScopeRead}, protocol.MethodAgentsCreate, false},
		{"Read scope access write method", []Scope{ScopeRead}, protocol.MethodChatSend, false},
		{"Read scope access read method", []Scope{ScopeRead}, protocol.MethodSessionsList, true},

		// Empty scopes
		{"Empty scopes", []Scope{}, protocol.MethodSessionsList, false},

		// No scope restriction
		{"Approve scope access approvals", []Scope{ScopeApprovals}, "approvals.list", true},
		{"Pairing scope access pairing", []Scope{ScopePairing}, "pairing.list", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pe.CanAccessWithScopes(tt.scopes, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoleFromScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []Scope
		expected Role
	}{
		{"Admin scope", []Scope{ScopeAdmin}, RoleAdmin},
		{"Write scope", []Scope{ScopeWrite}, RoleOperator},
		{"Read scope", []Scope{ScopeRead}, RoleViewer},
		{"Admin and write", []Scope{ScopeAdmin, ScopeWrite}, RoleAdmin},
		{"Write and read", []Scope{ScopeWrite, ScopeRead}, RoleOperator},
		{"Empty scopes", []Scope{}, RoleViewer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoleFromScopes(tt.scopes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMethodRole(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected Role
	}{
		// Admin methods
		{"Config apply", protocol.MethodConfigApply, RoleAdmin},
		{"Config patch", protocol.MethodConfigPatch, RoleAdmin},
		{"Agents create", protocol.MethodAgentsCreate, RoleAdmin},
		{"Agents update", protocol.MethodAgentsUpdate, RoleAdmin},
		{"Agents delete", protocol.MethodAgentsDelete, RoleAdmin},

		// Write methods
		{"Chat send", protocol.MethodChatSend, RoleOperator},
		{"Chat abort", protocol.MethodChatAbort, RoleOperator},
		{"Sessions delete", protocol.MethodSessionsDelete, RoleOperator},
		{"Sessions reset", protocol.MethodSessionsReset, RoleOperator},
		{"Sessions patch", protocol.MethodSessionsPatch, RoleOperator},
		{"Cron create", protocol.MethodCronCreate, RoleOperator},
		{"Cron update", protocol.MethodCronUpdate, RoleOperator},
		{"Cron delete", protocol.MethodCronDelete, RoleOperator},

		// Read methods
		{"Sessions list", protocol.MethodSessionsList, RoleViewer},
		{"Agents list", protocol.MethodAgentsList, RoleViewer},
		{"Skills list", protocol.MethodSkillsList, RoleViewer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MethodRole(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMethodScopes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected []Scope
	}{
		{"Admin method", protocol.MethodAgentsCreate, []Scope{ScopeAdmin}},
		{"Write method", protocol.MethodChatSend, []Scope{ScopeWrite, ScopeAdmin}},
		{"Read method", protocol.MethodSessionsList, []Scope{ScopeRead, ScopeWrite, ScopeAdmin}},
		{"Approvals prefix", "approvals.list", []Scope{ScopeApprovals, ScopeAdmin}},
		{"Pairing prefix", "pairing.list", []Scope{ScopePairing, ScopeAdmin}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MethodScopes(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAdminMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"Config apply", protocol.MethodConfigApply, true},
		{"Agents create", protocol.MethodAgentsCreate, true},
		{"Channels toggle", protocol.MethodChannelsToggle, true},
		{"Chat send", protocol.MethodChatSend, false},
		{"Sessions list", protocol.MethodSessionsList, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAdminMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWriteMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"Chat send", protocol.MethodChatSend, true},
		{"Sessions delete", protocol.MethodSessionsDelete, true},
		{"Cron create", protocol.MethodCronCreate, true},
		{"Pairing list", "pairing.list", true},
		{"Exec approval", "exec.approval.request", true},
		{"Agents list", protocol.MethodAgentsList, false},
		{"Sessions list", protocol.MethodSessionsList, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWriteMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}
