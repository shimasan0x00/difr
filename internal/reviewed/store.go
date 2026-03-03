package reviewed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Store manages a set of reviewed file paths with thread-safe access and disk persistence.
type Store struct {
	mu    sync.RWMutex
	files map[string]struct{}
	path  string
}

// NewStore creates a new Store that persists to the given path.
func NewStore(path string) *Store {
	return &Store{
		files: make(map[string]struct{}),
		path:  path,
	}
}

// Add marks a file path as reviewed.
func (s *Store) Add(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.files[path] = struct{}{}

	if err := s.saveLocked(); err != nil {
		delete(s.files, path)
		return err
	}
	return nil
}

// Remove unmarks a file path as reviewed.
func (s *Store) Remove(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.files[path]; !ok {
		return nil
	}

	delete(s.files, path)

	if err := s.saveLocked(); err != nil {
		s.files[path] = struct{}{}
		return err
	}
	return nil
}

// Has returns whether the given path is marked as reviewed.
func (s *Store) Has(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.files[path]
	return ok
}

// List returns all reviewed file paths sorted alphabetically.
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, 0, len(s.files))
	for p := range s.files {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

// Clear removes all reviewed file paths.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.files
	s.files = make(map[string]struct{})

	if err := s.saveLocked(); err != nil {
		s.files = old
		return err
	}
	return nil
}

// Load reads reviewed file paths from disk. Missing file is not an error.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return err
	}

	s.files = make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if p != "" {
			s.files[p] = struct{}{}
		}
	}
	return nil
}

// saveLocked persists reviewed files to disk. Must be called with s.mu write lock held.
func (s *Store) saveLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	paths := make([]string, 0, len(s.files))
	for p := range s.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	data, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "reviewed-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, s.path)
}
