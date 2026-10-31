// Command syncbuild copies the app name and version from build/config.yml
// into the Windows, Mac, and Linux packaging files.
//
// The binary name comes from APP_NAME in Taskfile.yml. The Mac minimum
// system version and the Linux AppStream entry are left as they are.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type info struct {
	appName     string
	productName string
	productID   string
	description string
	company     string
	copyright   string
	comments    string
	version     string
	homepage    string
	license     string
}

type target struct {
	rel   string
	apply func(string, info) (string, error)
}

func main() {
	root, err := findRoot()
	if err != nil {
		fail(err)
	}
	changed, err := sync(root)
	if err != nil {
		fail(err)
	}
	for _, path := range changed {
		fmt.Println(path)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(dir, "build", "config.yml")) && fileExists(filepath.Join(dir, "Taskfile.yml")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("build/config.yml not found; run this from the project directory")
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func sync(root string) ([]string, error) {
	desired, err := desiredTexts(root)
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, item := range packagingTargets {
		path := filepath.Join(root, filepath.FromSlash(item.rel))
		old, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.rel, err)
		}
		next := desired[item.rel]
		if next == string(old) {
			continue
		}
		if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
			return nil, fmt.Errorf("%s: %w", item.rel, err)
		}
		changed = append(changed, item.rel)
	}
	return changed, nil
}

func desiredTexts(root string) (map[string]string, error) {
	meta, err := load(root)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(packagingTargets))
	for _, item := range packagingTargets {
		path := filepath.Join(root, filepath.FromSlash(item.rel))
		old, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.rel, err)
		}
		next, err := item.apply(string(old), meta)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.rel, err)
		}
		out[item.rel] = next
	}
	return out, nil
}

var packagingTargets = []target{
	{"build/windows/wails.exe.manifest", applyManifest},
	{"build/windows/info.json", applyInfoJSON},
	{"build/windows/nsis/wails_tools.nsh", applyNSIS},
	{"build/darwin/Info.plist", applyPlist},
	{"build/darwin/Info.dev.plist", applyPlist},
	{"build/linux/nfpm/nfpm.yaml", applyNFPM},
}

func load(root string) (info, error) {
	configText, err := readFile(filepath.Join(root, "build", "config.yml"))
	if err != nil {
		return info{}, err
	}
	taskText, err := readFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		return info{}, err
	}
	var meta info
	fields := []struct {
		text    string
		pattern string
		label   string
		dest    *string
	}{
		{taskText, `(?m)^  APP_NAME: "([^"]+)"`, "APP_NAME in Taskfile.yml", &meta.appName},
		{configText, `(?m)^  productName: "([^"]+)"`, "info.productName", &meta.productName},
		{configText, `(?m)^  productIdentifier: "([^"]+)"`, "info.productIdentifier", &meta.productID},
		{configText, `(?m)^  description: "([^"]+)"`, "info.description", &meta.description},
		{configText, `(?m)^  companyName: "([^"]+)"`, "info.companyName", &meta.company},
		{configText, `(?m)^  copyright: "([^"]+)"`, "info.copyright", &meta.copyright},
		{configText, `(?m)^  comments: "([^"]+)"`, "info.comments", &meta.comments},
		{configText, `(?m)^  version: "([^"]+)"`, "info.version", &meta.version},
		{configText, `(?m)^  homepage: "([^"]+)"`, "info.homepage", &meta.homepage},
		{configText, `(?m)^  license: "([^"]+)"`, "info.license", &meta.license},
	}
	for _, field := range fields {
		value, err := readValue(field.text, field.pattern, field.label)
		if err != nil {
			return info{}, err
		}
		*field.dest = value
	}
	if strings.Contains(meta.productName, "/") || strings.Contains(meta.productName, `\`) {
		return info{}, fmt.Errorf("info.productName cannot contain a slash")
	}
	if strings.ContainsAny(meta.appName, `/\`) || strings.Contains(meta.appName, `"`) {
		return info{}, fmt.Errorf("APP_NAME cannot contain a slash or a quote")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+$`).MatchString(meta.productID) {
		return info{}, fmt.Errorf("info.productIdentifier must be a reverse-DNS id, such as com.example.app")
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(meta.version) {
		return info{}, fmt.Errorf("info.version must be X.Y.Z")
	}
	if !strings.HasPrefix(meta.homepage, "https://") && !strings.HasPrefix(meta.homepage, "http://") {
		return info{}, fmt.Errorf("info.homepage must be an http or https URL")
	}
	if strings.ContainsAny(meta.license, "<>&") {
		return info{}, fmt.Errorf("info.license must be an SPDX identifier, such as MIT")
	}
	return meta, nil
}

func readValue(text, pattern, label string) (string, error) {
	match := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if match == nil || strings.TrimSpace(match[1]) == "" {
		return "", fmt.Errorf("%s must be set", label)
	}
	return strings.TrimSpace(match[1]), nil
}

func readFile(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func applyManifest(text string, meta info) (string, error) {
	re := regexp.MustCompile(`(<assemblyIdentity type="win32" name=")[^"]*(" version=")[^"]*(" processorArchitecture="\*"/>)`)
	return transformOnce(text, re, "assemblyIdentity", func(sub []string) string {
		return sub[1] + meta.productID + sub[2] + meta.version + sub[3]
	})
}

func applyInfoJSON(text string, meta info) (string, error) {
	fields := []struct{ key, value string }{
		{"file_version", meta.version},
		{"ProductVersion", meta.version},
		{"CompanyName", meta.company},
		{"FileDescription", meta.description},
		{"LegalCopyright", meta.copyright},
		{"ProductName", meta.productName},
		{"Comments", meta.comments},
	}
	var err error
	for _, field := range fields {
		text, err = setJSON(text, field.key, field.value)
		if err != nil {
			return "", err
		}
	}
	return text, nil
}

func applyNSIS(text string, meta info) (string, error) {
	fields := []struct{ name, value string }{
		{"INFO_PROJECTNAME", meta.appName},
		{"INFO_COMPANYNAME", meta.company},
		{"INFO_PRODUCTNAME", meta.productName},
		{"INFO_PRODUCTVERSION", meta.version},
		{"INFO_COPYRIGHT", meta.copyright},
	}
	var err error
	for _, field := range fields {
		text, err = setDefine(text, field.name, field.value)
		if err != nil {
			return "", err
		}
	}
	return text, nil
}

func applyPlist(text string, meta info) (string, error) {
	fields := []struct{ key, value string }{
		{"CFBundleName", meta.productName},
		{"CFBundleExecutable", meta.appName},
		{"CFBundleIdentifier", meta.productID},
		{"CFBundleVersion", meta.version},
		{"CFBundleGetInfoString", meta.comments},
		{"CFBundleShortVersionString", meta.version},
		{"NSHumanReadableCopyright", meta.copyright},
	}
	var err error
	for _, field := range fields {
		text, err = setPlistString(text, field.key, field.value)
		if err != nil {
			return "", err
		}
	}
	return text, nil
}

func applyNFPM(text string, meta info) (string, error) {
	lines := []struct{ key, value string }{
		{"name", meta.appName},
		{"version", meta.version},
		{"description", meta.description},
		{"vendor", meta.company},
		{"homepage", meta.homepage},
		{"license", meta.license},
	}
	var err error
	for _, line := range lines {
		text, err = replaceYAMLKey(text, line.key, line.value)
		if err != nil {
			return "", err
		}
	}
	paths := []struct{ pattern, value, label string }{
		{`src: "\./bin/[^"]+"`, `src: "./bin/` + meta.appName + `"`, "binary source"},
		{`dst: "/usr/local/bin/[^"]+"`, `dst: "/usr/local/bin/` + meta.appName + `"`, "binary destination"},
		{`dst: "/usr/share/icons/hicolor/128x128/apps/[^"]+"`, `dst: "/usr/share/icons/hicolor/128x128/apps/` + meta.appName + `.png"`, "icon destination"},
		{`src: "\./build/linux/[^"]+\.desktop"`, `src: "./build/linux/` + meta.appName + `.desktop"`, "desktop source"},
		{`dst: "/usr/share/applications/[^"]+"`, `dst: "/usr/share/applications/` + meta.appName + `.desktop"`, "desktop destination"},
		{`src: "\./build/linux/[^"]+\.metainfo\.xml"`, `src: "./build/linux/` + meta.appName + `.metainfo.xml"`, "metainfo source"},
		{`dst: "/usr/share/metainfo/[^"]+"`, `dst: "/usr/share/metainfo/` + meta.productID + `.metainfo.xml"`, "metainfo destination"},
	}
	for _, path := range paths {
		text, err = replaceLiteral(text, path.pattern, path.value, path.label)
		if err != nil {
			return "", err
		}
	}
	return text, nil
}

func setPlistString(text, key, value string) (string, error) {
	re := regexp.MustCompile(`(<key>` + regexp.QuoteMeta(key) + `</key>\s*<string>)[^<]*(</string>)`)
	return transformOnce(text, re, key, func(sub []string) string {
		return sub[1] + xmlText(value) + sub[2]
	})
}

func setJSON(text, key, value string) (string, error) {
	quoted, err := jsonString(value)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`("` + regexp.QuoteMeta(key) + `"\s*:\s*)"(?:\\.|[^"\\])*"`)
	return transformOnce(text, re, key, func(sub []string) string {
		return sub[1] + quoted
	})
}

func setDefine(text, name, value string) (string, error) {
	re := regexp.MustCompile(`(?m)^([ \t]*!define ` + regexp.QuoteMeta(name) + ` ")[^"]*(")`)
	return transformOnce(text, re, name, func(sub []string) string {
		return sub[1] + nsisQuote(value) + sub[2]
	})
}

func replaceYAMLKey(text, key, value string) (string, error) {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:[^\r\n]*`)
	return transformOnce(text, re, key, func([]string) string {
		return key + ": " + yamlQuote(value)
	})
}

func replaceLiteral(text, pattern, repl, label string) (string, error) {
	return transformOnce(text, regexp.MustCompile(pattern), label, func([]string) string {
		return repl
	})
}

func transformOnce(text string, re *regexp.Regexp, label string, build func([]string) string) (string, error) {
	matches := re.FindAllStringSubmatch(text, -1)
	locs := re.FindAllStringIndex(text, -1)
	if len(matches) != 1 || len(locs) != 1 {
		return "", fmt.Errorf("%s must appear once", label)
	}
	return text[:locs[0][0]] + build(matches[0]) + text[locs[0][1]:], nil
}

func xmlText(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(value)
}

func jsonString(value string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

func yamlQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func nsisQuote(value string) string {
	value = strings.ReplaceAll(value, `$`, `$$`)
	value = strings.ReplaceAll(value, `"`, `$\"`)
	return value
}
