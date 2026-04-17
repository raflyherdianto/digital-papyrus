package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digitalpapyrus/backend/pkg/response"
	"github.com/digitalpapyrus/backend/tests/testutil"
)

func TestLogin_Success(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test-admin@test.com",
		"password": "TestAdmin@2026!",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
	if res.Data == nil {
		t.Fatal("expected data to contain token and user")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test-admin@test.com",
		"password": "WrongPassword123!",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"password": "SomePassword1!",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test-admin@test.com",
		"password": "",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMe_Authenticated(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	// Login first
	token := loginAndGetToken(t, env, "test-admin@test.com", "TestAdmin@2026!")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

// loginAndGetToken is a helper to authenticate and extract the JWT token.
func loginAndGetToken(t *testing.T, env *testutil.TestEnv, email, password string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status %d: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	data := res["data"].(map[string]interface{})
	return data["token"].(string)
}
