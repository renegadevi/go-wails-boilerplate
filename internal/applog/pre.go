// Package applog sets up the process log and the Wails system logger.
package applog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	log "github.com/renegadevi/prezerolog"
)

// Pre sends process logs through prezerolog.
type Pre struct{}

// logDefaults fill any app log setting still empty after the process environment and .env.
// A shell export wins over .env. .env wins over these.
var logDefaults = map[string]string{
	"LOG_NAME":           "boilerplate",
	"LOG_ENV":            "development",
	"LOG_FILE":           "true",
	"LOG_CONSOLE":        "true",
	"LOG_CONSOLE_LEVEL":  "debug",
	"LOG_FILE_LEVEL":     "debug",
	"LOG_CONSOLE_OUTPUT": "full",
}

// Init prepares console and rotated file logging.
// The log folder is absolute. Finder starts a Mac app with a working directory of /, and a relative "logs" folder panics there.
// prezerolog panics when it cannot create the log folder, so file logging is turned off
// when neither folder can be created. Problems found before the logger exists are logged once it is ready.
func Init() {
	var problems []error
	if err := applyLogDefaults(); err != nil {
		problems = append(problems, err)
	}
	dir, err := prepareLogDirectory(logDirectory(), filepath.Join(os.TempDir(), "boilerplate", "logs"))
	if err != nil {
		problems = append(problems, err)
	}
	if dir == "" {
		_ = os.Setenv("LOG_FILE", "false")
	} else {
		_ = os.Setenv("LOG_DIR", dir)
	}
	log.InitLogging()
	for _, problem := range problems {
		Pre{}.Warn("log setup", map[string]any{"err": problem.Error()})
	}
}

// prepareLogDirectory creates preferred, or fallback when that fails. It returns the folder
// in use, and an empty folder when neither can be created. The error describes every failure.
func prepareLogDirectory(preferred, fallback string) (string, error) {
	preferredErr := os.MkdirAll(preferred, 0o755)
	if preferredErr == nil {
		return preferred, nil
	}
	preferredErr = fmt.Errorf("create log folder %s: %w", preferred, preferredErr)
	if err := os.MkdirAll(fallback, 0o755); err != nil {
		return "", errors.Join(preferredErr, fmt.Errorf("create fallback log folder %s: %w; file logging is off", fallback, err))
	}
	return fallback, fmt.Errorf("%w; using %s", preferredErr, fallback)
}

// applyLogDefaults loads .env, then fills any log setting that is still empty.
// A missing .env is normal. A .env that cannot be read or parsed is returned, and the defaults still apply.
func applyLogDefaults() error {
	var loadErr error
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		loadErr = fmt.Errorf("read .env: %w", err)
	}
	for key, value := range logDefaults {
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return loadErr
}

// logDirectory is where process logs are written. An absolute LOG_DIR is kept. Anything else goes under the user config folder.
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
