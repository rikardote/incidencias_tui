package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/styles"
)

type LoginSuccessMsg struct {
	Client *api.Client
	Cfg    *config.Config
}
type LoginErrorMsg string

type loginField int

const (
	fieldUsername loginField = iota
	fieldPassword
	fieldURL
)

type LoginModel struct {
	client   *api.Client
	cfg      *config.Config
	username textinput.Model
	password textinput.Model
	apiURL   textinput.Model
	focus    loginField
	loading  bool
	errorMsg string
	showURL  bool
}

func NewLoginModel(cfg *config.Config) LoginModel {
	ui := textinput.New()
	ui.Placeholder = "usuario"
	ui.Prompt = ""
	ui.Focus()
	ui.TextStyle = styles.InputFocused
	ui.CharLimit = 64
	ui.Width = 40

	pi := textinput.New()
	pi.Placeholder = "contraseña"
	pi.Prompt = ""
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.TextStyle = styles.InputBox
	pi.CharLimit = 128
	pi.Width = 40

	au := textinput.New()
	au.Placeholder = "http://localhost:8190"
	au.Prompt = ""
	au.TextStyle = styles.InputBox
	au.CharLimit = 256
	au.SetValue(cfg.APIURL)
	au.Width = 40

	return LoginModel{
		client:   api.New(cfg.APIURL),
		cfg:      cfg,
		username: ui,
		password: pi,
		apiURL:   au,
		focus:    fieldUsername,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyTab, tea.KeyDown:
			m.nextField()

		case tea.KeyShiftTab, tea.KeyUp:
			m.prevField()

		case tea.KeyEnter:
			if m.focus == fieldPassword || m.focus == fieldUsername {
				if m.focus == fieldUsername {
					m.focus = fieldPassword
					m.updateFocus()
					return m, nil
				}
				return m, m.doLogin()
			} else if m.focus == fieldURL {
				m.showURL = false
				m.focus = fieldUsername
				m.updateFocus()
				return m, nil
			}

		case tea.KeyCtrlL:
			m.showURL = !m.showURL
			if m.showURL {
				m.focus = fieldURL
				m.updateFocus()
			}
			return m, nil
		}

	case LoginErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	var cmd tea.Cmd
	switch m.focus {
	case fieldUsername:
		m.username, cmd = m.username.Update(msg)
	case fieldPassword:
		m.password, cmd = m.password.Update(msg)
	case fieldURL:
		m.apiURL, cmd = m.apiURL.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m LoginModel) View() string {
	if m.showURL {
		return m.viewURL()
	}

	var s string

	// Logo
	s += "\n\n"
	s += styles.Title.Copy().Align(lipgloss.Center).Width(60).Render("⚡ INCIDENCIAS TUI")
	s += "\n"
	s += styles.Muted.Copy().Align(lipgloss.Center).Width(60).Render("Sistema de gestión de incidencias")
	s += "\n\n"

	// Form
	var form string
	form += styles.Label.Render("Usuario:") + " " + m.username.View() + "\n\n"
	form += styles.Label.Render("Contraseña:") + " " + m.password.View() + "\n"

	if m.errorMsg != "" {
		form += "\n" + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n"
	}

	if m.loading {
		form += "\n" + styles.InfoText.Render("● Iniciando sesión...") + "\n"
	}

	form += "\n"
	form += styles.Muted.Render("Tab: siguiente · Enter: entrar · Ctrl+L: API")

	card := styles.Panel.Copy().Width(60).Render(form)

	// Center card
	s += lipgloss.PlaceHorizontal(80, lipgloss.Center, card)
	s += "\n"

	return s
}

func (m LoginModel) viewURL() string {
	var s string
	s += "\n\n"
	s += styles.Title.Copy().Align(lipgloss.Center).Width(60).Render("Configurar API URL")
	s += "\n\n"

	var form string
	form += styles.Label.Render("URL:") + " " + m.apiURL.View() + "\n\n"
	form += styles.Muted.Render("Enter: confirmar · Esc: cancelar")

	card := styles.Panel.Copy().Width(60).Render(form)
	s += lipgloss.PlaceHorizontal(80, lipgloss.Center, card)

	return s
}

func (m *LoginModel) nextField() {
	switch m.focus {
	case fieldUsername:
		m.focus = fieldPassword
	case fieldPassword:
		if m.showURL {
			m.focus = fieldURL
		}
	case fieldURL:
		m.focus = fieldUsername
	}
	m.updateFocus()
}

func (m *LoginModel) prevField() {
	switch m.focus {
	case fieldPassword:
		m.focus = fieldUsername
	case fieldURL:
		m.focus = fieldPassword
	case fieldUsername:
		if m.showURL {
			m.focus = fieldURL
		}
	}
	m.updateFocus()
}

func (m *LoginModel) updateFocus() {
	m.username.Blur()
	m.password.Blur()
	m.apiURL.Blur()

	m.username.TextStyle = styles.InputBox
	m.password.TextStyle = styles.InputBox
	m.apiURL.TextStyle = styles.InputBox

	switch m.focus {
	case fieldUsername:
		m.username.Focus()
		m.username.TextStyle = styles.InputFocused
	case fieldPassword:
		m.password.Focus()
		m.password.TextStyle = styles.InputFocused
	case fieldURL:
		m.apiURL.Focus()
		m.apiURL.TextStyle = styles.InputFocused
	}
}

func (m *LoginModel) doLogin() tea.Cmd {
	username := m.username.Value()
	password := m.password.Value()

	if username == "" || password == "" {
		m.errorMsg = "Usuario y contraseña son requeridos"
		return nil
	}

	m.loading = true
	m.errorMsg = ""

	newURL := m.apiURL.Value()
	if newURL != "" && newURL != m.cfg.APIURL {
		m.cfg.APIURL = newURL
		m.client = api.New(newURL)
	}
	m.client.SetToken("")

	return func() tea.Msg {
		resp, err := m.client.Login(username, password, "incidencias-tui")
		if err != nil {
			return LoginErrorMsg(fmt.Sprintf("%v", err))
		}

		m.cfg.Token = resp.Token
		m.cfg.APIURL = m.client.GetBaseURL()
		if err := m.cfg.Save(); err != nil {
			return LoginErrorMsg(fmt.Sprintf("Error guardando config: %v", err))
		}

		return LoginSuccessMsg{
			Client: m.client,
			Cfg:    m.cfg,
		}
	}
}
