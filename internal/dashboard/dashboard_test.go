package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
	"github.com/Oliveszn/Schema-Watch/internal/store"
	"github.com/gin-gonic/gin"
)

func setupRouter(st *store.Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	d := New(st)
	d.RegisterRoutes(r)
	return r
}

func TestDashboard_ServesHTML(t *testing.T) {
	r := setupRouter(store.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__schema-watch/dashboard", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Fatal("expected a Content-Type header to be set")
	}
}

func TestDashboard_ListEndpoints_Empty(t *testing.T) {
	r := setupRouter(store.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__schema-watch/api/endpoints", nil)
	r.ServeHTTP(rec, req)

	var body struct {
		Endpoints []string `json:"endpoints"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(body.Endpoints) != 0 {
		t.Fatalf("expected no endpoints for a fresh store, got %v", body.Endpoints)
	}
}

func TestDashboard_ListEndpoints_ReturnsTrackedEndpoints(t *testing.T) {
	st := store.New()
	st.CheckAndUpdate("GET /users/1", schema.Schema{"id": schema.TypeNumber})

	r := setupRouter(st)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__schema-watch/api/endpoints", nil)
	r.ServeHTTP(rec, req)

	var body struct {
		Endpoints []string `json:"endpoints"`
	}
	json.Unmarshal(rec.Body.Bytes(), &body)

	if len(body.Endpoints) != 1 || body.Endpoints[0] != "GET /users/1" {
		t.Fatalf("expected [GET /users/1], got %v", body.Endpoints)
	}
}

func TestDashboard_GetHistory_RequiresEndpointParam(t *testing.T) {
	r := setupRouter(store.New())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__schema-watch/api/history", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when endpoint param missing, got %d", rec.Code)
	}
}

func TestDashboard_GetHistory_ReturnsRecordedDiffs(t *testing.T) {
	st := store.New()
	st.CheckAndUpdate("GET /users/1", schema.Schema{"id": schema.TypeNumber})
	st.CheckAndUpdate("GET /users/1", schema.Schema{"id": schema.TypeString}) // triggers a diff

	r := setupRouter(st)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__schema-watch/api/history?endpoint=GET+%2Fusers%2F1", nil)
	r.ServeHTTP(rec, req)

	var body struct {
		History []schema.Diff `json:"history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(body.History) != 1 {
		t.Fatalf("expected 1 recorded diff, got %d: %+v", len(body.History), body.History)
	}
	if !body.History[0].Breaking {
		t.Fatalf("expected recorded diff to be breaking, got %+v", body.History[0])
	}
}
