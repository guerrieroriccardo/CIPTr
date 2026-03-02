package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen is a navigable view in the TUI.
type Screen interface {
	tea.Model
	Title() string
}

// NavStack manages a stack of screens for drill-down navigation.
type NavStack struct {
	stack []Screen
}

// Push adds a screen to the top of the stack.
func (n *NavStack) Push(s Screen) {
	n.stack = append(n.stack, s)
}

// Pop removes and returns the top screen. Returns nil if only one screen remains.
func (n *NavStack) Pop() Screen {
	if len(n.stack) <= 1 {
		return nil
	}
	top := n.stack[len(n.stack)-1]
	n.stack = n.stack[:len(n.stack)-1]
	return top
}

// Current returns the screen on top of the stack.
func (n *NavStack) Current() Screen {
	if len(n.stack) == 0 {
		return nil
	}
	return n.stack[len(n.stack)-1]
}

// Len returns the number of screens on the stack.
func (n *NavStack) Len() int {
	return len(n.stack)
}

// Breadcrumb returns a string like "Clients > Acme Corp > Sites".
func (n *NavStack) Breadcrumb() string {
	titles := make([]string, len(n.stack))
	for i, s := range n.stack {
		titles[i] = s.Title()
	}
	return strings.Join(titles, " > ")
}
