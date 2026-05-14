//go:build windows

package scanner

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	iphlpapi            = windows.NewLazySystemDLL("iphlpapi.dll")
	getExtendedTcpTable = iphlpapi.NewProc("GetExtendedTcpTable")
	getExtendedUdpTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

const (
	// TCP_TABLE_OWNER_PID_LISTENER (4) — only LISTEN rows; fast and clean.
	// TCP_TABLE_OWNER_PID_ALL       (5) — all connections including ESTABLISHED.
	tcpTableOwnerPIDListener = 4
	tcpTableOwnerPIDAll      = 5
	udpTableOwnerPID         = 1 // UDP_TABLE_OWNER_PID
	afINET                   = 2 // AF_INET (IPv4)
)

// tcpStates maps Windows MIB state codes to readable strings.
var tcpStates = map[uint32]string{
	1: "CLOSED", 2: "LISTEN", 3: "SYN_SENT", 4: "SYN_RCVD",
	5: "ESTABLISHED", 6: "FIN_WAIT1", 7: "FIN_WAIT2", 8: "CLOSE_WAIT",
	9: "CLOSING", 10: "LAST_ACK", 11: "TIME_WAIT", 12: "DELETE_TCB",
}

// swapPort converts a Windows network-byte-order port DWORD to host order.
func swapPort(raw uint32) int {
	return int((raw>>8)&0xFF | (raw&0xFF)<<8)
}

// WindowsScanner implements Scanner using the Windows IP Helper API.
// opts.ListenOnly=true (default) shows only LISTEN-state TCP — real servers only.
type WindowsScanner struct {
	opts Options
}

func NewWindowsScanner() *WindowsScanner {
	return &WindowsScanner{opts: DefaultOptions()}
}

// NewWindowsScannerWithOptions creates a scanner with custom options.
func NewWindowsScannerWithOptions(opts Options) *WindowsScanner {
	return &WindowsScanner{opts: opts}
}

func (s *WindowsScanner) Scan() ([]PortEntry, error) {
	tableClass := uint32(tcpTableOwnerPIDListener)
	if !s.opts.ListenOnly {
		tableClass = tcpTableOwnerPIDAll
	}

	tcp, err := scanTCP(tableClass)
	if err != nil {
		return nil, fmt.Errorf("TCP scan: %w", err)
	}
	udp, err := scanUDP()
	if err != nil {
		udp = nil // non-fatal
	}
	return append(tcp, udp...), nil
}

// scanTCP calls GetExtendedTcpTable with the given tableClass.
func scanTCP(tableClass uint32) ([]PortEntry, error) {
	var size uint32
	getExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, afINET, uintptr(tableClass), 0)
	if size == 0 {
		size = 4096
	}

	buf := make([]byte, size)
	ret, _, _ := getExtendedTcpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		1, afINET, uintptr(tableClass), 0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable returned %d", ret)
	}

	// MIB_TCPTABLE_OWNER_PID: DWORD numEntries, then N rows of 24 bytes each
	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	const rowSize = 24
	offset := 4

	entries := make([]PortEntry, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		if offset+rowSize > len(buf) {
			break
		}
		row := buf[offset : offset+rowSize]
		state := *(*uint32)(unsafe.Pointer(&row[0]))
		rawPort := *(*uint32)(unsafe.Pointer(&row[8]))
		pid := *(*uint32)(unsafe.Pointer(&row[20]))

		stateName := tcpStates[state]
		if stateName == "" {
			stateName = "UNKNOWN"
		}
		entries = append(entries, PortEntry{
			Port:     swapPort(rawPort),
			Protocol: "TCP",
			PID:      int32(pid),
			State:    stateName,
		})
		offset += rowSize
	}
	return entries, nil
}

// scanUDP calls GetExtendedUdpTable and parses MIB_UDPTABLE_OWNER_PID.
func scanUDP() ([]PortEntry, error) {
	var size uint32
	getExtendedUdpTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, afINET, udpTableOwnerPID, 0)
	if size == 0 {
		size = 4096
	}

	buf := make([]byte, size)
	ret, _, _ := getExtendedUdpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		1, afINET, udpTableOwnerPID, 0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedUdpTable returned %d", ret)
	}

	// MIB_UDPTABLE_OWNER_PID: DWORD numEntries, then N rows of 12 bytes each
	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	const rowSize = 12
	offset := 4

	entries := make([]PortEntry, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		if offset+rowSize > len(buf) {
			break
		}
		row := buf[offset : offset+rowSize]
		rawPort := *(*uint32)(unsafe.Pointer(&row[4]))
		pid := *(*uint32)(unsafe.Pointer(&row[8]))

		entries = append(entries, PortEntry{
			Port:     swapPort(rawPort),
			Protocol: "UDP",
			PID:      int32(pid),
			State:    "—",
		})
		offset += rowSize
	}
	return entries, nil
}
