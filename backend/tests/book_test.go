package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	if res.Meta.Total < 4 {
		t.Fatalf("expected at least 4 seeded books, got %d", res.Meta.Total)
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books?search=TRANSFORMASI", nil)
	w := httptest.NewRecorder()

	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res response.APIResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Meta.Total < 1 {
		t.Fatal("expected at least 1 result for 'TRANSFORMASI' search")
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

	testISBN := fmt.Sprintf("978-%d", time.Now().UnixNano()%10000000000)
	body, _ := json.Marshal(map[string]interface{}{
		"title":    "Test Book Creation",
		"author":   "Test Author",
		"isbn":     testISBN,
		"badge":    "Limited Edition",
		"ggkey":    "GGK-TEST-001",
		"qrcbn":    "QRC-TEST-001",
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

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			ID    string `json:"id"`
			Badge string `json:"badge"`
			GGKEY string `json:"ggkey"`
			QRCBN string `json:"qrcbn"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if res.Data.GGKEY != "GGK-TEST-001" {
		t.Errorf("expected ggkey 'GGK-TEST-001', got '%s'", res.Data.GGKEY)
	}
	if res.Data.QRCBN != "QRC-TEST-001" {
		t.Errorf("expected qrcbn 'QRC-TEST-001', got '%s'", res.Data.QRCBN)
	}
	if res.Data.Badge != "Limited Edition" {
		t.Errorf("expected badge 'Limited Edition', got '%s'", res.Data.Badge)
	}

	// Verify persistence via GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/books/"+res.Data.ID, nil)
	getW := httptest.NewRecorder()
	env.Router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected status 200 on get, got %d", getW.Code)
	}
	var getRes struct {
		Data struct {
			Badge string `json:"badge"`
			GGKEY string `json:"ggkey"`
			QRCBN string `json:"qrcbn"`
		} `json:"data"`
	}
	json.Unmarshal(getW.Body.Bytes(), &getRes)
	if getRes.Data.GGKEY != "GGK-TEST-001" || getRes.Data.QRCBN != "QRC-TEST-001" || getRes.Data.Badge != "Limited Edition" {
		t.Errorf("persisted values mismatch: %+v", getRes.Data)
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
