package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
	"github.com/Oliveszn/Schema-Watch/internal/store"
)

func TestProxy_ForwardsResponseUnchanged(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1,"email":"a@b.com"}`))
	}))
	defer backend.Close()

	st := store.New()
	p, err := New(backend.URL, st, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, req)

	want := `{"id":1,"email":"a@b.com"}`
	if rec.Body.String() != want {
		t.Fatalf("client response was altered: got %q, want %q", rec.Body.String(), want)
	}
}

func TestProxy_RecordsBaselineOnFirstRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1}`))
	}))
	defer backend.Close()

	st := store.New()
	p, err := New(backend.URL, st, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	p.Handler().ServeHTTP(httptest.NewRecorder(), req)

	sch, ok := st.CurrentSchema("GET /users/1")
	if !ok {
		t.Fatal("expected schema to be recorded after first request")
	}
	if sch["id"] != schema.TypeNumber {
		t.Fatalf("unexpected recorded schema: %+v", sch)
	}
}

func TestProxy_FiresOnDiffWhenSchemaChanges(t *testing.T) {
	var bodyToReturn string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bodyToReturn))
	}))
	defer backend.Close()

	st := store.New()

	var mu sync.Mutex
	var gotDiff *schema.Diff
	onDiff := func(d *schema.Diff) {
		mu.Lock()
		defer mu.Unlock()
		gotDiff = d
	}

	p, err := New(backend.URL, st, onDiff)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	bodyToReturn = `{"id":1}`
	p.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/1", nil))

	mu.Lock()
	if gotDiff != nil {
		mu.Unlock()
		t.Fatalf("expected no diff on first request, got %+v", gotDiff)
	}
	mu.Unlock()

	bodyToReturn = `{"id":"uuid-1234"}` // type changed number -> string
	p.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/1", nil))

	mu.Lock()
	defer mu.Unlock()
	if gotDiff == nil {
		t.Fatal("expected onDiff to fire on second request with changed schema")
	}
	if !gotDiff.Breaking {
		t.Fatalf("expected breaking diff for type change, got %+v", gotDiff)
	}
}

func TestProxy_SkipsNonJSONResponses(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>not json</html>`))
	}))
	defer backend.Close()

	st := store.New()
	p, err := New(backend.URL, st, nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "not json") {
		t.Fatalf("expected HTML body forwarded unchanged, got %q", rec.Body.String())
	}
	if _, ok := st.CurrentSchema("GET /page"); ok {
		t.Fatal("expected non-JSON response to be skipped, not recorded in store")
	}
}
