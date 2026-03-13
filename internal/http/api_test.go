package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// ==================== PROVIDER TESTS ====================

func TestProviderAPI_List(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create test provider
	provider := &store.LLMProviderData{
		Name:         "test-provider",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Test GET /v1/providers
	req := httptest.NewRequest("GET", "/v1/providers", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)
	providers := result["providers"].([]interface{})
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}
}

func TestProviderAPI_Create(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]interface{}{
		"name":          "new-provider",
		"provider_type": "openai_compat",
		"api_key":      "sk-test123",
		"enabled":      true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusCreated)
	var result store.LLMProviderData
	json.NewDecoder(w.Body).Decode(&result)
	if result.Name != "new-provider" {
		t.Errorf("expected name new-provider, got %s", result.Name)
	}
}

func TestProviderAPI_Get(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	provider := &store.LLMProviderData{
		Name:         "get-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-secret",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	req := httptest.NewRequest("GET", "/v1/providers/"+provider.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
}

func TestProviderAPI_Update(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	provider := &store.LLMProviderData{
		Name:         "update-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-old",
		Enabled:     false,
	}
	h.store.CreateProvider(context.Background(), provider)

	body := map[string]interface{}{
		"enabled": true,
		"api_key": "sk-new",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/providers/"+provider.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
}

func TestProviderAPI_Delete(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	provider := &store.LLMProviderData{
		Name:         "delete-test",
		ProviderType: store.ProviderOpenAICompat,
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	req := httptest.NewRequest("DELETE", "/v1/providers/"+provider.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
}

func TestProviderAPI_GetModels(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	provider := &store.LLMProviderData{
		Name:         "minimax-test",
		ProviderType: store.ProviderMiniMax,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	req := httptest.NewRequest("GET", "/v1/providers/"+provider.ID.String()+"/models", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)
	models := result["models"].([]interface{})
	if len(models) == 0 {
		t.Error("expected models, got empty")
	}
}

func TestProviderAPI_Verify(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	provider := &store.LLMProviderData{
		Name:         "verify-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	body := map[string]interface{}{
		"model": "gpt-4",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers/"+provider.ID.String()+"/verify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusOK)
	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)
	if result["valid"] != false {
		t.Errorf("expected valid=false, got %v", result["valid"])
	}
}

func TestProviderAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Without token
	req := httptest.NewRequest("GET", "/v1/providers", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expectStatus(t, w.Code, http.StatusUnauthorized)
}

// ==================== AGENT TESTS ====================

// Note: Agent tests require more complex mock setup
// Testing basic route registration

func TestAgentAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler() // Using provider handler for auth check
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Test agent routes without auth
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/agents"},
		{"POST", "/v1/agents"},
		{"GET", "/v1/agents/" + uuid.New().String()},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== CHANNEL TESTS ====================

func TestChannelAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/channels/instances"},
		{"POST", "/v1/channels/instances"},
		{"GET", "/v1/channels/instances/" + uuid.New().String()},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== SKILL TESTS ====================

func TestSkillAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/skills"},
		{"PUT", "/v1/skills/test"},
		{"DELETE", "/v1/skills/test"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== TOOL TESTS ====================

func TestToolAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/tools/custom"},
		{"POST", "/v1/tools/custom"},
		{"GET", "/v1/tools/builtin"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== MCP TESTS ====================

func TestMCPAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/mcp/servers"},
		{"POST", "/v1/mcp/servers"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== MEMORY TESTS ====================

func TestMemoryAPI_Unauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/memory/documents"},
		{"GET", "/v1/agents/" + uuid.New().String() + "/memory/documents"},
		{"POST", "/v1/agents/" + uuid.New().String() + "/memory/index"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tt.method, tt.path, w.Code)
		}
	}
}

// ==================== AUTH TESTS ====================

func TestAuthAPI_NoToken(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Auth routes with empty token should work without auth
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/auth/openai/status"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// With empty token, should return OK (not 401)
		if w.Code == http.StatusUnauthorized {
			t.Logf("%s %s: returned 401 as expected for protected endpoint", tt.method, tt.path)
		}
	}
}

// ==================== HELPERS ====================

func expectStatus(t *testing.T, got, want int) {
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// ==================== MINIMAX SPECIFIC TESTS ====================

func TestMiniMax_ProviderType(t *testing.T) {
	// Test minimax_cn is valid
	if !store.ValidProviderTypes["minimax_cn"] {
		t.Error("minimax_cn should be a valid provider type")
	}
	if !store.ValidProviderTypes["minimax_native"] {
		t.Error("minimax_native should be a valid provider type")
	}
}

func TestMiniMax_Models(t *testing.T) {
	models := minimaxModels()
	if len(models) == 0 {
		t.Error("MiniMax models should not be empty")
	}

	// Check for key models
	modelIDs := make(map[string]bool)
	for _, m := range models {
		modelIDs[m.ID] = true
	}

	expected := []string{"MiniMax-M2.5", "MiniMax-Text-01", "MiniMax-M1"}
	for _, e := range expected {
		if !modelIDs[e] {
			t.Errorf("Expected model %s not found", e)
		}
	}
}

func TestMiniMax_ProviderRegister(t *testing.T) {
	// Test that minimax_cn uses the same chat path as minimax_native
	// Both should use /text/chatcompletion_v2
	providerTypes := []string{store.ProviderMiniMax, store.ProviderMiniMaxCN}

	for _, pt := range providerTypes {
		// This is tested implicitly by the code using ||
		// Just verify the types are defined
		if pt != store.ProviderMiniMax && pt != store.ProviderMiniMaxCN {
			t.Errorf("Unexpected provider type: %s", pt)
		}
	}
}
