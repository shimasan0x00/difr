package diff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readTestData(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return string(data)
}

func TestParse_SimpleAdd(t *testing.T) {
	raw := readTestData(t, "simple_add.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	require.Len(t, result.Files, 1)
	f := result.Files[0]
	assert.Equal(t, "hello.go", f.NewPath)
	assert.Equal(t, FileStatusModified, f.Status)
	assert.Equal(t, "go", f.Language)
	assert.False(t, f.IsBinary)

	require.Len(t, f.Hunks, 1)
	hunk := f.Hunks[0]
	assert.Equal(t, 1, hunk.OldStart)
	assert.Equal(t, 1, hunk.NewStart)

	// Count additions and deletions
	var adds, dels int
	for _, line := range hunk.Lines {
		switch line.Type {
		case LineAdd:
			adds++
		case LineDelete:
			dels++
		}
	}
	assert.Equal(t, 3, adds, "should have 3 additions")
	assert.Equal(t, 0, dels, "should have 0 deletions")
	assert.Equal(t, 3, f.Stats.Additions)
	assert.Equal(t, 0, f.Stats.Deletions)
}

func TestParse_MultiFile(t *testing.T) {
	raw := readTestData(t, "multi_file.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	require.Len(t, result.Files, 2)

	// First file: modified
	f0 := result.Files[0]
	assert.Equal(t, "main.go", f0.NewPath)
	assert.Equal(t, FileStatusModified, f0.Status)
	assert.Equal(t, 1, f0.Stats.Additions)
	assert.Equal(t, 1, f0.Stats.Deletions)

	// Second file: added
	f1 := result.Files[1]
	assert.Equal(t, "utils.go", f1.NewPath)
	assert.Equal(t, FileStatusAdded, f1.Status)
	assert.Equal(t, 3, f1.Stats.Additions)
	assert.Equal(t, 0, f1.Stats.Deletions)

	// Total stats
	assert.Equal(t, 4, result.Stats.Additions)
	assert.Equal(t, 1, result.Stats.Deletions)
}

func TestParse_DeletedFile(t *testing.T) {
	raw := readTestData(t, "delete.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	require.Len(t, result.Files, 1)
	f := result.Files[0]
	assert.Equal(t, "removed.go", f.OldPath)
	assert.Equal(t, FileStatusDeleted, f.Status)
	assert.Equal(t, 0, f.Stats.Additions)
	assert.Equal(t, 3, f.Stats.Deletions)
}

func TestParse_Rename(t *testing.T) {
	raw := readTestData(t, "rename.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	require.Len(t, result.Files, 1)
	f := result.Files[0]
	assert.Equal(t, "old_name.go", f.OldPath)
	assert.Equal(t, "new_name.go", f.NewPath)
	assert.Equal(t, FileStatusRenamed, f.Status)
}

func TestParse_Binary(t *testing.T) {
	raw := readTestData(t, "binary.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	require.Len(t, result.Files, 1)
	f := result.Files[0]
	assert.True(t, f.IsBinary)
	assert.Equal(t, FileStatusAdded, f.Status)
}

func TestParse_EmptyInput(t *testing.T) {
	result, err := Parse("")
	require.NoError(t, err)
	assert.Empty(t, result.Files)
	assert.Equal(t, 0, result.Stats.Additions)
	assert.Equal(t, 0, result.Stats.Deletions)
}

func TestParse_LineNumbers(t *testing.T) {
	raw := readTestData(t, "simple_add.diff")
	result, err := Parse(raw)
	require.NoError(t, err)

	hunk := result.Files[0].Hunks[0]
	// Verify line numbers are populated
	for _, line := range hunk.Lines {
		switch line.Type {
		case LineContext:
			assert.Greater(t, line.OldNumber, 0, "context line should have old number")
			assert.Greater(t, line.NewNumber, 0, "context line should have new number")
		case LineAdd:
			assert.Greater(t, line.NewNumber, 0, "add line should have new number")
			assert.Equal(t, 0, line.OldNumber, "add line should not have old number")
		case LineDelete:
			assert.Greater(t, line.OldNumber, 0, "delete line should have old number")
			assert.Equal(t, 0, line.NewNumber, "delete line should not have new number")
		}
	}
}
