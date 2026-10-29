package applog

import (
	"os"
	"path/filepath"

	log "github.com/renegadevi/prezerolog"
)

// Pre sends process logs through prezerolog. Device changes do not go through here.
type Pre struct{}

// Init prepares console and rotated file logging. Levels come from the environment, with development defaults.
// The log folder is absolute. Finder starts a Mac app with a working directory of /, and a relative "logs" folder panics there.
func Init() {
	dir := logDirectory()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		dir = filepath.Join(os.TempDir(), "boilerplate", "logs")
		_ = os.MkdirAll(dir, 0o755)
	}
	_ = os.Setenv("LOG_DIR", dir)
	defaults := map[string]string{
		"LOG_NAME":           "boilerplate",
		"LOG_ENV":            "development",
		"LOG_FILE":           "true",
		"LOG_CONSOLE":        "true",
		"LOG_CONSOLE_LEVEL":  "info",
		"LOG_FILE_LEVEL":     "debug",
		"LOG_CONSOLE_OUTPUT": "full",
	}
	for key, value := range defaults {
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	log.InitLogging()
}

// logDirectory is where process logs are written. An absolute LOG_DIR is kept. Anything else goes next to the database.
func logDirectory() string {
	if dir := os.Getenv("LOG_DIR"); dir != "" && filepath.IsAbs(dir) {
		return dir
	}
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "boilerplate", "logs")
}

func (Pre) Debug(msg string, fields map[string]any) {
	if len(fields) == 0 {
		log.Debug(msg)
		return
	}
	log.Debug(msg, fields)
}

func (Pre) Info(msg string, fields map[string]any) {
	if len(fields) == 0 {
		log.Info(msg)
		return
	}
	log.Info(msg, fields)
}

func (Pre) Warn(msg string, fields map[string]any) {
	if len(fields) == 0 {
		log.Warn(msg)
		return
	}
	log.Warn(msg, fields)
}

func (Pre) Error(msg string, err error) {
	if err == nil {
		log.Error(msg)
		return
	}
	log.Error(msg, err)
}

// Text is the current process log file. Export uses this. The window tail is shorter.
func Text() string {
	if log.AppLogger == nil {
		return ""
	}
	body, err := os.ReadFile(log.AppLogger.GetCurrentLogFile())
	if err != nil {
		return ""
	}
	return string(body)
}

// Tail returns the end of the current process log. The change-log export does not include it.
func Tail() string {
	if log.AppLogger == nil {
		return ""
	}
	path := log.AppLogger.GetCurrentLogFile()
	body, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := string(body)
	const max = 200
	lines := splitLines(text)
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += "\n"
		}
		out += line
	}
	return out
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}
