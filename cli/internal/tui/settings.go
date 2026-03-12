package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

type settingsLoadedMsg struct {
	values map[string]string
}

type settingsSavedMsg struct{ key string }
type settingsErrorMsg struct{ err error }

// settingField describes one configurable setting.
type settingField struct {
	key     string
	label   string
	options []string // selectable values
}

var hostnameSettings = []settingField{
	{key: "hostname_prefix_source", label: "Prefix Source", options: []string{"short_code", "name"}},
	{key: "hostname_prefix_position", label: "Prefix Position", options: []string{"before", "after", "none"}},
	{key: "hostname_num_digits", label: "Number Digits", options: []string{"1", "2", "3", "4", "5", "6"}},
}

// SettingsScreen allows admins to configure hostname nomenclature.
type SettingsScreen struct {
	client *apiclient.Client
	values map[string]string // current setting values
	focus  int               // index into hostnameSettings
	err    error
	loaded bool
	saved  string // flash message
}

func NewSettingsScreen(client *apiclient.Client) *SettingsScreen {
	return &SettingsScreen{
		client: client,
		values: make(map[string]string),
	}
}

func (s *SettingsScreen) Title() string { return "Settings" }

func (s *SettingsScreen) Init() tea.Cmd {
	client := s.client
	return func() tea.Msg {
		settings, err := client.GetSettings()
		if err != nil {
			return settingsErrorMsg{err: err}
		}
		vals := make(map[string]string)
		for _, st := range settings {
			vals[st.Key] = st.Value
		}
		return settingsLoadedMsg{values: vals}
	}
}

func (s *SettingsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsLoadedMsg:
		s.values = msg.values
		s.loaded = true
		s.err = nil
		return s, nil

	case settingsSavedMsg:
		s.saved = "Saved: " + msg.key
		s.err = nil
		return s, nil

	case settingsErrorMsg:
		s.err = msg.err
		s.loaded = true
		return s, nil

	case tea.KeyMsg:
		s.saved = ""

		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "tab", "down", "j":
			s.focus = (s.focus + 1) % len(hostnameSettings)
			return s, nil
		case "shift+tab", "up", "k":
			s.focus = (s.focus - 1 + len(hostnameSettings)) % len(hostnameSettings)
			return s, nil
		case "left", "h":
			return s, s.cycleOption(-1)
		case "right", "l":
			return s, s.cycleOption(1)
		case "enter":
			return s, s.cycleOption(1)
		}
	}
	return s, nil
}

// cycleOption changes the current setting value to the next/previous option and saves it.
func (s *SettingsScreen) cycleOption(dir int) tea.Cmd {
	if !s.loaded {
		return nil
	}
	field := hostnameSettings[s.focus]
	current := s.values[field.key]

	idx := 0
	for i, opt := range field.options {
		if opt == current {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(field.options)) % len(field.options)
	newVal := field.options[idx]
	s.values[field.key] = newVal

	client := s.client
	key := field.key
	return func() tea.Msg {
		if err := client.UpdateSetting(key, newVal); err != nil {
			return settingsErrorMsg{err: err}
		}
		return settingsSavedMsg{key: key}
	}
}

func (s *SettingsScreen) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Hostname Nomenclature") + "\n\n")

	if !s.loaded {
		b.WriteString("Loading settings...")
		return b.String()
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	for i, field := range hostnameSettings {
		cursor := "  "
		if i == s.focus {
			cursor = "> "
		}

		b.WriteString(cursor + field.label + ": ")
		for j, opt := range field.options {
			if j > 0 {
				b.WriteString("  ")
			}
			if opt == s.values[field.key] {
				b.WriteString(selectedStyle.Render(" " + opt + " "))
			} else {
				b.WriteString(optionStyle.Render(" " + opt + " "))
			}
		}
		b.WriteString("\n\n")
	}

	// Preview
	b.WriteString(s.renderPreview())
	b.WriteString("\n")

	if s.err != nil {
		b.WriteString(ErrorStyle.Render(s.err.Error()) + "\n")
	}
	if s.saved != "" {
		b.WriteString(SuccessStyle.Render(s.saved) + "\n")
	}

	b.WriteString("\n" + HelpStyle.Render("←/→ change value • ↑/↓ navigate • esc back"))
	return b.String()
}

// renderPreview shows example hostnames with the current settings.
func (s *SettingsScreen) renderPreview() string {
	source := s.values["hostname_prefix_source"]
	position := s.values["hostname_prefix_position"]
	digits := s.values["hostname_num_digits"]

	if digits == "" {
		digits = "3"
	}

	// Build example with a sample category.
	var prefix string
	switch source {
	case "name":
		prefix = "Notebook"
	default:
		prefix = "NB"
	}

	numFmt := fmt.Sprintf("%%0%ss", digits)
	numStr := fmt.Sprintf(numFmt, "1")

	var example string
	switch position {
	case "after":
		example = numStr + prefix
	case "none":
		example = numStr
	default:
		example = prefix + numStr
	}

	previewStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	return fmt.Sprintf("  Preview: %s  (e.g. %s, %s)",
		previewStyle.Render(example),
		previewStyle.Render(buildPreviewExample("SRV", "Server", source, position, numFmt)),
		previewStyle.Render(buildPreviewExample("PC", "PCDesktop", source, position, numFmt)),
	)
}

func buildPreviewExample(shortCode, name, source, position, numFmt string) string {
	prefix := shortCode
	if source == "name" {
		prefix = name
	}
	numStr := fmt.Sprintf(numFmt, "1")
	switch position {
	case "after":
		return numStr + prefix
	case "none":
		return numStr
	default:
		return prefix + numStr
	}
}
