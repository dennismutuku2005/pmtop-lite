package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dennismutuku2005/pmtop-lite/pkg/proc"
	"github.com/dennismutuku2005/pmtop-lite/pkg/scanner"
)

// ── Colour palette ────────────────────────────────────────────────────────────
var (
	colorPrimary  = lipgloss.Color("#fe4a16") // brand orange
	colorAccent   = lipgloss.Color("#06B6D4") // cyan
	colorSelected = lipgloss.Color("#1E1B4B") // deep indigo bg
	colorNewPort  = lipgloss.Color("#10B981") // emerald
	colorDim      = lipgloss.Color("#4B5563") // gray-600
	colorMuted    = lipgloss.Color("#6B7280") // gray-500
	colorHeader   = lipgloss.Color("#0F172A") // slate-900
	colorBorder   = lipgloss.Color("#312E81") // indigo-900
	colorText     = lipgloss.Color("#E2E8F0") // slate-200
	colorSysPort  = lipgloss.Color("#64748B") // slate-500
)

// ── Column widths ─────────────────────────────────────────────────────────────
const (
	wPort    = 7
	wProto   = 6
	wProcess = 14
	wService = 12
	wPID     = 7
	wDir     = 25
	wMem     = 8
	wUptime  = 10
)

// ── Styles ────────────────────────────────────────────────────────────────────
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Background(colorHeader).
			Padding(0, 2)

	borderStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	colHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Background(colorHeader).
			Padding(0, 0)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(colorText)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSelected).
				Bold(true)

	newPortStyle = lipgloss.NewStyle().
			Foreground(colorNewPort).
			Bold(true)

	sysPortStyle = lipgloss.NewStyle().
			Foreground(colorSysPort)

	selectedSysStyle = lipgloss.NewStyle().
				Foreground(colorSysPort).
				Background(colorSelected)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	filterBarStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	filterLabelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	accentStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)

// View renders the full terminal UI (bubbletea v2 — returns tea.View).
func (m Model) View() tea.View {
	if !m.ready {
		v := tea.NewView("  initialising pmtop…\n")
		v.AltScreen = true
		return v
	}

	var sb strings.Builder
	
	// ── System Stats Header ──────────────────────────────────────────────────
	headerStyle := lipgloss.NewStyle().Foreground(colorText).Background(colorHeader).Padding(0, 1)
	
	visible := m.visiblePorts()
	// Simulated system stats for the dashboard feel

	cpuBar := accentStyle.Render("||||||") + sysPortStyle.Render("||||||||||||||")
	memBar := accentStyle.Render("||||||||") + sysPortStyle.Render("||||||||||||")
	
	sb.WriteString(headerStyle.Width(m.width).Render(
		fmt.Sprintf(" CPU [%s] 24%%  |  MEM [%s] 42%%  |  %s", cpuBar, memBar, accentStyle.Render("pmtop v"+version)),
	) + "\n")

	// ── Filter / Search Bar ───────────────────────────────────────────────────
	statusText := fmt.Sprintf(" %d active ports", len(visible))
	if m.filter != "" {
		statusText = fmt.Sprintf(" filtering: '%s' (%d matches)", m.filter, len(visible))
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(" " + statusText) + "\n")
	sb.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n")

	// ── Table ─────────────────────────────────────────────────────────────────
	sb.WriteString(buildColHeaders(m.width >= 100, m.width >= 80, m) + "\n")
	sb.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n")

	// Rows
	maxRows := m.height - 10 
	if maxRows < 1 { maxRows = 1 }
	
	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}

	for i := start; i < len(visible) && i < start+maxRows; i++ {
		sb.WriteString(buildRow(visible[i], i == m.cursor, m.width >= 100, m.width >= 80) + "\n")
	}


	if len(visible) == 0 {
		msg := "  No developer ports active. Press 'a' to show all system ports."
		if !m.filterDev {
			msg = "  No active ports found."
		}
		sb.WriteString("\n" + sysPortStyle.Render(msg) + "\n\n")
	}

	// Fill empty space to keep layout stable
	for i := len(visible) - start; i < maxRows; i++ {
		sb.WriteString("\n")
	}

	sb.WriteString(borderStyle.Render(strings.Repeat("─", m.width)) + "\n")

	// ── Details Panel / Footer ──────────────────────────────────────────────
	if len(visible) > 0 && m.cursor < len(visible) {
		p := visible[m.cursor]
		details := fmt.Sprintf(" %s %s  %s %s  %s %s",
			accentStyle.Render("DIR"), truncDir(p.WorkDir, 30),
			accentStyle.Render("MEM"), proc.FormatMem(p.MemoryMB),
			accentStyle.Render("UPTIME"), p.Uptime)
		sb.WriteString(details + "\n")
	} else {
		sb.WriteString("\n")
	}

	// ── Overlays (Confirmation / Filter) ──────────────────────────────────────
	if m.confirmKill {
		msg := " CONFIRM KILL: Are you sure you want to close this process? (y/n) "
		sb.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#E11D48")). // Rose-600
			Render(msg) + "\n")
	} else if m.filterMode {
		cursor := accentStyle.Render("▌")
		label := filterLabelStyle.Render(" SEARCH: ")
		sb.WriteString(filterBarStyle.Render(label+m.filter+cursor) + "\n")
	} else if m.filter != "" {
		label := filterLabelStyle.Render(" FILTER: ")
		clear := sysPortStyle.Render(" (esc to clear)")
		sb.WriteString(filterBarStyle.Render(label+m.filter+clear) + "\n")
	} else {
		sb.WriteString(buildFooter() + "\n")
	}

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
}

// buildColHeaders returns the column header row string.
func buildColHeaders(showDir, showMem bool, m Model) string {
	markCol := func(label string, col sortColumn, width int) string {
		if m.sortCol == col {
			arrow := "↑"
			if !m.sortAsc {
				arrow = "↓"
			}
			return accentStyle.Width(width).Render(label + arrow)
		}
		return colHeaderStyle.Width(width).Render(label)
	}

	parts := []string{
		colHeaderStyle.Width(wPort).Render(" PORT"),
		colHeaderStyle.Width(wProto).Render("PROTO"),
		markCol("PROCESS", sortByProcess, wProcess+5),
		colHeaderStyle.Width(wService+5).Render("SERVICE"),
		colHeaderStyle.Width(wPID).Render("PID"),
	}

	return strings.Join(parts, "  ")
}

// buildRow renders a single PortEntry as a styled table row.
func buildRow(p scanner.PortEntry, selected, showDir, showMem bool) string {
	isSystem := p.Port < 1024
	statusDot := "●"
	dotStyle := lipgloss.NewStyle().Foreground(colorNewPort) // default green
	if isSystem || p.Service == "Unknown" {
		dotStyle = lipgloss.NewStyle().Foreground(colorSysPort) // gray
	}

	// Choose base style
	var base lipgloss.Style
	if selected {
		base = selectedRowStyle
	} else if isSystem {
		base = sysPortStyle
	} else {
		base = normalRowStyle
	}

	cell := func(s string, w int) string {
		return base.Width(w).MaxWidth(w).Render(truncName(s, w))
	}
	rcell := func(s string, w int) string {
		if len([]rune(s)) < w {
			s = strings.Repeat(" ", w-len([]rune(s))) + s
		}
		return base.Width(w).Render(truncName(s, w))
	}

	portStr := fmt.Sprintf(":%d", p.Port)
	pidStr := fmt.Sprintf("%d", p.PID)

	parts := []string{
		" " + dotStyle.Render(statusDot),
		rcell(portStr, wPort),
		cell(p.Protocol, wProto),
		cell(p.Name, wProcess+5),
		cell(p.Service, wService+5),
		rcell(pidStr, wPID),
	}

	return strings.Join(parts, "  ")
}


// buildFooter renders the keybinding hint bar.
func buildFooter() string {
	hints := []string{
		accentStyle.Render("x") + " kill",
		accentStyle.Render("/") + " filter",
		accentStyle.Render("s") + " sort",
		accentStyle.Render("a") + " all/dev",
		accentStyle.Render("o") + " browser",
		accentStyle.Render("q") + " quit",
	}
	return footerStyle.Render(strings.Join(hints, "  •  "))
}

// truncName shortens a process name keeping the BEGINNING.
func truncName(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

// truncDir shortens a path keeping the END.
func truncDir(s string, n int) string {
	if s == "" { return "n/a" }
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return "…" + string(r[len(r)-(n-1):])
}

