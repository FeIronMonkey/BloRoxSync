package watch_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"roblox-sync/sync-agent/internal/watch"
)

func waitEvent(t *testing.T, ch <-chan watch.Event, d time.Duration) (watch.Event, bool) {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			t.Fatal("events channel closed")
		}
		return ev, true
	case <-time.After(d):
		return watch.Event{}, false
	}
}

func assertNoEvent(t *testing.T, ch <-chan watch.Event, d time.Duration) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("unexpected watch event %+v", ev)
	case <-time.After(d):
	}
}

func startWatch(t *testing.T, roots ...string) *watch.Watcher {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	w, err := watch.Start(ctx, roots)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })
	return w
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestWatchCoalescesWrites_WhenTwoWritesWithinDebounce(t *testing.T) {
	dir := t.TempDir()
	w := startWatch(t, dir)
	path := filepath.Join(dir, "Foo.luau")
	if err := os.WriteFile(path, []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("print(2)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ev, ok := waitEvent(t, w.Events(), 400*time.Millisecond)
	if !ok {
		t.Fatal("expected one coalesced event")
	}
	if filepath.Clean(ev.Path) != filepath.Clean(path) {
		t.Fatalf("path %q, want %q", ev.Path, path)
	}
	assertNoEvent(t, w.Events(), 150*time.Millisecond)
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#echo-suppression
func TestWatchIgnoresPath_WhenNoteWriteWithin500ms(t *testing.T) {
	dir := t.TempDir()
	w := startWatch(t, dir)
	path := filepath.Join(dir, "Foo.luau")
	w.NoteWrite(path)
	if err := os.WriteFile(path, []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertNoEvent(t, w.Events(), 200*time.Millisecond)
	time.Sleep(watch.SelfWriteIgnore)
	if err := os.WriteFile(path, []byte("print(2)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ev, ok := waitEvent(t, w.Events(), 400*time.Millisecond)
	if !ok {
		t.Fatal("expected event after ignore window")
	}
	if filepath.Clean(ev.Path) != filepath.Clean(path) {
		t.Fatalf("path %q, want %q", ev.Path, path)
	}
}

// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestWatchEmitsNothing_WhenIdle(t *testing.T) {
	dir := t.TempDir()
	w := startWatch(t, dir)
	assertNoEvent(t, w.Events(), 200*time.Millisecond)
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#rbx-sync-dependencies
func TestWatchGoModRequiresFsnotify_WhenModuleExists(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "github.com/fsnotify/fsnotify") {
		t.Fatal("sync-agent/go.mod must require github.com/fsnotify/fsnotify")
	}
}
