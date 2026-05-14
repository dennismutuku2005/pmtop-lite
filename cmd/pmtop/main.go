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
	"github.com/dennismutuku2005/pmtop-lite/pkg/services"
)


const version = "0.1.0"

func main() {
	if len(os.Args) == 1 {

		printHelp()
		os.Exit(0)
	}

	firstArg := os.Args[1]
	
	// If first arg is a number, assume it's a port check: pmtop 3000
	if port, err := strconv.Atoi(firstArg); err == nil {
		checkPort(port)
		os.Exit(0)
	}

	switch firstArg {
	case "dash":
		showAllInTUI := false
		for _, arg := range os.Args[2:] {
			if arg == "--all" || arg == "-a" {
				showAllInTUI = true
				break
			}
		}
		p := tea.NewProgram(tui.NewModel(showAllInTUI))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "pmtop error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		showAllInList := false
		for _, arg := range os.Args[2:] {
			if arg == "--all" || arg == "-a" {
				showAllInList = true
				break
			}
		}
		runListMode(showAllInList)
	case "close":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please specify a port (e.g. pmtop close 3000)")
			os.Exit(1)
		}
		closePort(os.Args[2])
	case "log":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please specify a port (e.g. pmtop log 8080)")
			os.Exit(1)
		}
		watchPort(os.Args[2])
	case "help":
		printHelp()
	case "--version", "-v":
		fmt.Printf("pmtop v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	case "--json":
		runJSONMode(false) // Default to clean JSON
	case "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command '%s'. Run 'pmtop help' for usage.\n", firstArg)
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

// checkPort checks a single port and prints its status.
func checkPort(port int) {
	sc := scanner.NewPlatformScannerWithOptions(scanner.Options{ListenOnly: false})
	ports, _ := sc.Scan()
	for _, p := range ports {
		if p.Port == port {
			info := proc.Resolve(p.PID)
			service := services.Detect(p.Port, info.Name)
			fmt.Printf("● Port %d is IN USE by '%s' (%s, PID %d)\n", port, info.Name, service, p.PID)
			return
		}
	}
	fmt.Printf("○ Port %d is FREE\n", port)
}

// runListMode prints a non-interactive list of active ports.
func runListMode(showAll bool) {
	sc := scanner.NewPlatformScannerWithOptions(scanner.Options{ListenOnly: true})
	ports, _ := sc.Scan()
	
	fmt.Printf("%-7s  %-15s  %-15s  %-7s\n", "PORT", "PROCESS", "SERVICE", "PID")
	fmt.Println(strings.Repeat("─", 50))
	
	count := 0
	for _, p := range ports {
		info := proc.Resolve(p.PID)
		service := services.Detect(p.Port, info.Name)
		
		// Apply Developer Intelligence Filter unless --all is passed
		if !showAll && !services.IsDev(p.Port, service, info.Name) {
			continue
		}
		
		fmt.Printf(":%-6d  %-15s  %-15s  %-7d\n", p.Port, info.Name, service, p.PID)
		count++
	}
	if count == 0 {
		fmt.Println("No active developer ports found. Use 'pmtop list --all' to see everything.")
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
	fmt.Printf(`
  PMTOP v%s — Professional Developer Port Intelligence
  ─────────────────────────────────────────────────────
  pmtop is a high-performance utility designed to help 
  developers manage their local stack with laser focus.

  COMMANDS:
    pmtop              Show this information screen
    pmtop dash         Launch the interactive dashboard (TUI)
    pmtop <port>       Quick check if a port is in use (e.g. pmtop 3306)
    pmtop list         Show a clean list of active developer ports
    pmtop list --all   Show ALL active system ports (includes noise)
    pmtop close <port> Safely kill the process using a specific port
    pmtop log <port>   Watch real-time activity on a specific port
    pmtop help         Show this guide
    pmtop --version    Show version info

  FEATURES:
    ● Auto-detects Node, MySQL, Oracle, Java, Tomcat, etc.
    ● Automatically silences Chrome, Spotify, and System noise.
    ● Native Windows support for deep process resolution.

  TUI SHORTCUTS (within 'pmtop dash'):
    [a] Toggle Noise Filter    [x] Kill Process
    [/] Search/Filter          [s] Cycle Sort
    [o] Open in Browser        [q] Quit
`, version)
}



