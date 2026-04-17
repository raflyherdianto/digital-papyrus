package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digitalpapyrus/backend/pkg/response"
	"github.com/digitalpapyrus/backend/tests/testutil"
)

func TestListServices_Public(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var res response.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !res.Success {
		t.Fatal("expected success=true")
	}

	// Should have 4 seeded packages
	data, ok := res.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) < 4 {
		t.Fatalf("expected at least 4 seeded services, got %d", len(data))
	}
}

func TestGetService_NotFound(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services/nonexistent-id", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCreateService_CustomerForbidden(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	token := loginAndGetToken(t, env, "customer@digitalpapyrus.web.id", "Demo@2026!")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	// Customer should get 403 (Forbidden) since only admin can manage services
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteService_AuthorForbidden(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	token := loginAndGetToken(t, env, "author@digitalpapyrus.web.id", "Demo@2026!")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/services/some-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthCheck(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res response.APIResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	if !res.Success {
		t.Fatal("expected success=true for health check")
	}
}
