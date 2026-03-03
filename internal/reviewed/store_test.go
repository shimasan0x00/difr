package reviewed_test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/shimasan0x00/difr/internal/reviewed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *reviewed.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "reviewed-files.json")
	return reviewed.NewStore(path)
}

func newTestStoreAt(t *testing.T, path string) *reviewed.Store {
	t.Helper()
	return reviewed.NewStore(path)
}

func TestAddAndHas(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.Add("src/main.go"))

	assert.True(t, s.Has("src/main.go"))
	assert.False(t, s.Has("src/other.go"))
}

func TestRemove(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.Add("a.go"))
	require.NoError(t, s.Remove("a.go"))

	assert.False(t, s.Has("a.go"))
}

func TestRemoveNonExistent(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.Remove("nonexistent.go"))
}

func TestList(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.Add("c.go"))
	require.NoError(t, s.Add("a.go"))
	require.NoError(t, s.Add("b.go"))

	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, s.List())
}

func TestListEmpty(t *testing.T) {
	s := newTestStore(t)

	assert.Empty(t, s.List())
}

func TestClear(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, s.Add("a.go"))
	require.NoError(t, s.Add("b.go"))
	require.NoError(t, s.Clear())

	assert.Empty(t, s.List())
	assert.False(t, s.Has("a.go"))
}

func TestPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reviewed-files.json")

	s1 := newTestStoreAt(t, path)
	require.NoError(t, s1.Add("x.go"))
	require.NoError(t, s1.Add("y.go"))

	s2 := newTestStoreAt(t, path)
	require.NoError(t, s2.Load())

	assert.True(t, s2.Has("x.go"))
	assert.True(t, s2.Has("y.go"))
	assert.Equal(t, []string{"x.go", "y.go"}, s2.List())
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "reviewed-files.json")
	s := newTestStoreAt(t, path)

	require.NoError(t, s.Load())
	assert.Empty(t, s.List())
}

func TestConcurrentAccess(t *testing.T) {
	s := newTestStore(t)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := filepath.Join("dir", string(rune('a'+i%26))+".go")
			_ = s.Add(path)
			s.Has(path)
			s.List()
			_ = s.Remove(path)
		}(i)
	}
	wg.Wait()
}
