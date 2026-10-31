package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckedInFilesMatchConfig(t *testing.T) {
	root := repoRoot(t)
	desired, err := desiredTexts(root)
	if err != nil {
		t.Fatal(err)
	}
	for rel, next := range desired {
		old, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		if string(old) != next {
			t.Errorf("%s does not match build/config.yml and Taskfile.yml", rel)
		}
	}
}

func TestSyncUpdatesIdentityAndKeepsLocalLines(t *testing.T) {
	dir := copyPackagingFiles(t)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `version: "0.1.0"`, `version: "2.3.4"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `productName: "Boilerplate"`, `productName: "New Name"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `companyName: "Your Name"`, `companyName: "Smith & Co"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `copyright: "Copyright"`, `copyright: "A & B"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `productIdentifier: "com.wails.boilerplate"`, `productIdentifier: "com.example.newapp"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `homepage: "https://github.com/renegadevi/go-wails-boilerplate"`, `homepage: "https://example.com/new"`)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `license: "MIT"`, `license: "Apache-2.0"`)
	replaceIn(t, filepath.Join(dir, "Taskfile.yml"), `APP_NAME: "boilerplate"`, `APP_NAME: "my-app"`)

	changed, err := sync(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != len(packagingTargets) {
		t.Fatalf("updated %v", changed)
	}

	plist := readText(t, filepath.Join(dir, "build", "darwin", "Info.plist"))
	if !strings.Contains(plist, "<string>14.0.0</string>") {
		t.Fatal("Mac minimum version was rewritten")
	}
	for _, want := range []string{"New Name", "my-app", "com.example.newapp", "2.3.4", "A &amp; B"} {
		if !strings.Contains(plist, want) {
			t.Fatalf("Info.plist missing %s", want)
		}
	}
	infoJSON := readText(t, filepath.Join(dir, "build", "windows", "info.json"))
	if !strings.Contains(infoJSON, `"CompanyName": "Smith & Co"`) {
		t.Fatal("info.json company was not updated")
	}

	devPlist := readText(t, filepath.Join(dir, "build", "darwin", "Info.dev.plist"))
	if !strings.Contains(devPlist, "<string>14.0.0</string>") || !strings.Contains(devPlist, "NSAllowsLocalNetworking") {
		t.Fatal("dev plist lost its local keys")
	}

	manifest := readText(t, filepath.Join(dir, "build", "windows", "wails.exe.manifest"))
	if !strings.Contains(manifest, `name="com.example.newapp" version="2.3.4"`) {
		t.Fatal("manifest identity was not updated")
	}
	if !strings.Contains(manifest, `name="Microsoft.Windows.Common-Controls" version="6.0.0.0"`) {
		t.Fatal("common controls dependency was rewritten")
	}

	nsh := readText(t, filepath.Join(dir, "build", "windows", "nsis", "wails_tools.nsh"))
	if !strings.Contains(nsh, `!define INFO_PROJECTNAME "my-app"`) || !strings.Contains(nsh, `!define INFO_PRODUCTVERSION "2.3.4"`) {
		t.Fatal("NSIS defines were not updated")
	}
	if !strings.Contains(nsh, `${INFO_PRODUCTNAME}`) {
		t.Fatal("NSIS lost a variable reference")
	}

	nfpm := readText(t, filepath.Join(dir, "build", "linux", "nfpm", "nfpm.yaml"))
	for _, want := range []string{
		`name: "my-app"`,
		`version: "2.3.4"`,
		`vendor: "Smith & Co"`,
		`homepage: "https://example.com/new"`,
		`license: "Apache-2.0"`,
		`dst: "/usr/share/metainfo/com.example.newapp.metainfo.xml"`,
		`libgtk-4-1`,
		`release: "1"`,
	} {
		if !strings.Contains(nfpm, want) {
			t.Fatalf("nfpm.yaml missing %s", want)
		}
	}
	if strings.Contains(nfpm, "\r\n") == false || strings.Contains(strings.ReplaceAll(nfpm, "\r\n", ""), "\n") {
		t.Fatal("line endings changed")
	}

	again, err := sync(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Fatalf("second sync wrote %v", again)
	}
}

func TestVersionMustBeThreeParts(t *testing.T) {
	dir := copyPackagingFiles(t)
	replaceIn(t, filepath.Join(dir, "build", "config.yml"), `version: "0.1.0"`, `version: "1.2"`)
	_, err := sync(dir)
	if err == nil || !strings.Contains(err.Error(), "info.version must be X.Y.Z") {
		t.Fatalf("got %v", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func copyPackagingFiles(t *testing.T) string {
	t.Helper()
	src := repoRoot(t)
	dir := t.TempDir()
	rels := []string{"build/config.yml", "Taskfile.yml"}
	for _, item := range packagingTargets {
		rels = append(rels, item.rel)
	}
	for _, rel := range rels {
		body, err := os.ReadFile(filepath.Join(src, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		dest := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func replaceIn(t *testing.T, path, old, next string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, old) {
		t.Fatalf("%s does not contain %s", path, old)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(text, old, next, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
