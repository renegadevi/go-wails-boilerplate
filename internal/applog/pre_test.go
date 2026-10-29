package applog

import (
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

func TestAbsoluteLogDirIsKept(t *testing.T) {
	const want = "/tmp/boilerplate-custom-logs"
	t.Setenv("LOG_DIR", want)
	if got := logDirectory(); got != want {
		t.Fatalf("got %s", got)
	}
}
