package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/auth"
)

type loginSuccessMsg struct{}
type loginErrorMsg struct{ err error }

type LoginScreen struct {
	client    *apiclient.Client
	serverURL textinput.Model
	username  textinput.Model
	password  textinput.Model
	focus     int // 0=serverURL, 1=username, 2=password
	err       error
	loading   bool
	width     int
	anon      bool // anonymous mode: don't persist token/server URL
}

func NewLoginScreen(client *apiclient.Client) *LoginScreen {
	return newLoginScreen(client, false)
}

// NewAnonLoginScreen creates a login screen that does not persist the token or server URL.
func NewAnonLoginScreen(client *apiclient.Client) *LoginScreen {
	return newLoginScreen(client, true)
}

func newLoginScreen(client *apiclient.Client, anon bool) *LoginScreen {
	s := textinput.New()
	s.Placeholder = "https://your-server.example.com"
	s.Focus()
	s.CharLimit = 256
	s.Width = 40
	// Pre-fill with the current base URL (strip /api/v1 suffix for cleaner display).
	s.SetValue(strings.TrimSuffix(client.BaseURL, "/api/v1"))

	u := textinput.New()
	u.Placeholder = "username"
	u.CharLimit = 64
	u.Width = 40

	p := textinput.New()
	p.Placeholder = "password (empty for guest)"
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '*'
	p.CharLimit = 128
	p.Width = 40

	return &LoginScreen{
		client:    client,
		serverURL: s,
		username:  u,
		password:  p,
		anon:      anon,
	}
}

func (l *LoginScreen) Title() string { return "Login" }

func (l *LoginScreen) Init() tea.Cmd { return textinput.Blink }

func (l *LoginScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		inputWidth := min(50, msg.Width-20)
		if inputWidth < 20 {
			inputWidth = 20
		}
		l.serverURL.Width = inputWidth
		l.username.Width = inputWidth
		l.password.Width = inputWidth
		return l, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			l.focus = (l.focus + 1) % 3
			l.serverURL.Blur()
			l.username.Blur()
			l.password.Blur()
			switch l.focus {
			case 0:
				l.serverURL.Focus()
			case 1:
				l.username.Focus()
			case 2:
				l.password.Focus()
			}
			return l, nil
		case "shift+tab", "up":
			l.focus = (l.focus + 2) % 3
			l.serverURL.Blur()
			l.username.Blur()
			l.password.Blur()
			switch l.focus {
			case 0:
				l.serverURL.Focus()
			case 1:
				l.username.Focus()
			case 2:
				l.password.Focus()
			}
			return l, nil
		case "enter":
			if l.loading {
				return l, nil
			}
			// Advance through fields until all are filled.
			srv := strings.TrimRight(strings.TrimSpace(l.serverURL.Value()), "/")
			if srv == "" {
				l.err = fmt.Errorf("server URL is required")
				return l, nil
			}
			if l.focus < 1 {
				l.focus = 1
				l.serverURL.Blur()
				l.username.Focus()
				return l, nil
			}
			u := strings.TrimSpace(l.username.Value())
			if u == "" {
				l.err = fmt.Errorf("username is required")
				if l.focus != 1 {
					l.focus = 1
					l.serverURL.Blur()
					l.password.Blur()
					l.username.Focus()
				}
				return l, nil
			}
			p := l.password.Value()
			l.loading = true
			l.err = nil
			// Update client base URL to what the user entered.
			l.client.BaseURL = srv + "/api/v1"
			client := l.client
			serverURL := srv
			anon := l.anon
			if p == "" {
				return l, func() tea.Msg {
					token, err := client.GuestLogin(u)
					if err != nil {
						return loginErrorMsg{err: err}
					}
					client.Token = token
					if !anon {
						_ = auth.SaveToken(token)
						_ = auth.SaveServerURL(serverURL)
					}
					return loginSuccessMsg{}
				}
			}
			return l, func() tea.Msg {
				token, err := client.Login(u, p)
				if err != nil {
					return loginErrorMsg{err: err}
				}
				client.Token = token
				if !anon {
					_ = auth.SaveToken(token)
					_ = auth.SaveServerURL(serverURL)
				}
				return loginSuccessMsg{}
			}
		}

	case loginSuccessMsg:
		l.loading = false
		return l, func() tea.Msg {
			return PushScreenMsg{Screen: NewMenu()}
		}

	case loginErrorMsg:
		l.loading = false
		l.err = msg.err
		return l, nil
	}

	var cmd tea.Cmd
	switch l.focus {
	case 0:
		l.serverURL, cmd = l.serverURL.Update(msg)
	case 1:
		l.username, cmd = l.username.Update(msg)
	case 2:
		l.password, cmd = l.password.Update(msg)
	}
	return l, cmd
}

func (l *LoginScreen) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("CIPTr Login"))
	b.WriteString("\n\n")
	b.WriteString("Server:   " + l.serverURL.View())
	b.WriteString("\n")
	b.WriteString("Username: " + l.username.View())
	b.WriteString("\n")
	b.WriteString("Password: " + l.password.View())
	b.WriteString("\n\n")

	if l.loading {
		b.WriteString("Authenticating...")
	} else if l.err != nil {
		b.WriteString(ErrorStyle.Render(l.err.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("tab: next field • enter: login (empty password = guest) • ctrl+c: quit"))
	return b.String()
}
