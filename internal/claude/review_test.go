package claude

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReviewComments_ExtractsFromMarkdownWrappedJSON(t *testing.T) {
	data, err := os.ReadFile("testdata/review_response.json")
	require.NoError(t, err)

	comments := ParseReviewComments(string(data))

	require.Len(t, comments, 2)

	assert.Equal(t, "main.go", comments[0].FilePath)
	assert.Equal(t, 15, comments[0].Line)
	assert.NotEmpty(t, comments[0].Body)

	assert.Equal(t, "utils.go", comments[1].FilePath)
}

func TestParseReviewComments_ParsesPureJSONArray(t *testing.T) {
	input := `[{"filePath":"a.go","line":1,"body":"fix"}]`

	comments := ParseReviewComments(input)

	require.Len(t, comments, 1)
	assert.Equal(t, "fix", comments[0].Body)
}

func TestParseReviewComments_ReturnsEmptyForNoJSON(t *testing.T) {
	input := "This code looks great, no issues found."

	comments := ParseReviewComments(input)

	assert.Empty(t, comments)
}

func TestParseReviewComments_ReturnsEmptyForInvalidJSON(t *testing.T) {
	input := "```json\n[{invalid}]\n```"

	comments := ParseReviewComments(input)

	assert.Empty(t, comments)
}

func TestExtractJSONArray_SkipsPastNonMatchingArrays(t *testing.T) {
	// First array has no filePath, second one does.
	// Ensures the scanner advances past the first array without rescanning.
	input := `Some text [1, 2, 3] and then [{"filePath":"a.go","line":1,"body":"ok"}]`

	comments := ParseReviewComments(input)

	require.Len(t, comments, 1)
	assert.Equal(t, "a.go", comments[0].FilePath)
}

func TestExtractJSONArray_HandlesNestedArrays(t *testing.T) {
	input := `[[1,2],[3,4]] then [{"filePath":"b.go","line":5,"body":"nested"}]`

	comments := ParseReviewComments(input)

	require.Len(t, comments, 1)
	assert.Equal(t, "b.go", comments[0].FilePath)
}
