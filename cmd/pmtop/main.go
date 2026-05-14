package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/dennismutuku2005/pmtop-lite/internal/tui"
	"github.com/dennismutuku2005/pmtop-lite/pkg/proc"
	"github.com/dennismutuku2005/pmtop-lite/pkg/scanner"
)

const version = "0.1.0"

func main() {
	showAll := false
	for _, arg := range os.Args[1:] {
		if arg == "--all" || arg == "-a" {
			showAll = true
			break
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "close":
			if len(os.Args) < 3 {
				fmt.Println("Error: Please specify a port (e.g. pmtop close 3000)")
				os.Exit(1)
			}
			closePort(os.Args[2])
			os.Exit(0)
		case "log":
			if len(os.Args) < 3 {
				fmt.Println("Error: Please specify a port (e.g. pmtop log 8080)")
				os.Exit(1)
			}
			watchPort(os.Args[2])
			os.Exit(0)
		case "--version", "-v":
			fmt.Printf("pmtop v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
			os.Exit(0)
		case "--json":
			runJSONMode(showAll)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	p := tea.NewProgram(tui.NewModel(showAll))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pmtop error: %v\n", err)
		os.Exit(1)
	}
}

// runJSONMode scans ports once and prints JSON — useful for scripting.
func runJSONMode(showAll bool) {
	opts := scanner.DefaultOptions()
	opts.ListenOnly = !showAll
	sc := scanner.NewPlatformScannerWithOptions(opts)
	ports, err := sc.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ports); err != nil {
		fmt.Fprintf(os.Stderr, "json error: %v\n", err)
		os.Exit(1)
	}
}

func closePort(portStr string) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Error: Invalid port '%s'\n", portStr)
		os.Exit(1)
	}

	sc := scanner.NewPlatformScannerWithOptions(scanner.Options{ListenOnly: false})
	ports, err := sc.Scan()
	if err != nil {
		fmt.Printf("Error scanning ports: %v\n", err)
		os.Exit(1)
	}

	var found *scanner.PortEntry
	for _, p := range ports {
		if p.Port == port {
			found = &p
			break
		}
	}

	if found == nil {
		fmt.Printf("No process found using port %d\n", port)
		return
	}

	// Resolve process info
	info := proc.Resolve(found.PID)
	fmt.Printf("Found process '%s' (PID %d) using port %d.\n", info.Name, found.PID, port)
	fmt.Print("Are you sure you want to close it? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		p, err := os.FindProcess(int(found.PID))
		if err == nil {
			err = p.Kill()
		}
		if err != nil {
			fmt.Printf("Error: Could not close process: %v\n", err)
		} else {
			fmt.Printf("Success: Port %d has been released. Process '%s' was closed.\n", port, info.Name)
		}
	} else {
		fmt.Println("Cancelled.")
	}
}

func printHelp() {
	fmt.Printf(`pmtop v%s — Premium Developer Port Monitor

Usage:
  pmtop              Start the dashboard (Developer Ports only by default)
  pmtop --all        Start the dashboard showing ALL system ports
  pmtop log <port>   Watch activity on a specific port in real-time
  pmtop close <port> Kill the process using the specified port
  pmtop --json       Print open ports as JSON and exit
  pmtop --version    Show version
  pmtop --help       Show this help

TUI Shortcuts:
  [a] Toggle Developer Ports vs All
  [x] Kill process (with confirmation)
  [/] Search/Filter
  [s] Cycle Sort
  [o] Open in browser
  [q] Quit
`, version)
}

