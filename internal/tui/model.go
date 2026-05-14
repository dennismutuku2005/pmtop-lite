// Package tui implements the pmtop terminal UI using Bubble Tea v2.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"github.com/dennismutuku2005/pmtop-lite/pkg/scanner"
	"github.com/dennismutuku2005/pmtop-lite/pkg/services"
	"github.com/dennismutuku2005/pmtop-lite/pkg/state"
)

const version = "0.1.0"

// sortColumn enumerates the columns we can sort by.
type sortColumn int

const (
	sortByPort sortColumn = iota
	sortByProcess
	sortByMem
	sortByUptime
	sortColumnCount
)

var sortColumnNames = []string{"PORT", "PROCESS", "MEM", "UPTIME"}

// Model is the Bubble Tea application model.
type Model struct {
	ports      []scanner.PortEntry
	cursor     int
	filter     string
	filterMode bool
	filterBuf  []rune // rune buffer for filter input
	sortCol    sortColumn
	sortAsc    bool
	width      int
	height     int
	keys       keyMap
	mgr        *state.Manager
	cancel     context.CancelFunc
	ready      bool
	showAll    bool // true = show all connections; false = LISTEN-only
	filterDev  bool // true = only show developer ports (default)
	confirmKill bool
}

// portUpdateMsg wraps state.PortUpdateMsg for Bubble Tea.
type portUpdateMsg state.PortUpdateMsg

// waitForUpdate returns a Cmd that blocks until the next manager update.
func waitForUpdate(ch <-chan state.PortUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		return portUpdateMsg(<-ch)
	}
}

// NewModel constructs the initial model, starting the state manager immediately.
// showAll=false is the clean default; showAll=true is set via --all flag.
func NewModel(showAll bool) Model {
	ctx, cancel := context.WithCancel(context.Background())
	mgr := state.New(showAll)
	go mgr.Run(ctx)

	return Model{
		keys:      defaultKeys,
		sortCol:   sortByPort,
		sortAsc:   true,
		width:     80,
		height:    24,
		mgr:       mgr,
		cancel:    cancel,
		showAll:   showAll,
		filterDev: true, // Default to developer ports only
	}
}

// Init returns the command that waits for the first port update.
func (m Model) Init() tea.Cmd {
	return waitForUpdate(m.mgr.Updates())
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case portUpdateMsg:
		m.ports = msg.Ports
		if m.cursor >= len(m.visiblePorts()) && m.cursor > 0 {
			m.cursor = len(m.visiblePorts()) - 1
		}
		return m, waitForUpdate(m.mgr.Updates())

	case tea.KeyPressMsg:
		key := msg.String()

		// ── Filter mode ──────────────────────────────────────────────────
		if m.filterMode {
			switch key {
			case "esc", "enter":
				if key == "esc" {
					m.filterBuf = nil
					m.filter = ""
				}
				m.filterMode = false
			case "backspace":
				if len(m.filterBuf) > 0 {
					m.filterBuf = m.filterBuf[:len(m.filterBuf)-1]
					m.filter = string(m.filterBuf)
				}
			default:
				// Accept printable characters only
				if len(msg.Text) == 1 {
					r := rune(msg.Text[0])
					if unicode.IsPrint(r) {
						m.filterBuf = append(m.filterBuf, r)
						m.filter = string(m.filterBuf)
					}
				}
			}
			return m, nil
		}

		// ── Normal mode ───────────────────────────────────────────────────
		switch {
		case m.confirmKill:
			if key == "y" || key == "Y" {
				visible := m.visiblePorts()
				if m.cursor < len(visible) {
					m.mgr.Kill(visible[m.cursor].PID)
				}
				m.confirmKill = false
			} else {
				m.confirmKill = false
			}
			return m, nil

		case m.keys.Quit.contains(key):
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit

		case m.keys.Up.contains(key):
			if m.cursor > 0 {
				m.cursor--
			}

		case m.keys.Down.contains(key):
			visible := m.visiblePorts()
			if m.cursor < len(visible)-1 {
				m.cursor++
			}

		case m.keys.Kill.contains(key):
			m.confirmKill = true

		case m.keys.Filter.contains(key):
			m.filterMode = true
			m.filterBuf = []rune(m.filter)

		case m.keys.Sort.contains(key):
			if m.sortCol == sortColumnCount-1 {
				m.sortCol = 0
				m.sortAsc = !m.sortAsc
			} else {
				m.sortCol++
			}

		case key == "a" || key == "A": // Toggle dev-only vs all
			m.filterDev = !m.filterDev

		case m.keys.Open.contains(key):
			visible := m.visiblePorts()
			if m.cursor < len(visible) {
				p := visible[m.cursor]
				openBrowser(fmt.Sprintf("http://localhost:%d", p.Port))
			}

		case m.keys.Copy.contains(key):
			visible := m.visiblePorts()
			if m.cursor < len(visible) {
				copyToClipboard(fmt.Sprintf("%d", visible[m.cursor].Port))
			}
		}
	}

	return m, nil
}

// visiblePorts applies filter and sort to m.ports.
func (m Model) visiblePorts() []scanner.PortEntry {
	ports := make([]scanner.PortEntry, 0, len(m.ports))
	filter := strings.ToLower(m.filter)

	for _, p := range m.ports {
		if m.filterDev && !services.IsDev(p.Port, p.Service, p.Name) {
			continue
		}

		if filter != "" {
			portStr := fmt.Sprintf(":%d", p.Port)
			if !strings.Contains(strings.ToLower(p.Name), filter) &&
				!strings.Contains(strings.ToLower(p.WorkDir), filter) &&
				!strings.Contains(portStr, filter) &&
				!strings.Contains(strings.ToLower(p.Protocol), filter) {
				continue
			}
		}
		ports = append(ports, p)
	}

	// Sort
	sort.SliceStable(ports, func(i, j int) bool {
		a, b := ports[i], ports[j]
		var less bool
		switch m.sortCol {
		case sortByPort:
			less = a.Port < b.Port
		case sortByProcess:
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case sortByMem:
			less = a.MemoryMB < b.MemoryMB
		case sortByUptime:
			less = a.StartTime.Before(b.StartTime)
		default:
			less = a.Port < b.Port
		}
		if m.sortAsc {
			return less
		}
		return !less
	})

	return ports
}

// headerText returns the top status bar text.
func (m Model) headerText() string {
	visible := m.visiblePorts()
	total := len(m.ports)
	mode := "LISTEN"
	if m.showAll {
		mode = "ALL"
	}
	if m.filter != "" {
		return fmt.Sprintf("pmtop v%s   %d / %d ports   sort: %s   mode: %s   refresh: 2s",
			version, len(visible), total, sortColumnNames[m.sortCol], mode)
	}
	return fmt.Sprintf("pmtop v%s   %d ports   sort: %s   mode: %s   refresh: 2s",
		version, total, sortColumnNames[m.sortCol], mode)
}
