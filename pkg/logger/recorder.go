package logger

import (
	"fmt"
	"time"
)

type EventType string

const (
	EventOpen  EventType = "OPEN"
	EventClose EventType = "CLOSE"
	EventMove  EventType = "MOVE" // Process changed on same port
)

type PortEvent struct {
	Timestamp time.Time
	Port      int
	Protocol  string
	Type      EventType
	Process   string
	PID       int32
}

func (e PortEvent) String() string {
	return fmt.Sprintf("[%s] %s: Port %d/%s by %s (PID %d)", 
		e.Timestamp.Format("15:04:05"), e.Type, e.Port, e.Protocol, e.Process, e.PID)
}

// History stores a list of recent port events.
type History struct {
	Events []PortEvent
	max    int
}

func NewHistory(maxSize int) *History {
	return &History{
		Events: make([]PortEvent, 0),
		max:    maxSize,
	}
}

func (h *History) Add(e PortEvent) {
	h.Events = append([]PortEvent{e}, h.Events...) // Prepend
	if len(h.Events) > h.max {
		h.Events = h.Events[:h.max]
	}
}
