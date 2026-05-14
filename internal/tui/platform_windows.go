// Platform utilities: clipboard and browser — Windows implementation.
//go:build windows

package tui

import (
	"os/exec"
)

func copyToClipboard(text string) {
	cmd := exec.Command("cmd", "/c", "echo", text, "|", "clip")
	_ = cmd.Run()
}

func openBrowser(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	_ = cmd.Start()
}
