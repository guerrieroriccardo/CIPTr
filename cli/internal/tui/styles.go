package tui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle is used for screen titles and headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // light blue
			MarginBottom(1)

	// BreadcrumbStyle is used for the navigation breadcrumb.
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // gray
			MarginBottom(1)

	// HelpStyle is used for the bottom help bar.
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// ErrorStyle is used for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // red
			Bold(true)

	// SuccessStyle is used for success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")). // green
			Bold(true)
)
