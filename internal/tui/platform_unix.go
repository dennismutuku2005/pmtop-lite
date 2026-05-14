//go:build !windows

package tui

import (
	"os/exec"
	"runtime"
)

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	_ = cmd.Start()
	_, _ = pipe.Write([]byte(text))
	pipe.Close()
	_ = cmd.Wait()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
