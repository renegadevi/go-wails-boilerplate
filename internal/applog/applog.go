// Package applog is the process log, separate from the change log the user exports.
package applog

import "fmt"

// Logger records what the process did. Device changes go to SQLite, not here.
type Logger interface {
	Debug(msg string, fields map[string]any)
	Info(msg string, fields map[string]any)
	Warn(msg string, fields map[string]any)
	Error(msg string, err error)
}

// Nop drops every line. Tests use it so they do not open a log file.
type Nop struct{}

func (Nop) Debug(string, map[string]any) {}
func (Nop) Info(string, map[string]any)  {}
func (Nop) Warn(string, map[string]any)  {}
func (Nop) Error(string, error)          {}

// Merge copies fields into a new map so a caller can add one key without mutating a shared map.
func Merge(base map[string]any, key string, val any) map[string]any {
	out := make(map[string]any, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	out[key] = val
	return out
}

// ErrText is the string stored under err. A nil error is an empty string.
func ErrText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Format is a single debug line used when a logger field map would hide the point.
func Format(msg string, err error) string {
	if err == nil {
		return msg
	}
	return fmt.Sprintf("%s: %s", msg, err.Error())
}
