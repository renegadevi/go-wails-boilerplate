package applog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestConsoleHandlerUses24HourClock(t *testing.T) {
	var buf bytes.Buffer
	h := newConsoleHandler(&buf, true)
	rec := slog.NewRecord(time.Date(2026, 10, 30, 16, 52, 40, 0, time.UTC), slog.LevelInfo, "Build Info:", 0)
	rec.Add("Wails", "v3.0.0")
	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatal(err)
	}
	const want = "16:52:40 INF Build Info: Wails=v3.0.0\n"
	if got := buf.String(); got != want {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(buf.String(), "PM") || strings.Contains(buf.String(), "AM") {
		t.Fatalf("clock should be 24-hour, got %q", buf.String())
	}
}

func TestConsoleHandlerSkipsDebug(t *testing.T) {
	h := newConsoleHandler(&bytes.Buffer{}, true)
	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("debug is below the Wails system logger")
	}
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info should be logged")
	}
}
