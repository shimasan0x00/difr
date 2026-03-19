package comment

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("comment not found")

type Comment struct {
	ID             string    `json:"id"`
	FilePath       string    `json:"filePath"`
	Line           int       `json:"line"`
	Body           string    `json:"body"`
	ReviewCategory string    `json:"reviewCategory,omitempty"`
	Severity       string    `json:"severity,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty"`
}

type UpdateFields struct {
	Body           string
	ReviewCategory string
	Severity       string
}

var validCategories = map[string]bool{
	"":     true,
	"MUST": true,
	"IMO":  true,
	"Q":    true,
	"FYI":  true,
}

var validSeverities = map[string]bool{
	"":         true,
	"Critical": true,
	"High":     true,
	"Middle":   true,
	"Low":      true,
}

func ValidateCategory(v string) bool {
	return validCategories[v]
}

func ValidateSeverity(v string) bool {
	return validSeverities[v]
}

type Store struct {
	mu       sync.RWMutex
	comments map[string]*Comment
	path     string
	nextID   int
}

func NewStore(path string) *Store {
	return &Store{
		comments: make(map[string]*Comment),
		path:     path,
		nextID:   1,
	}
}

func (s *Store) Create(c *Comment) (*Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment := &Comment{
		ID:             fmt.Sprintf("c%d", s.nextID),
		FilePath:       c.FilePath,
		Line:           c.Line,
		Body:           c.Body,
		ReviewCategory: c.ReviewCategory,
		Severity:       c.Severity,
		CreatedAt:      time.Now(),
	}
	s.nextID++
	s.comments[comment.ID] = comment

	if err := s.saveLocked(); err != nil {
		// Rollback on save failure
		delete(s.comments, comment.ID)
		s.nextID--
		return nil, fmt.Errorf("persisting comment: %w", err)
	}

	copy := *comment
	return &copy, nil
}

func (s *Store) Get(id string) (*Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.comments[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *c
	return &copy, nil
}

func (s *Store) List(filePath string) []*Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Comment
	for _, c := range s.comments {
		if filePath == "" || c.FilePath == filePath {
			copy := *c
			result = append(result, &copy)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

func (s *Store) Update(id string, fields UpdateFields) (*Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.comments[id]
	if !ok {
		return nil, ErrNotFound
	}

	oldBody := c.Body
	oldCategory := c.ReviewCategory
	oldSeverity := c.Severity
	oldUpdatedAt := c.UpdatedAt

	c.Body = fields.Body
	c.ReviewCategory = fields.ReviewCategory
	c.Severity = fields.Severity
	c.UpdatedAt = time.Now()

	if err := s.saveLocked(); err != nil {
		c.Body = oldBody
		c.ReviewCategory = oldCategory
		c.Severity = oldSeverity
		c.UpdatedAt = oldUpdatedAt
		return nil, fmt.Errorf("persisting comment: %w", err)
	}

	copy := *c
	return &copy, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.comments[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.comments, id)

	if err := s.saveLocked(); err != nil {
		s.comments[id] = c
		return fmt.Errorf("persisting deletion: %w", err)
	}
	return nil
}

// DeleteAll removes all comments and resets the ID counter.
func (s *Store) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldComments := s.comments
	oldNextID := s.nextID

	s.comments = make(map[string]*Comment)
	s.nextID = 1

	if err := s.saveLocked(); err != nil {
		s.comments = oldComments
		s.nextID = oldNextID
		return fmt.Errorf("persisting delete all: %w", err)
	}
	return nil
}

// Save persists all comments to disk. Thread-safe.
// Uses a write lock to prevent concurrent filesystem writes.
// Note: CUD operations (Create/Update/Delete) auto-persist, so explicit Save()
// is only needed after Load() or for external callers.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// saveLocked persists comments to disk. Must be called with s.mu write lock held.
func (s *Store) saveLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.commentSlice(), "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename to prevent corruption on crash
	tmp, err := os.CreateTemp(dir, "comments-*.tmp")
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

	var comments []*Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return err
	}

	s.comments = make(map[string]*Comment)
	maxID := 0
	for _, c := range comments {
		if c.ID == "" || c.FilePath == "" || c.Line < 0 {
			continue // skip invalid comments
		}
		s.comments[c.ID] = c
		var n int
		if _, err := fmt.Sscanf(c.ID, "c%d", &n); err == nil && n > maxID {
			maxID = n
		}
	}
	s.nextID = maxID + 1
	return nil
}

func (s *Store) commentSlice() []*Comment {
	result := make([]*Comment, 0, len(s.comments))
	for _, c := range s.comments {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}
