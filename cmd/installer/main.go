package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	styleTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).MarginBottom(1)
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	styleBox     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("#7C3AED"))
)

type model struct {
	step        int
	installing  bool
	done        bool
	err         error
	progress    float64
	status      string
	installPath string
}

type progressMsg float64
type statusMsg string
type doneMsg struct{}
type errMsg error

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.installing && !m.done && msg.String() == "enter" {
			m.installing = true
			return m, m.doInstall
		}
	case progressMsg:
		m.progress = float64(msg)
	case statusMsg:
		m.status = string(msg)
	case doneMsg:
		m.done = true
		m.installing = false
	case errMsg:
		m.err = error(msg)
		m.installing = false
	}
	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder

	s.WriteString(styleTitle.Render("pmtop Suite Installer"))
	s.WriteString("\n")

	if m.err != nil {
		s.WriteString(styleError.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\nPress 'q' to quit.")
		return tea.NewView(styleBox.Render(s.String()))
	}

	if m.done {
		s.WriteString(styleSuccess.Render("✓ Installation Complete!"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Installed to: %s\n", m.installPath))
		s.WriteString("The 'pmtop' command is now available globally.\n")
		s.WriteString("\nPress 'q' to exit.")
		return tea.NewView(styleBox.Render(s.String()))
	}

	if m.installing {
		s.WriteString(fmt.Sprintf("Installing... [%.0f%%]\n", m.progress*100))
		s.WriteString(styleDim.Render(m.status))
	} else {
		s.WriteString(fmt.Sprintf("Target Directory: %s\n\n", m.installPath))
		s.WriteString("Press " + styleTitle.Render("ENTER") + " to start installation.\n")
		s.WriteString(styleDim.Render("(Esc to cancel)"))
	}

	return tea.NewView(styleBox.Render(s.String()))
}

func (m model) doInstall() tea.Msg {
	dest := m.installPath
	
	// Create Dir
	if err := os.MkdirAll(dest, 0755); err != nil {
		return errMsg(err)
	}

	// Copy CLI
	src := "pmtop.exe"
	if _, err := os.Stat(src); err != nil {
		return errMsg(fmt.Errorf("source binary %s not found. Build it first!", src))
	}
	
	if err := copyFile(src, filepath.Join(dest, "pmtop.exe")); err != nil {
		return errMsg(err)
	}

	// Update PATH (Windows only for now)
	if runtime.GOOS == "windows" {
		cmd := fmt.Sprintf(`if ($env:Path -notlike "*%s*") { [Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";%s", "User") }`, dest, dest)
		exec.Command("powershell", "-Command", cmd).Run()
	}

	return doneMsg{}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func main() {
	p := tea.NewProgram(model{
		installPath: filepath.Join(os.Getenv("LOCALAPPDATA"), "pmtop-lite"),
	})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Installer error: %v\n", err)
		os.Exit(1)
	}
}
