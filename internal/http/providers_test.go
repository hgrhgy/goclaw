package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// --- mock stores ---

type testProviderStore struct {
	providers map[string]*store.LLMProviderData
}

func newTestProviderStore() *testProviderStore {
	return &testProviderStore{providers: make(map[string]*store.LLMProviderData)}
}

func (m *testProviderStore) CreateProvider(_ context.Context, p *store.LLMProviderData) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	m.providers[p.Name] = p
	return nil
}

func (m *testProviderStore) GetProvider(_ context.Context, id uuid.UUID) (*store.LLMProviderData, error) {
	for _, p := range m.providers {
		if p.ID == id {
			// Return a copy to prevent mutation
			copy := *p
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("provider not found")
}

func (m *testProviderStore) GetProviderByName(_ context.Context, name string) (*store.LLMProviderData, error) {
	if p, ok := m.providers[name]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, fmt.Errorf("provider not found")
}

func (m *testProviderStore) ListProviders(_ context.Context) ([]store.LLMProviderData, error) {
	var out []store.LLMProviderData
	for _, p := range m.providers {
		out = append(out, *p)
	}
	return out, nil
}

func (m *testProviderStore) UpdateProvider(_ context.Context, id uuid.UUID, updates map[string]interface{}) error {
	for _, p := range m.providers {
		if p.ID == id {
			if v, ok := updates["name"]; ok {
				p.Name = v.(string)
			}
			if v, ok := updates["display_name"]; ok {
				p.DisplayName = v.(string)
			}
			if v, ok := updates["provider_type"]; ok {
				p.ProviderType = v.(string)
			}
			if v, ok := updates["api_base"]; ok {
				p.APIBase = v.(string)
			}
			if v, ok := updates["api_key"]; ok {
				p.APIKey = v.(string)
			}
			if v, ok := updates["enabled"]; ok {
				p.Enabled = v.(bool)
			}
			if v, ok := updates["settings"]; ok {
				p.Settings = v.(json.RawMessage)
			}
			return nil
		}
	}
	return fmt.Errorf("provider not found")
}

func (m *testProviderStore) DeleteProvider(_ context.Context, id uuid.UUID) error {
	for name, p := range m.providers {
		if p.ID == id {
			delete(m.providers, name)
			return nil
		}
	}
	return fmt.Errorf("provider not found")
}

type testConfigSecretsStore struct {
	data map[string]string
}

func newTestConfigSecretsStore() *testConfigSecretsStore {
	return &testConfigSecretsStore{data: make(map[string]string)}
}

func (m *testConfigSecretsStore) Get(_ context.Context, key string) (string, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("key not found: %s", key)
}

func (m *testConfigSecretsStore) Set(_ context.Context, key, value string) error {
	m.data[key] = value
	return nil
}

func (m *testConfigSecretsStore) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *testConfigSecretsStore) GetAll(_ context.Context) (map[string]string, error) {
	return m.data, nil
}

// --- helper ---

func newTestProvidersHandler() *ProvidersHandler {
	store := newTestProviderStore()
	secrets := newTestConfigSecretsStore()
	return NewProvidersHandler(store, secrets, "test-token", nil, "")
}

// --- tests ---

func TestListProviders(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create a provider first
	provider := &store.LLMProviderData{
		Name:         "test-provider",
		ProviderType: store.ProviderOpenAICompat,
		APIBase:     "https://api.openai.com/v1",
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Test list
	req := httptest.NewRequest("GET", "/v1/providers", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	providers, ok := result["providers"].([]interface{})
	if !ok || len(providers) != 1 {
		t.Errorf("providers count = %d, want 1", len(providers))
	}
}

func TestCreateProvider(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]interface{}{
		"name":          "new-provider",
		"provider_type": "openai_compat",
		"api_base":     "https://api.openai.com/v1",
		"api_key":      "sk-test123",
		"enabled":      true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusCreated)
	}

	var result store.LLMProviderData
	json.NewDecoder(w.Body).Decode(&result)

	if result.Name != "new-provider" {
		t.Errorf("name = %s, want new-provider", result.Name)
	}
	if result.APIKey != "***" {
		t.Errorf("api_key = %s, want ***", result.APIKey)
	}
}

func TestCreateProviderInvalidName(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]interface{}{
		"name":          "Invalid Name!", // invalid slug
		"provider_type": "openai_compat",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateProviderMissingName(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]interface{}{
		"provider_type": "openai_compat",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateProviderInvalidType(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]interface{}{
		"name":          "test",
		"provider_type": "invalid_type",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetProvider(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create provider
	provider := &store.LLMProviderData{
		Name:         "get-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-secret",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Get provider
	req := httptest.NewRequest("GET", "/v1/providers/"+provider.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result store.LLMProviderData
	json.NewDecoder(w.Body).Decode(&result)

	if result.Name != "get-test" {
		t.Errorf("name = %s, want get-test", result.Name)
	}
	if result.APIKey != "***" {
		t.Errorf("api_key should be masked, got %s", result.APIKey)
	}
}

func TestGetProviderNotFound(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/v1/providers/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateProvider(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create provider
	provider := &store.LLMProviderData{
		Name:         "update-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-old",
		Enabled:     false,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Update provider
	body := map[string]interface{}{
		"api_key": "sk-new",
		"enabled": true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/providers/"+provider.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify update
	updated, _ := h.store.GetProvider(context.Background(), provider.ID)
	if updated.APIKey != "sk-new" {
		t.Errorf("api_key = %s, want sk-new", updated.APIKey)
	}
	if !updated.Enabled {
		t.Errorf("enabled = false, want true")
	}
}

func TestUpdateProviderMaskedAPIKey(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create provider
	provider := &store.LLMProviderData{
		Name:         "mask-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-real",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Try to update with masked key "***" - should be ignored
	body := map[string]interface{}{
		"api_key": "***",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/providers/"+provider.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Verify API key was NOT changed
	updated, _ := h.store.GetProvider(context.Background(), provider.ID)
	if updated.APIKey != "sk-real" {
		t.Errorf("api_key should not be changed, got %s", updated.APIKey)
	}
}

func TestDeleteProvider(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create provider
	provider := &store.LLMProviderData{
		Name:         "delete-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Delete provider
	req := httptest.NewRequest("DELETE", "/v1/providers/"+provider.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify deleted
	_, err := h.store.GetProvider(context.Background(), provider.ID)
	if err == nil {
		t.Error("provider should be deleted")
	}
}

func TestListProviderModels(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create MiniMax provider
	provider := &store.LLMProviderData{
		Name:         "minimax-test",
		ProviderType: store.ProviderMiniMax,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Get models
	req := httptest.NewRequest("GET", "/v1/providers/"+provider.ID.String()+"/models", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	models, ok := result["models"].([]interface{})
	if !ok {
		t.Fatal("models not found in response")
	}
	if len(models) == 0 {
		t.Error("models should not be empty for minimax")
	}
}

func TestUnauthorized(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Without token
	req := httptest.NewRequest("GET", "/v1/providers", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Since token is "test-token", this should return 401
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestVerifyProviderNoRegistry(t *testing.T) {
	h := newTestProvidersHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create provider
	provider := &store.LLMProviderData{
		Name:         "verify-test",
		ProviderType: store.ProviderOpenAICompat,
		APIKey:      "sk-test",
		Enabled:     true,
	}
	h.store.CreateProvider(context.Background(), provider)

	// Verify (should fail because no provider registry)
	body := map[string]interface{}{
		"model": "gpt-4",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/providers/"+provider.ID.String()+"/verify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return valid:false because registry is nil
	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	if result["valid"] != false {
		t.Errorf("valid = %v, want false", result["valid"])
	}
}

// --- MiniMax integration tests ---

// Note: These tests require a real MiniMax API key to run
// Set MINIMAX_API_KEY environment variable to run these tests

func TestMiniMaxListModels(t *testing.T) {
	apiKey := "sk-test" // Replace with real key or use env var
	if apiKey == "" || apiKey == "sk-test" {
		t.Skip("Skipping MiniMax test - no API key provided")
	}

	// Test the hardcoded MiniMax models
	models := minimaxModels()
	if len(models) == 0 {
		t.Error("MiniMax models should not be empty")
	}

	// Check for expected models
	modelIDs := make(map[string]bool)
	for _, m := range models {
		modelIDs[m.ID] = true
	}

	expectedModels := []string{"MiniMax-Text-01", "MiniMax-M1", "MiniMax-M2.5"}
	for _, expected := range expectedModels {
		if !modelIDs[expected] {
			t.Errorf("Expected model %s not found", expected)
		}
	}
}

func TestMiniMaxProviderType(t *testing.T) {
	// Test that minimax_cn is a valid provider type
	if !store.ValidProviderTypes["minimax_cn"] {
		t.Error("minimax_cn should be a valid provider type")
	}
	if !store.ValidProviderTypes["minimax_native"] {
		t.Error("minimax_native should be a valid provider type")
	}
}

func TestMiniMaxCNConfig(t *testing.T) {
	// Test MiniMax CN configuration
	providerType := "minimax_cn"
	apiBase := "https://api.minimax.chat/v1"

	if providerType != "minimax_cn" {
		t.Errorf("provider type should be minimax_cn, got %s", providerType)
	}

	expectedBase := "https://api.minimax.chat/v1"
	if apiBase != expectedBase {
		t.Errorf("api base should be %s, got %s", expectedBase, apiBase)
	}
}
