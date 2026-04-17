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

func TestListBooks_Public(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books", nil)
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

	if res.Meta == nil {
		t.Fatal("expected meta to be present for paginated response")
	}

	if res.Meta.Total < 6 {
		t.Fatalf("expected at least 6 seeded books, got %d", res.Meta.Total)
	}
}

func TestListBooks_WithPagination(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books?page=1&per_page=2", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res response.APIResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Meta.PerPage != 2 {
		t.Fatalf("expected per_page=2, got %d", res.Meta.PerPage)
	}
}

func TestListBooks_WithSearch(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books?search=Silent+Echo", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res response.APIResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Meta.Total < 1 {
		t.Fatal("expected at least 1 result for 'Silent Echo' search")
	}
}

func TestGetBook_NotFound(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/nonexistent-id", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCreateBook_Authenticated(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	token := loginAndGetToken(t, env, "test-admin@test.com", "TestAdmin@2026!")

	body, _ := json.Marshal(map[string]interface{}{
		"title":    "Test Book Creation",
		"author":   "Test Author",
		"isbn":     "978-0-00-000000-1",
		"price":    100000,
		"status":   "draft",
		"stock":    50,
		"category": "Fiction",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateBook_Unauthenticated(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body, _ := json.Marshal(map[string]interface{}{
		"title":  "Unauthenticated Book",
		"author": "Nobody",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestCreateBook_ValidationError(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	token := loginAndGetToken(t, env, "test-admin@test.com", "TestAdmin@2026!")

	// Missing required fields
	body, _ := json.Marshal(map[string]interface{}{
		"title": "",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteBook_CustomerForbidden(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	token := loginAndGetToken(t, env, "customer@digitalpapyrus.web.id", "Demo@2026!")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/books/some-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}
}
