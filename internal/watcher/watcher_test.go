package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFileWithRetry writes a file and retries if the watcher hasn't picked it up,
// to handle the race between watcher setup and file operations.
func writeFileWithRetry(t *testing.T, path string, content []byte, events <-chan Event, timeout time.Duration) Event {
	t.Helper()
	deadline := time.After(timeout)
	require.NoError(t, os.WriteFile(path, content, 0o644))

	for {
		select {
		case ev := <-events:
			return ev
		case <-deadline:
			t.Fatal("timed out waiting for file change event")
			return Event{} // unreachable
		}
	}
}

func TestWatcher_DetectsFileModification(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0o644))

	w, err := New(dir)
	require.NoError(t, err)
	defer w.Close()

	ev := writeFileWithRetry(t, testFile, []byte("package main\nfunc main() {}"), w.Events(), 5*time.Second)
	assert.NotEmpty(t, ev.Path, "event should have a file path")
}

func TestWatcher_DetectsNewFileCreation(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	require.NoError(t, err)
	defer w.Close()

	ev := writeFileWithRetry(t, filepath.Join(dir, "new.go"), []byte("package main"), w.Events(), 5*time.Second)
	assert.NotEmpty(t, ev.Path)
}

func TestWatcher_DebouncesRapidChangesIntoSingleEvent(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("v1"), 0o644))

	const debounce = 200 * time.Millisecond
	w, err := NewWithDebounce(dir, debounce)
	require.NoError(t, err)
	defer w.Close()

	// Write a sync file and wait for the event to confirm the watcher is ready
	syncFile := filepath.Join(dir, "sync.go")
	writeFileWithRetry(t, syncFile, []byte("sync"), w.Events(), 5*time.Second)

	// Now perform rapid changes within the debounce window
	for i := 0; i < 5; i++ {
		require.NoError(t, os.WriteFile(testFile, []byte("v"+string(rune('2'+i))), 0o644))
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for debounce to settle, then collect events
	count := 0
	timeout := time.After(debounce * 5)
	for {
		select {
		case <-w.Events():
			count++
		case <-timeout:
			goto done
		}
	}
done:
	assert.Equal(t, 1, count, "rapid changes should be debounced into a single event")
}

func TestWatcher_CloseReturnsNoError(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	require.NoError(t, err)

	assert.NoError(t, w.Close())
}

func TestWatcher_EventsChannelClosesAfterClose(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	require.NoError(t, err)

	require.NoError(t, w.Close())

	// Events channel should be closed after Close()
	select {
	case _, ok := <-w.Events():
		assert.False(t, ok, "events channel should be closed after Close()")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for events channel to close")
	}
}
