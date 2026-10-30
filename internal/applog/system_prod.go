//go:build production

package applog

// systemLogs keeps Wails system lines in dev. Production builds turn them off.
const systemLogs = false
