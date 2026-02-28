package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd_DefaultFlags(t *testing.T) {
	cmd := NewRootCmd()
	require.NotNil(t, cmd)

	tests := []struct {
		flag     string
		expected interface{}
	}{
		{"port", 3333},
		{"host", "127.0.0.1"},
		{"mode", "split"},
		{"no-open", false},
		{"no-claude", false},
		{"watch", false},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case int:
				val, err := cmd.Flags().GetInt(tt.flag)
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			case string:
				val, err := cmd.Flags().GetString(tt.flag)
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			case bool:
				val, err := cmd.Flags().GetBool(tt.flag)
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			}
		})
	}
}

func TestNewRootCmd_PortBoundaryValidation(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"port 0 is invalid", "0", true},
		{"port 1 is valid (min)", "1", false},
		{"port 65535 is valid (max)", "65535", false},
		{"port 65536 is invalid", "65536", true},
		{"port -1 is invalid", "-1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			// Prevent actual server start: just validate flags
			cmd.SetArgs([]string{"--port", tt.port, "--no-open", "staged"})
			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "port")
			} else {
				// Valid port will fail because we don't have a git repo, which is expected
				// The important thing is it's NOT a port validation error
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid port")
				}
			}
		})
	}
}

func TestNewRootCmd_RejectsInvalidMode(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--mode", "invalid"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode")
}

func TestNewRootCmd_RejectsTooManyArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"a", "b", "c"})

	err := cmd.Execute()

	require.Error(t, err)
}
