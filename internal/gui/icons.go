package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// GetServiceIcon returns a Fyne resource icon based on the service name.
func GetServiceIcon(service string) fyne.Resource {
	s := strings.ToLower(service)
	switch {
	case strings.Contains(s, "node") || strings.Contains(s, "next"):
		return theme.SettingsIcon() // Placeholder for Node
	case strings.Contains(s, "sql"):
		return theme.StorageIcon()
	case strings.Contains(s, "http") || strings.Contains(s, "web"):
		return theme.ComputerIcon()
	case strings.Contains(s, "redis") || strings.Contains(s, "cache"):
		return theme.HistoryIcon()
	case strings.Contains(s, "java") || strings.Contains(s, "tom"):
		return theme.FileApplicationIcon()
	case strings.Contains(s, "python") || strings.Contains(s, "django"):
		return theme.DocumentCreateIcon()
	case strings.Contains(s, "docker"):
		return theme.ViewFullScreenIcon()
	default:
		return theme.HelpIcon()
	}
}
