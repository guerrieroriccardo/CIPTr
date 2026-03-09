package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

type passwordChangedMsg struct{}
type passwordErrorMsg struct{ err error }

type ChangePasswordScreen struct {
	client      *apiclient.Client
	oldPassword textinput.Model
	newPassword textinput.Model
	confirm     textinput.Model
	focus       int // 0=old, 1=new, 2=confirm
	err         error
	saving      bool
	done        bool
}

func NewChangePasswordScreen(client *apiclient.Client) *ChangePasswordScreen {
	old := textinput.New()
	old.Placeholder = "current password"
	old.EchoMode = textinput.EchoPassword
	old.EchoCharacter = '*'
	old.CharLimit = 128
	old.Width = 30
	old.Focus()

	np := textinput.New()
	np.Placeholder = "new password"
	np.EchoMode = textinput.EchoPassword
	np.EchoCharacter = '*'
	np.CharLimit = 128
	np.Width = 30

	conf := textinput.New()
	conf.Placeholder = "confirm new password"
	conf.EchoMode = textinput.EchoPassword
	conf.EchoCharacter = '*'
	conf.CharLimit = 128
	conf.Width = 30

	return &ChangePasswordScreen{
		client:      client,
		oldPassword: old,
		newPassword: np,
		confirm:     conf,
	}
}

func (s *ChangePasswordScreen) Title() string { return "Change Password" }

func (s *ChangePasswordScreen) Init() tea.Cmd { return textinput.Blink }

func (s *ChangePasswordScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "tab", "down":
			return s, s.nextField()
		case "shift+tab", "up":
			return s, s.prevField()
		case "enter":
			if s.saving || s.done {
				return s, nil
			}
			return s, s.submit()
		}

	case passwordChangedMsg:
		s.saving = false
		s.done = true
		s.err = nil
		return s, nil

	case passwordErrorMsg:
		s.saving = false
		s.err = msg.err
		return s, nil
	}

	var cmd tea.Cmd
	switch s.focus {
	case 0:
		s.oldPassword, cmd = s.oldPassword.Update(msg)
	case 1:
		s.newPassword, cmd = s.newPassword.Update(msg)
	case 2:
		s.confirm, cmd = s.confirm.Update(msg)
	}
	return s, cmd
}

func (s *ChangePasswordScreen) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Change Password"))
	b.WriteString("\n\n")
	b.WriteString("Old Password: " + s.oldPassword.View())
	b.WriteString("\n")
	b.WriteString("New Password: " + s.newPassword.View())
	b.WriteString("\n")
	b.WriteString("    Confirm:  " + s.confirm.View())
	b.WriteString("\n\n")

	if s.saving {
		b.WriteString("Saving...")
	} else if s.done {
		b.WriteString(SuccessStyle.Render("Password changed successfully! Press esc to go back."))
	} else if s.err != nil {
		b.WriteString(ErrorStyle.Render(s.err.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("tab: next field • enter: submit • esc: back"))
	return b.String()
}

func (s *ChangePasswordScreen) nextField() tea.Cmd {
	s.blurAll()
	s.focus = (s.focus + 1) % 3
	s.focusCurrent()
	return nil
}

func (s *ChangePasswordScreen) prevField() tea.Cmd {
	s.blurAll()
	s.focus = (s.focus + 2) % 3
	s.focusCurrent()
	return nil
}

func (s *ChangePasswordScreen) blurAll() {
	s.oldPassword.Blur()
	s.newPassword.Blur()
	s.confirm.Blur()
}

func (s *ChangePasswordScreen) focusCurrent() {
	switch s.focus {
	case 0:
		s.oldPassword.Focus()
	case 1:
		s.newPassword.Focus()
	case 2:
		s.confirm.Focus()
	}
}

func (s *ChangePasswordScreen) submit() tea.Cmd {
	oldPw := s.oldPassword.Value()
	newPw := s.newPassword.Value()
	confirmPw := s.confirm.Value()

	if oldPw == "" || newPw == "" || confirmPw == "" {
		s.err = fmt.Errorf("all fields are required")
		return nil
	}
	if newPw != confirmPw {
		s.err = fmt.Errorf("new passwords do not match")
		return nil
	}

	s.saving = true
	s.err = nil
	client := s.client
	return func() tea.Msg {
		var result map[string]string
		err := client.Put("/change-password", map[string]string{
			"old_password": oldPw,
			"new_password": newPw,
		}, &result)
		if err != nil {
			return passwordErrorMsg{err: err}
		}
		return passwordChangedMsg{}
	}
}
