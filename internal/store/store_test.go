package store

import (
	"sync"
	"testing"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

func TestCheckAndUpdate_FirstSeenReturnsNilNoDiff(t *testing.T) {
	s := New()
	sch := schema.Schema{"id": schema.TypeNumber}

	diff := s.CheckAndUpdate("GET /users", sch)
	if diff != nil {
		t.Fatalf("expected nil diff on first sighting, got %+v", diff)
	}

	got, ok := s.CurrentSchema("GET /users")
	if !ok {
		t.Fatal("expected schema to be recorded after first sighting")
	}
	if got["id"] != schema.TypeNumber {
		t.Fatalf("unexpected stored schema: %+v", got)
	}
}

func TestCheckAndUpdate_NoChangeReturnsNil(t *testing.T) {
	s := New()
	sch := schema.Schema{"id": schema.TypeNumber}

	s.CheckAndUpdate("GET /users", sch) // baseline

	diff := s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeNumber})
	if diff != nil {
		t.Fatalf("expected nil diff for unchanged schema, got %+v", diff)
	}
}

func TestCheckAndUpdate_DetectsChangeAndRecordsHistory(t *testing.T) {
	s := New()
	s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeNumber})

	diff := s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeString})
	if diff == nil {
		t.Fatal("expected a diff for changed field type, got nil")
	}
	if !diff.Breaking {
		t.Fatalf("expected type change to be breaking, got %+v", diff)
	}

	hist := s.History("GET /users")
	if len(hist) != 1 {
		t.Fatalf("expected 1 diff in history, got %d: %+v", len(hist), hist)
	}
}

func TestCheckAndUpdate_UpdatesBaselineAfterDiff(t *testing.T) {
	s := New()
	s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeNumber})
	s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeString})

	// third call with the same (now-current) schema should be a no-op
	diff := s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeString})
	if diff != nil {
		t.Fatalf("expected nil diff once schema stabilizes at new shape, got %+v", diff)
	}
}

func TestHistory_EmptyForUnknownEndpoint(t *testing.T) {
	s := New()
	hist := s.History("GET /never-seen")
	if hist == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(hist) != 0 {
		t.Fatalf("expected empty history, got %+v", hist)
	}
}

func TestEndpoints_TracksDistinctEndpoints(t *testing.T) {
	s := New()
	s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeNumber})
	s.CheckAndUpdate("GET /orders", schema.Schema{"id": schema.TypeNumber})

	got := s.Endpoints()
	if len(got) != 2 {
		t.Fatalf("expected 2 tracked endpoints, got %d: %v", len(got), got)
	}
}

func TestCheckAndUpdate_ConcurrentAccess(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.CheckAndUpdate("GET /users", schema.Schema{"id": schema.TypeNumber})
		}(i)
	}
	wg.Wait()

	if _, ok := s.CurrentSchema("GET /users"); !ok {
		t.Fatal("expected schema to be recorded after concurrent writes")
	}
}
