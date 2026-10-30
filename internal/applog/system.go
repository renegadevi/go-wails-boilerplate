package applog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// consoleTime matches the prezerolog console clock: local, 24-hour, with seconds.
const consoleTime = "15:04:05"

// System is the Wails system logger. Dev builds use the app log's clock.
// Production returns nil so Wails keeps its silent logger.
func System() *slog.Logger {
	if !systemLogs {
		return nil
	}
	noColor := !isatty.IsTerminal(os.Stderr.Fd())
	return slog.New(newConsoleHandler(colorable.NewColorable(os.Stderr), noColor))
}

type consoleHandler struct {
	mu      *sync.Mutex
	w       io.Writer
	noColor bool
	prefix  string
	attrs   []slog.Attr
}

func newConsoleHandler(w io.Writer, noColor bool) *consoleHandler {
	return &consoleHandler{mu: &sync.Mutex{}, w: w, noColor: noColor}
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer
	if !r.Time.IsZero() {
		h.writePart(&buf, colorGray, r.Time.Format(consoleTime))
		buf.WriteByte(' ')
	}
	label, color := consoleLevel(r.Level)
	h.writePart(&buf, color, label)
	buf.WriteByte(' ')
	h.writePart(&buf, colorWhite, r.Message)
	for _, a := range h.attrs {
		h.writeAttr(&buf, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&buf, a)
		return true
	})
	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := h.clone()
	next.attrs = append(next.attrs, attrs...)
	return next
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	next := h.clone()
	if name == "" {
		return next
	}
	if next.prefix == "" {
		next.prefix = name
		return next
	}
	next.prefix = next.prefix + "." + name
	return next
}

func (h *consoleHandler) clone() *consoleHandler {
	next := *h
	next.attrs = append([]slog.Attr(nil), h.attrs...)
	return &next
}

func (h *consoleHandler) writeAttr(buf *bytes.Buffer, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		for _, g := range a.Value.Group() {
			key := g.Key
			if a.Key != "" {
				key = a.Key + "." + key
			}
			h.writeAttr(buf, slog.Attr{Key: key, Value: g.Value})
		}
		return
	}
	key := a.Key
	if h.prefix != "" {
		key = h.prefix + "." + key
	}
	buf.WriteByte(' ')
	h.writePart(buf, colorCyan, key)
	buf.WriteByte('=')
	buf.WriteString(a.Value.String())
}

func (h *consoleHandler) writePart(buf *bytes.Buffer, color, text string) {
	if h.noColor || color == "" {
		buf.WriteString(text)
		return
	}
	buf.WriteString(color)
	buf.WriteString(text)
	buf.WriteString(colorReset)
}

func consoleLevel(level slog.Level) (string, string) {
	switch {
	case level < slog.LevelInfo:
		return "DBG", colorCyan
	case level < slog.LevelWarn:
		return "INF", colorGreen
	case level < slog.LevelError:
		return "WRN", colorYellow
	default:
		return "ERR", colorRed
	}
}

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[38;5;245m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[97m"
)
