package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// ResultsStore holds check results per service, thread-safe.
type ResultsStore struct {
	mu      sync.RWMutex
	Results map[string][]CheckResult `json:"results"`
	path    string
	maxAge  time.Duration
}

// NewResultsStore loads existing data or creates a new store.
func NewResultsStore(path string, maxAge time.Duration) (*ResultsStore, error) {
	store := &ResultsStore{
		Results: make(map[string][]CheckResult),
		path:    path,
		maxAge:  maxAge,
	}

	// Try to load existing data
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, store); err != nil {
			// Corrupted file, start fresh
			fmt.Fprintf(os.Stderr, "warning: corrupted data file, starting fresh: %v\n", err)
			store.Results = make(map[string][]CheckResult)
		}
	}

	// Prune old entries on startup
	store.prune()

	return store, nil
}

// AddResult inserts a check result and prunes old data.
func (s *ResultsStore) AddResult(r CheckResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Results[r.ServiceName] = append(s.Results[r.ServiceName], r)
	s.pruneLocked()
	s.save()
}

// GetResults returns results for a service, newest first.
func (s *ResultsStore) GetResults(serviceName string) []CheckResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.Results[serviceName]
	sorted := make([]CheckResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, sorted[i].Timestamp)
		t2, _ := time.Parse(time.RFC3339, sorted[j].Timestamp)
		return t1.After(t2)
	})

	return sorted
}

// GetUptime calculates uptime percentage over the given duration.
func (s *ResultsStore) GetUptime(serviceName string, since time.Duration) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.Results[serviceName]
	if len(results) == 0 {
		return 100.0 // assume up if no data
	}

	cutoff := time.Now().Add(-since)
	total, success := 0, 0
	for _, r := range results {
		t, err := time.Parse(time.RFC3339, r.Timestamp)
		if err != nil || t.Before(cutoff) {
			continue
		}
		total++
		if r.Success {
			success++
		}
	}

	if total == 0 {
		return 100.0
	}
	return float64(success) / float64(total) * 100.0
}

// ServiceNames returns sorted list of service names with data.
func (s *ResultsStore) ServiceNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.Results))
	for name := range s.Results {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// prune removes entries older than maxAge.
func (s *ResultsStore) prune() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
}

func (s *ResultsStore) pruneLocked() {
	cutoff := time.Now().Add(-s.maxAge)
	for name, results := range s.Results {
		var kept []CheckResult
		for _, r := range results {
			t, err := time.Parse(time.RFC3339, r.Timestamp)
			if err != nil || t.After(cutoff) || t.Equal(cutoff) {
				kept = append(kept, r)
			}
		}
		if len(kept) == 0 {
			delete(s.Results, name)
		} else {
			s.Results[name] = kept
		}
	}
}

func (s *ResultsStore) save() {
	data, err := json.MarshalIndent(s.Results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling results: %v\n", err)
		return
	}

	dir := s.path[:len(s.path)]
	for i := len(s.path) - 1; i >= 0; i-- {
		if s.path[i] == '/' {
			dir = s.path[:i]
			break
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating data dir: %v\n", err)
		return
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		return
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming temp file: %v\n", err)
		return
	}
}
