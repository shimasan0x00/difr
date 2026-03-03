package claude

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLookPath struct {
	found bool
}

func (m *mockLookPath) LookPath(file string) (string, error) {
	if m.found {
		return "/usr/bin/claude", nil
	}
	return "", &lookPathError{file}
}

type lookPathError struct {
	name string
}

func (e *lookPathError) Error() string {
	return e.name + ": executable not found"
}

// Compile-time check: Client implements Runner interface.
var _ Runner = (*Client)(nil)

func TestNewClient_Available(t *testing.T) {
	client, err := NewClient(WithLookPath(&mockLookPath{found: true}))

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.True(t, client.Available())
}

func TestNewClient_ReturnsErrorWhenCLINotFound(t *testing.T) {
	_, err := NewClient(WithLookPath(&mockLookPath{found: false}))

	assert.Error(t, err)
}

func TestLimitedBuffer_TruncatesAtMax(t *testing.T) {
	lb := &limitedBuffer{max: 10}

	n, err := lb.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	n, err = lb.Write([]byte("world!!!"))
	require.NoError(t, err)
	assert.Equal(t, 5, n) // only 5 bytes remaining

	assert.Equal(t, 10, lb.Len())
	assert.Equal(t, "helloworld", lb.String())
}

func TestLimitedBuffer_DiscardsAfterFull(t *testing.T) {
	lb := &limitedBuffer{max: 5}

	lb.Write([]byte("12345"))
	n, err := lb.Write([]byte("extra"))

	require.NoError(t, err)
	assert.Equal(t, 5, n) // reports full write length
	assert.Equal(t, "12345", lb.String())
}

func TestLimitedBuffer_LargeWrite(t *testing.T) {
	lb := &limitedBuffer{max: maxStderrSize}
	large := strings.Repeat("x", maxStderrSize+1024)

	lb.Write([]byte(large))

	assert.Equal(t, maxStderrSize, lb.Len())
}
