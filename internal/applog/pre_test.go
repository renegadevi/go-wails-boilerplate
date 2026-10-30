package applog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogDirectoryIsAbsolute(t *testing.T) {
	t.Setenv("LOG_DIR", "logs")
	dir := logDirectory()
	if !filepath.IsAbs(dir) {
		t.Fatalf("log directory must be absolute, got %s", dir)
	}
}

func TestEnvFileSetsLogLevel(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(".env", []byte("LOG_CONSOLE_LEVEL=warn\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unsetEnv(t, "LOG_CONSOLE_LEVEL")
	applyLogDefaults()
	if got := os.Getenv("LOG_CONSOLE_LEVEL"); got != "warn" {
		t.Fatalf("got %s", got)
	}
}

func TestLogLevelDefaultIsDebug(t *testing.T) {
	t.Chdir(t.TempDir())
	unsetEnv(t, "LOG_CONSOLE_LEVEL")
	unsetEnv(t, "LOG_FILE_LEVEL")
	applyLogDefaults()
	if got := os.Getenv("LOG_CONSOLE_LEVEL"); got != "debug" {
		t.Fatalf("console level %s", got)
	}
	if got := os.Getenv("LOG_FILE_LEVEL"); got != "debug" {
		t.Fatalf("file level %s", got)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func TestAbsoluteLogDirIsKept(t *testing.T) {
	const want = "/tmp/boilerplate-custom-logs"
	t.Setenv("LOG_DIR", want)
	if got := logDirectory(); got != want {
		t.Fatalf("got %s", got)
	}
}
