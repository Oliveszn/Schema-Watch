package store

import (
	"sync"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

type Store struct {
	mu        sync.RWMutex
	snapshots map[string]schema.Schema
	history   map[string][]schema.Diff
}

func New() *Store {
	return &Store{
		snapshots: make(map[string]schema.Schema),
		history:   make(map[string][]schema.Diff),
	}
}

func (s *Store) CheckAndUpdate(endpoint string, newSchema schema.Schema) *schema.Diff {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, seenBefore := s.snapshots[endpoint]
	s.snapshots[endpoint] = newSchema

	if !seenBefore {
		return nil
	}

	changes := schema.Compare(old, newSchema)
	diff := schema.NewDiff(endpoint, changes)
	if diff != nil {
		s.history[endpoint] = append(s.history[endpoint], *diff)
	}
	return diff
}

func (s *Store) History(endpoint string) []schema.Diff {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]schema.Diff, len(s.history[endpoint]))
	copy(out, s.history[endpoint])
	return out
}

func (s *Store) Endpoints() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, 0, len(s.snapshots))
	for endpoint := range s.snapshots {
		out = append(out, endpoint)
	}
	return out
}

func (s *Store) CurrentSchema(endpoint string) (schema.Schema, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sch, ok := s.snapshots[endpoint]
	return sch, ok
}
