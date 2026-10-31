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
	if err := applyLogDefaults(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("LOG_CONSOLE_LEVEL"); got != "warn" {
		t.Fatalf("got %s", got)
	}
}

func TestUnreadableEnvFileIsReportedAndDefaultsApply(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir(".env", 0o755); err != nil {
		t.Fatal(err)
	}
	unsetEnv(t, "LOG_CONSOLE_LEVEL")
	if err := applyLogDefaults(); err == nil {
		t.Fatal("expected an error for an unreadable .env")
	}
	if got := os.Getenv("LOG_CONSOLE_LEVEL"); got != "debug" {
		t.Fatalf("console level %s", got)
	}
}

func TestLogLevelDefaultIsDebug(t *testing.T) {
	t.Chdir(t.TempDir())
	unsetEnv(t, "LOG_CONSOLE_LEVEL")
	unsetEnv(t, "LOG_FILE_LEVEL")
	if err := applyLogDefaults(); err != nil {
		t.Fatalf("a missing .env is not an error: %v", err)
	}
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

// blockedDirectory returns a path that cannot be created, because its parent is a file.
func blockedDirectory(t *testing.T) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(file, "logs")
}

func TestPrepareLogDirectoryUsesPreferred(t *testing.T) {
	preferred := filepath.Join(t.TempDir(), "logs")
	dir, err := prepareLogDirectory(preferred, blockedDirectory(t))
	if err != nil || dir != preferred {
		t.Fatalf("got %q, %v", dir, err)
	}
}

func TestPrepareLogDirectoryFallsBackAndReports(t *testing.T) {
	fallback := filepath.Join(t.TempDir(), "logs")
	dir, err := prepareLogDirectory(blockedDirectory(t), fallback)
	if dir != fallback {
		t.Fatalf("got %q", dir)
	}
	if err == nil {
		t.Fatal("expected the preferred folder failure to be reported")
	}
}

func TestPrepareLogDirectoryReportsWhenBothFail(t *testing.T) {
	dir, err := prepareLogDirectory(blockedDirectory(t), blockedDirectory(t))
	if dir != "" || err == nil {
		t.Fatalf("got %q, %v", dir, err)
	}
}

func TestAbsoluteLogDirIsKept(t *testing.T) {
	const want = "/tmp/boilerplate-custom-logs"
	t.Setenv("LOG_DIR", want)
	if got := logDirectory(); got != want {
		t.Fatalf("got %s", got)
	}
}
