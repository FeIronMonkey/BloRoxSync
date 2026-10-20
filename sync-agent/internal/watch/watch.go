package watch

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	Debounce        = 50 * time.Millisecond
	SelfWriteIgnore = 500 * time.Millisecond
)

type Event struct {
	Path string
}

type Watcher struct {
	events      chan Event
	fw          *fsnotify.Watcher
	mu          sync.Mutex
	pending     map[string]*time.Timer
	ignoreUntil map[string]time.Time
	closed      bool
}

func Start(ctx context.Context, roots []string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		events:      make(chan Event, 64),
		fw:          fw,
		pending:     map[string]*time.Timer{},
		ignoreUntil: map[string]time.Time{},
	}
	for _, root := range roots {
		if err := os.MkdirAll(root, 0o755); err != nil {
			_ = fw.Close()
			return nil, err
		}
		if err := w.addRecursive(root); err != nil {
			_ = fw.Close()
			return nil, err
		}
	}
	go w.loop(ctx)
	return w, nil
}

func (w *Watcher) Events() <-chan Event {
	return w.events
}

func (w *Watcher) NoteWrite(path string) {
	p := filepath.Clean(path)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ignoreUntil[p] = time.Now().Add(SelfWriteIgnore)
	if t, ok := w.pending[p]; ok {
		t.Stop()
		delete(w.pending, p)
	}
}

func (w *Watcher) Close() error {
	if w.fw == nil {
		return nil
	}
	return w.fw.Close()
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.fw.Add(path)
		}
		return nil
	})
}

func (w *Watcher) loop(ctx context.Context) {
	defer func() {
		_ = w.fw.Close()
		w.mu.Lock()
		if !w.closed {
			close(w.events)
			w.closed = true
		}
		for _, t := range w.pending {
			t.Stop()
		}
		w.mu.Unlock()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-w.fw.Errors:
			if !ok {
				return
			}
			_ = err
		case ev, ok := <-w.fw.Events:
			if !ok {
				return
			}
			w.handle(ev)
		}
	}
}

func (w *Watcher) handle(ev fsnotify.Event) {
	path := filepath.Clean(ev.Name)
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".rbx-sync-") {
		return
	}
	if ev.Has(fsnotify.Create) {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			_ = w.fw.Add(path)
		}
	}
	if !ev.Has(fsnotify.Write) && !ev.Has(fsnotify.Create) && !ev.Has(fsnotify.Remove) && !ev.Has(fsnotify.Rename) {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if until, ok := w.ignoreUntil[path]; ok && time.Now().Before(until) {
		return
	}
	if t, ok := w.pending[path]; ok {
		t.Stop()
	}
	w.pending[path] = time.AfterFunc(Debounce, func() {
		w.emit(path)
	})
}

func (w *Watcher) emit(path string) {
	w.mu.Lock()
	delete(w.pending, path)
	if w.closed {
		w.mu.Unlock()
		return
	}
	if until, ok := w.ignoreUntil[path]; ok && time.Now().Before(until) {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()
	select {
	case w.events <- Event{Path: path}:
	default:
	}
}
