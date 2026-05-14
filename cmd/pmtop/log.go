package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dennismutuku2005/pmtop-lite/pkg/state"
)

func watchPort(portStr string) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Error: Invalid port '%s'\n", portStr)
		os.Exit(1)
	}

	fmt.Printf("Watching port %d for activity... (Ctrl+C to stop)\n", port)
	
	mgr := state.New(true) // Show all connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	go mgr.Run(ctx)

	lastEventCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping logger.")
			return
		case <-time.After(1 * time.Second):
			events := mgr.History.Events
			if len(events) > lastEventCount {
				for i := len(events) - 1; i >= lastEventCount; i-- {
					e := events[i]
					if e.Port == port {
						fmt.Println(e.String())
					}
				}
				lastEventCount = len(events)
			}
		}
	}
}
