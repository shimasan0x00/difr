package claude

import (
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
