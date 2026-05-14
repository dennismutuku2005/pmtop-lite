// Package state manages the live port list, diffs, and process killing.
package state

import (
	"context"
	"os"
	"time"

	"github.com/dennismutuku2005/pmtop-lite/pkg/logger"
	"github.com/dennismutuku2005/pmtop-lite/pkg/proc"
	"github.com/dennismutuku2005/pmtop-lite/pkg/scanner"
	"github.com/dennismutuku2005/pmtop-lite/pkg/services"
)

// PortUpdateMsg is sent to the TUI channel on every scan tick.
type PortUpdateMsg struct {
	Ports []scanner.PortEntry
}

// Manager owns the port map and drives periodic scanning.
type Manager struct {
	sc      scanner.Scanner
	updates chan PortUpdateMsg
	killCh  chan int32
	prev    map[portKey]scanner.PortEntry // tracks which ports existed last tick
	History *logger.History
	showAll bool               // if true, show ESTABLISHED connections too
}

type portKey struct {
	port  int
	proto string
}

// New creates a Manager with the platform-appropriate scanner.
// showAll=false → LISTEN-only (clean); showAll=true → all connections.
func New(showAll bool) *Manager {
	opts := scanner.DefaultOptions()
	opts.ListenOnly = !showAll
	return &Manager{
		sc:      scanner.NewPlatformScannerWithOptions(opts),
		updates: make(chan PortUpdateMsg, 1),
		killCh:  make(chan int32, 1),
		prev:    make(map[portKey]scanner.PortEntry),
		History: logger.NewHistory(100),
		showAll: showAll,
	}
}

// Updates returns the read-only channel the TUI listens on.
func (m *Manager) Updates() <-chan PortUpdateMsg {
	return m.updates
}

// Kill sends SIGTERM / TerminateProcess to a PID.
func (m *Manager) Kill(pid int32) {
	m.killCh <- pid
}

// Run starts the scan loop. Call in a goroutine.
func (m *Manager) Run(ctx context.Context) {
	m.scan() // immediate first scan

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.scan()
		case pid := <-m.killCh:
			m.killPID(pid)
			time.Sleep(100 * time.Millisecond)
			m.scan()
		}
	}
}

// scan fetches current ports, resolves process info, diffs vs previous, and sends.
func (m *Manager) scan() {
	raw, err := m.sc.Scan()
	if err != nil {
		return
	}

	current := make(map[portKey]scanner.PortEntry, len(raw))
	seen := make(map[portKey]bool) // deduplication
	resolved := make([]scanner.PortEntry, 0, len(raw))

	for _, e := range raw {
		// Skip invalid ports and the System Idle Process (PID 0)
		// NOTE: PID 4 is 'System' and often handles Port 80/443 on Windows.
		if e.Port <= 0 || e.Port > 65535 || (e.PID == 0 && !services.IsDev(e.Port, "", "")) {
			continue
		}



		k := portKey{e.Port, e.Protocol}

		// Deduplicate: same port+protocol can appear multiple times
		// (e.g. IPv4+IPv6, or multiple ESTABLISHED on same listening port)
		if seen[k] {
			continue
		}
		seen[k] = true
		prevEntry, existed := m.prev[k]
		e.IsNew = !existed
		current[k] = e

		// Event Detection
		if !existed {
			m.History.Add(logger.PortEvent{
				Timestamp: time.Now(), Port: e.Port, Protocol: e.Protocol,
				Type: logger.EventOpen, Process: e.Name, PID: e.PID,
			})
		} else if prevEntry.PID != e.PID {
			m.History.Add(logger.PortEvent{
				Timestamp: time.Now(), Port: e.Port, Protocol: e.Protocol,
				Type: logger.EventMove, Process: e.Name, PID: e.PID,
			})
		}

		// Resolve process info (best-effort)
		info := proc.Resolve(e.PID)
		e.Name = info.Name
		e.WorkDir = info.WorkDir
		e.MemoryMB = info.MemoryMB
		e.Uptime = info.Uptime
		e.StartTime = info.StartTime
		e.Service = services.Detect(e.Port, e.Name)

		resolved = append(resolved, e)
	}

	// Detect CLOSE events
	for k, prevEntry := range m.prev {
		if !seen[k] {
			m.History.Add(logger.PortEvent{
				Timestamp: time.Now(), Port: prevEntry.Port, Protocol: prevEntry.Protocol,
				Type: logger.EventClose, Process: prevEntry.Name, PID: prevEntry.PID,
			})
		}
	}

	m.prev = current

	// Non-blocking send; drop update if TUI hasn't consumed last one
	select {
	case m.updates <- PortUpdateMsg{Ports: resolved}:
	default:
		// Replace stale update
		select {
		case <-m.updates:
		default:
		}
		m.updates <- PortUpdateMsg{Ports: resolved}
	}
}

// killPID terminates a process using OS-appropriate method.
func (m *Manager) killPID(pid int32) {
	p, err := os.FindProcess(int(pid))
	if err != nil {
		return
	}
	// On Windows, os.Process.Kill() calls TerminateProcess.
	// On Unix, it sends SIGKILL.
	_ = p.Kill()
}
