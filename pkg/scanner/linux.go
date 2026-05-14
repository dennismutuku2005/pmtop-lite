//go:build !windows

package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LinuxScanner reads /proc/net/tcp and /proc/net/udp.
type LinuxScanner struct {
	opts       Options
	inodeCache map[string]int32
}

func NewLinuxScanner() *LinuxScanner { return &LinuxScanner{opts: DefaultOptions()} }

// NewPlatformScanner returns a Linux/macOS scanner with default options.
func NewPlatformScanner() Scanner {
	return NewLinuxScanner()
}

// NewPlatformScannerWithOptions returns a Linux/macOS scanner with custom options.
func NewPlatformScannerWithOptions(opts Options) Scanner {
	return &LinuxScanner{opts: opts}
}

func (s *LinuxScanner) Scan() ([]PortEntry, error) {
	s.inodeCache = buildInodeCache()
	var all []PortEntry

	for _, p := range []struct{ file, proto string }{
		{"/proc/net/tcp", "TCP"}, {"/proc/net/tcp6", "TCP"},
		{"/proc/net/udp", "UDP"}, {"/proc/net/udp6", "UDP"},
	} {
		entries, err := parseNetFile(p.file, p.proto, s.inodeCache, s.opts.ListenOnly)
		if err == nil {
			all = append(all, entries...)
		}
	}
	return all, nil
}

func parseNetFile(path, proto string, inodes map[string]int32, listenOnly bool) ([]PortEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []PortEntry
	sc := bufio.NewScanner(f)
	sc.Scan() // skip header
	for sc.Scan() {
		line := strings.Fields(sc.Text())
		if len(line) < 10 {
			continue
		}
		parts := strings.Split(line[1], ":")
		if len(parts) != 2 {
			continue
		}
		portVal, err := strconv.ParseUint(parts[1], 16, 16)
		if err != nil {
			continue
		}
		stateVal, _ := strconv.ParseUint(line[3], 16, 8)
		if proto == "TCP" && listenOnly && stateVal != 0x0A {
			continue
		}
		inode := line[9]
		pid := inodes[inode]
		stateName := "LISTEN"
		if proto == "UDP" {
			stateName = "—"
		}
		entries = append(entries, PortEntry{
			Port: int(portVal), Protocol: proto, PID: pid, State: stateName,
		})
	}
	return entries, nil
}

func buildInodeCache() map[string]int32 {
	cache := make(map[string]int32)
	procs, _ := filepath.Glob("/proc/[0-9]*/fd")
	for _, fdDir := range procs {
		p := strings.Split(fdDir, "/")
		if len(p) < 3 {
			continue
		}
		pid, err := strconv.ParseInt(p[2], 10, 32)
		if err != nil {
			continue
		}
		links, _ := filepath.Glob(fdDir + "/*")
		for _, link := range links {
			target, err := os.Readlink(link)
			if err != nil {
				continue
			}
			if strings.HasPrefix(target, "socket:[") {
				inode := strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")
				cache[inode] = int32(pid)
			}
		}
	}
	return cache
}

var _ Scanner = (*LinuxScanner)(nil)
