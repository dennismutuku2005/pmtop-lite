// Package tui defines keyboard bindings for pmtop.
package tui

// keyAction represents a named keyboard action with its bound keys.
type keyAction struct {
	keys []string
	help string
}

// keyMap holds all key bindings.
type keyMap struct {
	Up     keyAction
	Down   keyAction
	Kill   keyAction
	Filter keyAction
	Sort   keyAction
	Copy   keyAction
	Open   keyAction
	Quit   keyAction
}

// defaultKeys returns the default key bindings.
// NOTE: "x" kills (not "k") to avoid conflicting with vim-up navigation.
// "/" activates filter mode like vim search.
var defaultKeys = keyMap{
	Up:     keyAction{keys: []string{"up", "k"}, help: "↑/k  up"},
	Down:   keyAction{keys: []string{"down", "j"}, help: "↓/j  down"},
	Kill:   keyAction{keys: []string{"x"}, help: "x  kill"},
	Filter: keyAction{keys: []string{"/"}, help: "/  filter"},
	Sort:   keyAction{keys: []string{"s"}, help: "s  sort"},
	Copy:   keyAction{keys: []string{"c"}, help: "c  copy"},
	Open:   keyAction{keys: []string{"o"}, help: "o  browser"},
	Quit:   keyAction{keys: []string{"q", "ctrl+c"}, help: "q  quit"},
}

// contains checks if a key string is in an action's keys list.
func (a keyAction) contains(s string) bool {
	for _, k := range a.keys {
		if k == s {
			return true
		}
	}
	return false
}
