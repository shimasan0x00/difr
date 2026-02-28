package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file change event.
type Event struct {
	Path string
	Op   string
}

// Watcher watches a directory for file changes with debouncing.
type Watcher struct {
	fsWatcher  *fsnotify.Watcher
	events     chan Event
	done       chan struct{}
	closeOnce  sync.Once
	debounce   time.Duration
	mu         sync.Mutex
	pending    *Event
}

const defaultDebounce = 100 * time.Millisecond

// New creates a new Watcher with default debounce.
func New(dir string) (*Watcher, error) {
	return NewWithDebounce(dir, defaultDebounce)
}

// NewWithDebounce creates a new Watcher with custom debounce duration.
// Recursively watches all subdirectories, skipping hidden dirs and common non-source dirs.
func NewWithDebounce(dir string, debounce time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := addRecursive(fsw, dir); err != nil {
		fsw.Close()
		return nil, err
	}

	w := &Watcher{
		fsWatcher: fsw,
		events:    make(chan Event, 1),
		done:      make(chan struct{}),
		debounce:  debounce,
	}

	go w.loop()
	return w, nil
}

// Events returns the channel of debounced file change events.
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsWatcher.Close()
}

// skipDirs contains directory names to skip during recursive watching.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	".difr":        true,
}

// addRecursive adds dir and all subdirectories to the watcher,
// skipping hidden directories (starting with ".") and common non-source dirs.
func addRecursive(fsw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible directories
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		// Skip hidden dirs (except the root itself) and known non-source dirs
		if path != root && (strings.HasPrefix(name, ".") || skipDirs[name]) {
			return filepath.SkipDir
		}
		return fsw.Add(path)
	})
}

// closeEvents closes the events channel exactly once, safe for concurrent callers.
func (w *Watcher) closeEvents() {
	w.closeOnce.Do(func() {
		close(w.events)
	})
}

// trySendEvent sends an event to the events channel, returning false if the
// channel is closed (recovering from the panic on send to a closed channel).
func (w *Watcher) trySendEvent(ev Event) (sent bool) {
	defer func() {
		if r := recover(); r != nil {
			sent = false
		}
	}()
	select {
	case w.events <- ev:
		return true
	default:
		return false
	}
}

func (w *Watcher) loop() {
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
		w.closeEvents()
	}()

	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			w.mu.Lock()
			w.pending = &Event{
				Path: ev.Name,
				Op:   ev.Op.String(),
			}
			w.mu.Unlock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, func() {
				select {
				case <-w.done:
					return
				default:
				}
				w.mu.Lock()
				p := w.pending
				w.pending = nil
				w.mu.Unlock()
				if p != nil {
					w.trySendEvent(*p)
				}
			})
		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}
