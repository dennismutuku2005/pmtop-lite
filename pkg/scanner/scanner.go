// Package scanner defines the shared PortEntry type and Scanner interface
// used across all platform implementations.
package scanner

import "time"

// PortEntry holds all information about a single open port,
// including resolved process details filled in by the state manager.
type PortEntry struct {
	Port      int
	Protocol  string // "TCP" or "UDP"
	PID       int32
	State     string // "LISTEN", "ESTABLISHED", etc.
	IsNew     bool   // true for one tick after first detection

	// Resolved process fields (filled by proc.Resolver)
	Name      string
	WorkDir   string
	MemoryMB  float64
	Uptime    string
	StartTime time.Time
	Service   string // "MySQL", "Node.js", etc.
}

// Scanner is implemented per platform.
type Scanner interface {
	Scan() ([]PortEntry, error)
}

// Options controls scanner behaviour, shared across platforms.
type Options struct {
	// ListenOnly=true (default): show only LISTEN-state TCP ports.
	// ListenOnly=false: show all connections including ESTABLISHED.
	ListenOnly bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{ListenOnly: true}
}
