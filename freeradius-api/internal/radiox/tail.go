// Package radiox streams the FreeRADIUS log file to the web panel. It keeps a
// small in-memory ring of recent lines (for late joiners) and fans new lines
// out to subscribers. Transport is Server-Sent Events (one-way, no external
// dependency); see sse.go.
package radiox

import (
	"context"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// LogLine is one parsed record from radius.log.
type LogLine struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Raw     string `json:"raw"`

	// Structured fields inferred from the message (best effort).
	Type   string `json:"type"`   // AUTH_OK | AUTH_FAIL | ACCT | SYS
	Status string `json:"status"` // ok | fail | acct | sys
	User   string `json:"user"`   // User-Name / account, empty when unknown
	NAS    string `json:"nas"`    // client IP from "(from client ...)" / "client <x>"
}

var (
	reUser   = regexp.MustCompile(`\[([^\[\]]+)\]`)
	reNASIP  = regexp.MustCompile(`\b(\d{1,3}(?:\.\d{1,3}){3})\b`)
	reClient = regexp.MustCompile(`client\s+([^\s)]+)`)
)

// parseLine splits "Mon Sep 28 13:48:47 2026 : Info: message" into fields.
// Lines that do not match the format are kept whole as the message.
func parseLine(raw string) LogLine {
	l := LogLine{Raw: raw}
	rest := raw
	if i := strings.Index(raw, " : "); i >= 0 {
		l.Time = raw[:i]
		rest = raw[i+3:]
	}

	// FreeRADIUS emits "Level: message" for most lines, but the Auth log
	// destination writes "Auth: Login OK: [...]" / "Auth: Login incorrect: [...]".
	if j := strings.Index(rest, ": "); j >= 0 {
		lvl := rest[:j]
		switch lvl {
		case "Info", "Warning", "Error", "Debug":
			l.Level = lvl
			l.Message = rest[j+2:]
		case "Auth":
			l.Level = "Info"
			l.Message = rest[j+2:]
			classify(&l, l.Message)
		default:
			l.Message = rest
		}
	} else {
		l.Message = rest
	}
	if l.Type == "" {
		classify(&l, l.Message)
	}
	return l
}

// classify derives Type/Status/User/NAS from a FreeRADIUS auth/accounting
// message. It is intentionally heuristic — radius.log is free-form.
func classify(l *LogLine, msg string) {
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "login ok"), strings.Contains(low, "access-accept"):
		l.Type, l.Status = "AUTH_OK", "ok"
	case strings.Contains(low, "login incorrect"), strings.Contains(low, "access-reject"),
		strings.Contains(low, "rejected"), strings.Contains(low, "authentication failed"):
		l.Type, l.Status = "AUTH_FAIL", "fail"
	case strings.HasPrefix(low, "started"), strings.HasPrefix(low, "stopped"),
		strings.Contains(low, "acct"):
		l.Type, l.Status = "ACCT", "acct"
	default:
		l.Type, l.Status = "SYS", "sys"
	}
	if m := reUser.FindStringSubmatch(msg); m != nil {
		l.User = strings.TrimSpace(m[1])
	}
	if m := reNASIP.FindStringSubmatch(msg); m != nil {
		l.NAS = m[1]
	} else if m := reClient.FindStringSubmatch(msg); m != nil {
		l.NAS = m[1]
	}
}

// Hub fans log lines out to SSE subscribers.
type Hub struct {
	mu   sync.Mutex
	subs map[chan LogLine]struct{}
}

func NewHub() *Hub { return &Hub{subs: map[chan LogLine]struct{}{}} }

func (h *Hub) Subscribe() chan LogLine {
	ch := make(chan LogLine, 128)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan LogLine) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// Publish is non-blocking: a slow subscriber drops lines rather than stalling
// the tailer.
func (h *Hub) Publish(l LogLine) {
	h.mu.Lock()
	for ch := range h.subs {
		select {
		case ch <- l:
		default:
		}
	}
	h.mu.Unlock()
}

// Tailer follows a log file, keeping a bounded ring of recent lines and
// publishing new ones. Rotation/truncation is handled by re-reading from the
// start when the file shrinks.
type Tailer struct {
	path string
	hub  *Hub

	mu   sync.Mutex
	ring []LogLine
	cap  int
}

func NewTailer(path string, hub *Hub, backlog int) *Tailer {
	if backlog < 1 {
		backlog = 200
	}
	return &Tailer{path: path, hub: hub, cap: backlog}
}

func (t *Tailer) add(l LogLine) {
	t.mu.Lock()
	t.ring = append(t.ring, l)
	if len(t.ring) > t.cap {
		t.ring = t.ring[len(t.ring)-t.cap:]
	}
	t.mu.Unlock()
}

// Recent returns a copy of the current ring (oldest first).
func (t *Tailer) Recent() []LogLine {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]LogLine, len(t.ring))
	copy(out, t.ring)
	return out
}

// Run tails the file until ctx is done. The first pass loads the last `cap`
// lines into the ring and starts streaming only from the end.
func (t *Tailer) Run(ctx context.Context) {
	var off int64
	var carry []byte
	var f *os.File

	openFromStart := func() {
		if f != nil {
			f.Close()
			f = nil
		}
		nf, err := os.Open(t.path)
		if err != nil {
			return
		}
		f = nf
		off = 0
	}

	// Initial load: read the whole file, keep the last N lines in the ring,
	// then leave the offset at EOF so we only emit new lines afterwards.
	if data, err := os.ReadFile(t.path); err == nil {
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		start := 0
		if len(lines) > t.cap {
			start = len(lines) - t.cap
		}
		for _, ln := range lines[start:] {
			if ln != "" {
				t.add(parseLine(ln))
			}
		}
		var size int64
		if fi, err := os.Stat(t.path); err == nil {
			size = fi.Size()
		}
		openFromStart() // resets off to 0
		if f != nil {
			if _, err := f.Seek(size, 0); err == nil {
				off = size
			}
		}
	}
	if f == nil {
		openFromStart()
	}

	// ponytail: 400ms polling instead of fsnotify — zero deps, handles
	// rotation/truncation trivially. Swap to fsnotify if latency matters.
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if f != nil {
				f.Close()
			}
			return
		case <-ticker.C:
		}

		if f == nil {
			openFromStart()
			if f == nil {
				continue
			}
		}
		fi, err := f.Stat()
		if err != nil {
			// File replaced/removed (log rotation): reopen.
			if f != nil {
				f.Close()
				f = nil
			}
			continue
		}
		if fi.Size() < off {
			// Truncated: restart from the beginning.
			if _, err := f.Seek(0, 0); err == nil {
				off = 0
				carry = nil
			}
		}
		if fi.Size() == off {
			continue
		}
		buf := make([]byte, fi.Size()-off)
		n, rerr := f.ReadAt(buf, off)
		if n <= 0 {
			if rerr != nil {
				f.Close()
				f = nil
			}
			continue
		}
		off += int64(n)
		carry = append(carry, buf[:n]...)
		for {
			i := indexByte(carry, '\n')
			if i < 0 {
				break
			}
			line := string(carry[:i])
			carry = carry[i+1:]
			if line == "" {
				continue
			}
			l := parseLine(line)
			t.add(l)
			t.hub.Publish(l)
		}
	}
}

func indexByte(b []byte, c byte) int {
	for i := range b {
		if b[i] == c {
			return i
		}
	}
	return -1
}
