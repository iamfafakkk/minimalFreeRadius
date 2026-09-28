package radiox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	l := parseLine("Mon Sep 28 13:48:47 2026 : Info: Ready to process requests")
	if l.Time != "Mon Sep 28 13:48:47 2026" || l.Level != "Info" || l.Message != "Ready to process requests" {
		t.Fatalf("unexpected parse: %+v", l)
	}
	e := parseLine("Mon Sep 28 13:48:47 2026 : Error: Boom")
	if e.Level != "Error" || e.Message != "Boom" {
		t.Fatalf("unexpected error parse: %+v", e)
	}
	if e.Type != "SYS" || e.Status != "sys" {
		t.Fatalf("error line should classify as SYS: %+v", e)
	}
}

// TestClassifyAuth covers the heuristic that turns plain Auth lines into the
// structured fields rendered by the Radius Logs table.
func TestClassifyAuth(t *testing.T) {
	ok := parseLine("Mon Sep 28 13:48:47 2026 : Auth: Login OK: [budi] (from client mikrotik-bandar port 0)")
	if ok.Type != "AUTH_OK" || ok.Status != "ok" || ok.User != "budi" {
		t.Fatalf("unexpected auth ok: %+v", ok)
	}
	fail := parseLine("Mon Sep 28 13:48:47 2026 : Auth: Login incorrect: [nisa] (from client 10.10.0.5)")
	if fail.Type != "AUTH_FAIL" || fail.Status != "fail" || fail.User != "nisa" || fail.NAS != "10.10.0.5" {
		t.Fatalf("unexpected auth fail: %+v", fail)
	}
}

// TestTailerPublishesNewLines appends a line and expects it on the hub, while
// the pre-existing backlog is available via Recent().
func TestTailerPublishesNewLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "radius.log")
	if err := os.WriteFile(path, []byte("seed : Info: old line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hub := NewHub()
	tl := NewTailer(path, hub, 50)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go tl.Run(ctx)

	// Wait for the initial backlog to be loaded (it is *not* published).
	deadline := time.Now().Add(2 * time.Second)
	for len(tl.Recent()) == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if len(tl.Recent()) != 1 {
		t.Fatalf("backlog not loaded: %+v", tl.Recent())
	}

	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString("Mon Sep 28 14:00:00 2026 : Error: new line\n")
	f.Close()

	select {
	case got := <-ch:
		if got.Level != "Error" || got.Message != "new line" {
			t.Fatalf("unexpected published line: %+v", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for published line")
	}

	if len(tl.Recent()) != 2 {
		t.Fatalf("want 2 buffered lines, got %d", len(tl.Recent()))
	}
}
