// Package proc resolves human-readable process information from a PID.
package proc

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	gopsutil "github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo holds all display information for a process.
type ProcessInfo struct {
	Name      string
	WorkDir   string
	MemoryMB  float64
	Uptime    string
	StartTime time.Time
}

// homeDir is cached at startup.
var homeDir string

func init() {
	if u, err := user.Current(); err == nil {
		homeDir = u.HomeDir
	}
}

// Resolve returns process info for the given PID. Never returns an error;
// unknown fields are filled with sensible defaults.
func Resolve(pid int32) ProcessInfo {
	p, err := gopsutil.NewProcess(pid)
	if err != nil {
		return ProcessInfo{Name: "unknown", WorkDir: "unknown", Uptime: "—"}
	}

	info := ProcessInfo{}

	// Name
	if name, err := p.Name(); err == nil && name != "" {
		// Strip .exe suffix on Windows for cleaner display
		info.Name = strings.TrimSuffix(name, ".exe")
	} else {
		info.Name = fmt.Sprintf("PID%d", pid)
	}

	// Working directory
	if cwd, err := p.Cwd(); err == nil && cwd != "" {
		info.WorkDir = shortenPath(cwd)
	} else {
		info.WorkDir = "unknown"
	}

	// Memory (RSS in MB)
	if mem, err := p.MemoryInfo(); err == nil && mem != nil {
		info.MemoryMB = float64(mem.RSS) / 1024 / 1024
	}

	// Start time and uptime
	if createMs, err := p.CreateTime(); err == nil {
		info.StartTime = time.Unix(0, createMs*int64(time.Millisecond))
		info.Uptime = formatUptime(time.Since(info.StartTime))
	} else {
		info.Uptime = "—"
	}

	return info
}

// shortenPath replaces the home directory prefix with ~/ and
// truncates very long paths from the left with ...
func shortenPath(path string) string {
	// Replace absolute home dir with ~/
	if homeDir != "" {
		rel, err := filepath.Rel(homeDir, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			path = "~/" + rel
		}
	}
	// Normalise separators
	path = filepath.ToSlash(path)

	// Truncate from left if longer than 32 chars
	const maxLen = 32
	if len(path) > maxLen {
		// Find a slash boundary
		cut := path[len(path)-maxLen:]
		if idx := strings.Index(cut, "/"); idx >= 0 {
			cut = cut[idx:]
		}
		path = "..." + cut
	}
	return path
}

// formatUptime converts a duration to a human-readable string.
// Examples: "44m", "2h 14m", "6d 11h"
func formatUptime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

// FormatMem formats megabytes as a short string.
func FormatMem(mb float64) string {
	if mb < 1 {
		return fmt.Sprintf("%.0fK", mb*1024)
	}
	if mb >= 1024 {
		return fmt.Sprintf("%.1fG", mb/1024)
	}
	return fmt.Sprintf("%.0fM", mb)
}

// FileExists is a convenience used in tests.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
