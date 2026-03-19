package comment

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_AssignsIDAndTimestamp(t *testing.T) {
	store := newTestStore(t)

	created, err := store.Create(&Comment{
		FilePath: "main.go",
		Line:     10,
		Body:     "This needs refactoring",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "main.go", created.FilePath)
	assert.Equal(t, 10, created.Line)
	assert.False(t, created.CreatedAt.IsZero(), "CreatedAt should be populated")
}

func TestGet_ReturnsCreatedComment(t *testing.T) {
	store := newTestStore(t)
	created, err := store.Create(&Comment{FilePath: "main.go", Line: 10, Body: "This needs refactoring"})
	require.NoError(t, err)

	got, err := store.Get(created.ID)

	require.NoError(t, err)
	assert.Equal(t, "This needs refactoring", got.Body)
	assert.Equal(t, created.ID, got.ID)
}

func TestGet_ReturnsErrNotFoundForNonexistent(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Get("nonexistent")

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList_ReturnsAllComments(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Create(&Comment{FilePath: "a.go", Line: 1, Body: "first"})
	require.NoError(t, err)
	_, err = store.Create(&Comment{FilePath: "b.go", Line: 2, Body: "second"})
	require.NoError(t, err)
	_, err = store.Create(&Comment{FilePath: "a.go", Line: 5, Body: "third"})
	require.NoError(t, err)

	all := store.List("")

	assert.Len(t, all, 3)
}

func TestList_FiltersByFilePath(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Create(&Comment{FilePath: "a.go", Line: 1, Body: "first"})
	require.NoError(t, err)
	_, err = store.Create(&Comment{FilePath: "b.go", Line: 2, Body: "second"})
	require.NoError(t, err)
	_, err = store.Create(&Comment{FilePath: "a.go", Line: 5, Body: "third"})
	require.NoError(t, err)

	filtered := store.List("a.go")

	assert.Len(t, filtered, 2)
	for _, c := range filtered {
		assert.Equal(t, "a.go", c.FilePath)
	}
}

func TestUpdate_ChangesBodyAndSetsUpdatedAt(t *testing.T) {
	store := newTestStore(t)
	created, err := store.Create(&Comment{FilePath: "main.go", Line: 1, Body: "old"})
	require.NoError(t, err)

	updated, err := store.Update(created.ID, UpdateFields{Body: "new body"})

	require.NoError(t, err)
	assert.Equal(t, "new body", updated.Body)
	assert.False(t, updated.UpdatedAt.IsZero(), "UpdatedAt should be set after update")
}

func TestUpdate_ChangesAllFields(t *testing.T) {
	store := newTestStore(t)
	created, err := store.Create(&Comment{FilePath: "main.go", Line: 1, Body: "old", ReviewCategory: "FYI", Severity: "Low"})
	require.NoError(t, err)

	updated, err := store.Update(created.ID, UpdateFields{Body: "new", ReviewCategory: "MUST", Severity: "Critical"})

	require.NoError(t, err)
	assert.Equal(t, "new", updated.Body)
	assert.Equal(t, "MUST", updated.ReviewCategory)
	assert.Equal(t, "Critical", updated.Severity)
}

func TestUpdate_ReturnsErrNotFoundForNonexistent(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Update("nonexistent", UpdateFields{Body: "body"})

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_RemovesComment(t *testing.T) {
	store := newTestStore(t)
	created, err := store.Create(&Comment{FilePath: "main.go", Line: 1, Body: "delete me"})
	require.NoError(t, err)

	err = store.Delete(created.ID)

	require.NoError(t, err)
	_, err = store.Get(created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_ReturnsErrNotFoundForNonexistent(t *testing.T) {
	store := newTestStore(t)

	err := store.Delete("nonexistent")

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPersistence_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.json")

	// Arrange: Create (auto-saves)
	store1 := NewStore(path)
	_, err := store1.Create(&Comment{FilePath: "main.go", Line: 1, Body: "persistent"})
	require.NoError(t, err)

	// Act: Load in a new store
	store2 := NewStore(path)
	require.NoError(t, store2.Load())

	// Assert
	all := store2.List("")
	require.Len(t, all, 1)
	assert.Equal(t, "persistent", all[0].Body)
}

func TestLoad_NonExistentFileReturnsEmptyStore(t *testing.T) {
	store := NewStore("/tmp/nonexistent-difr-comments.json")

	err := store.Load()

	assert.NoError(t, err)
	assert.Empty(t, store.List(""))
}

func TestConcurrentAccess_AllWritesSucceed(t *testing.T) {
	store := newTestStore(t)

	errs := make(chan error, 50)
	var wg sync.WaitGroup
	const goroutines = 50
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Create(&Comment{FilePath: "main.go", Line: 1, Body: "concurrent"})
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	all := store.List("")
	assert.Len(t, all, goroutines)
}

func TestConcurrentUpdateAndDelete_AllOperationsSucceedOrReturnErrNotFound(t *testing.T) {
	store := newTestStore(t)

	// Arrange: Create comments to operate on
	const numComments = 20
	ids := make([]string, numComments)
	for i := 0; i < numComments; i++ {
		c, err := store.Create(&Comment{FilePath: "main.go", Line: i + 1, Body: "original"})
		require.NoError(t, err)
		ids[i] = c.ID
	}

	// Act: Concurrent updates and deletes
	var wg sync.WaitGroup
	const goroutines = 100
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := ids[i%numComments]
			if i%2 == 0 {
				_, err := store.Update(id, UpdateFields{Body: "updated"})
				if err != nil && err != ErrNotFound {
					errs <- err
				}
			} else {
				err := store.Delete(id)
				if err != nil && err != ErrNotFound {
					errs <- err
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	// Assert: No unexpected errors
	for err := range errs {
		require.NoError(t, err)
	}

	// Verify store is in a consistent state: all remaining comments are accessible
	all := store.List("")
	for _, c := range all {
		got, err := store.Get(c.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, got.Body)
	}
}

func TestSave_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "comments.json")

	store := NewStore(path)
	_, err := store.Create(&Comment{FilePath: "a.go", Line: 1, Body: "test"})
	require.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err, "file should exist after save")
}

func TestGet_ReturnedPointerDoesNotMutateStore(t *testing.T) {
	store := newTestStore(t)
	created, err := store.Create(&Comment{FilePath: "main.go", Line: 10, Body: "original"})
	require.NoError(t, err)

	// Act: Mutate the returned pointer
	got, err := store.Get(created.ID)
	require.NoError(t, err)
	got.Body = "mutated externally"

	// Assert: Store's internal state is unchanged
	got2, err := store.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "original", got2.Body)
}

func TestDeleteAll_RemovesAllCommentsAndResetsID(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Create(&Comment{FilePath: "a.go", Line: 1, Body: "first"})
	require.NoError(t, err)
	_, err = store.Create(&Comment{FilePath: "b.go", Line: 2, Body: "second"})
	require.NoError(t, err)

	require.NoError(t, store.DeleteAll())

	assert.Empty(t, store.List(""))

	// After DeleteAll, next created comment should get c1 again
	created, err := store.Create(&Comment{FilePath: "c.go", Line: 1, Body: "new"})
	require.NoError(t, err)
	assert.Equal(t, "c1", created.ID)
}

func TestCreate_WithReviewCategoryAndSeverity(t *testing.T) {
	store := newTestStore(t)

	created, err := store.Create(&Comment{
		FilePath:       "main.go",
		Line:           10,
		Body:           "Fix this",
		ReviewCategory: "MUST",
		Severity:       "Critical",
	})

	require.NoError(t, err)
	assert.Equal(t, "MUST", created.ReviewCategory)
	assert.Equal(t, "Critical", created.Severity)

	got, err := store.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "MUST", got.ReviewCategory)
	assert.Equal(t, "Critical", got.Severity)
}

func TestCreate_WithEmptyCategoryAndSeverity(t *testing.T) {
	store := newTestStore(t)

	created, err := store.Create(&Comment{
		FilePath: "main.go",
		Line:     10,
		Body:     "No category",
	})

	require.NoError(t, err)
	assert.Empty(t, created.ReviewCategory)
	assert.Empty(t, created.Severity)
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"", true},
		{"MUST", true},
		{"IMO", true},
		{"Q", true},
		{"FYI", true},
		{"invalid", false},
		{"must", false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.valid, ValidateCategory(tc.input), "ValidateCategory(%q)", tc.input)
	}
}

func TestValidateSeverity(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"", true},
		{"Critical", true},
		{"High", true},
		{"Middle", true},
		{"Low", true},
		{"invalid", false},
		{"critical", false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.valid, ValidateSeverity(tc.input), "ValidateSeverity(%q)", tc.input)
	}
}

func TestLoad_BackwardCompatibility_WithoutNewFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.json")

	// Write JSON without reviewCategory/severity fields
	data := `[{"id":"c1","filePath":"main.go","line":10,"body":"old comment","createdAt":"2026-01-01T00:00:00Z"}]`
	require.NoError(t, os.WriteFile(path, []byte(data), 0o644))

	store := NewStore(path)
	require.NoError(t, store.Load())

	all := store.List("")
	require.Len(t, all, 1)
	assert.Equal(t, "old comment", all[0].Body)
	assert.Empty(t, all[0].ReviewCategory)
	assert.Empty(t, all[0].Severity)
}

func TestPersistence_WithReviewCategoryAndSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.json")

	store1 := NewStore(path)
	_, err := store1.Create(&Comment{
		FilePath:       "main.go",
		Line:           1,
		Body:           "important",
		ReviewCategory: "MUST",
		Severity:       "High",
	})
	require.NoError(t, err)

	store2 := NewStore(path)
	require.NoError(t, store2.Load())

	all := store2.List("")
	require.Len(t, all, 1)
	assert.Equal(t, "MUST", all[0].ReviewCategory)
	assert.Equal(t, "High", all[0].Severity)
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.json")
	return NewStore(path)
}
