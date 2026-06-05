package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/styles"
)

// Login messages
type LoginSuccessMsg struct {
	Client *api.Client
	Cfg    *config.Config
}
type LoginErrorMsg string

// loginField identifies which field is focused
type loginField int

const (
	fieldUsername loginField = iota
	fieldPassword
	fieldURL
)

// LoginModel is the login screen
type LoginModel struct {
	client      *api.Client
	cfg         *config.Config
	username    textinput.Model
	password    textinput.Model
	apiURL      textinput.Model
	focus       loginField
	loading     bool
	errorMsg    string
	showURL     bool
}

// NewLoginModel creates a new login screen
func NewLoginModel(cfg *config.Config) LoginModel {
	ui := textinput.New()
	ui.Placeholder = "usuario"
	ui.Prompt = ""
	ui.Focus()
	ui.TextStyle = styles.InputFocusedStyle
	ui.CharLimit = 64

	pi := textinput.New()
	pi.Placeholder = "contraseña"
	pi.Prompt = ""
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.TextStyle = styles.InputStyle
	pi.CharLimit = 128

	au := textinput.New()
	au.Placeholder = "http://localhost:8190"
	au.Prompt = ""
	au.TextStyle = styles.InputStyle
	au.CharLimit = 256
	au.SetValue(cfg.APIURL)

	return LoginModel{
		client:   api.New(cfg.APIURL),
		cfg:      cfg,
		username: ui,
		password: pi,
		apiURL:   au,
		focus:    fieldUsername,
		showURL:  false,
	}
}

// Init implements tea.Model
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
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
				// If on username, go to password
				if m.focus == fieldUsername {
					m.focus = fieldPassword
					m.updateFocus()
					return m, nil
				}
				// If on password, try login
				return m, m.doLogin()
			} else if m.focus == fieldURL {
				m.showURL = false
				m.focus = fieldUsername
				m.updateFocus()
				return m, nil
			}

		case tea.KeyCtrlL:
			// Toggle URL editing
			m.showURL = !m.showURL
			if m.showURL {
				m.focus = fieldURL
				m.updateFocus()
			}
			return m, nil
		}

	case LoginSuccessMsg:
		return m, nil // The parent model will handle this

	case LoginErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update the focused input
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

// View implements tea.Model
func (m LoginModel) View() string {
	var s string

	s += styles.TitleStyle.Render("🔐 Inicio de Sesión")
	s += "\n\n"

	if m.showURL {
		s += styles.LabelStyle.Render("API URL:")
		s += m.apiURL.View()
		s += "\n\n"
		s += styles.HelpStyle.Render("Presiona Enter para confirmar, Tab para siguiente")
		return s
	}

	s += styles.LabelStyle.Render("Usuario:")
	s += m.username.View()
	s += "\n\n"
	s += styles.LabelStyle.Render("Contraseña:")
	s += m.password.View()
	s += "\n\n"

	if m.errorMsg != "" {
		s += styles.ErrorStyle.Render("✗ " + m.errorMsg)
		s += "\n\n"
	}

	if m.loading {
		s += styles.InfoStyle.Render("Iniciando sesión...")
	} else {
		s += styles.HelpStyle.Render("Tab/↓: siguiente · ↑: anterior · Enter: login")
		s += "\n"
		s += styles.HelpStyle.Render("Ctrl+L: cambiar URL · Esc/Ctrl+C: salir")
	}

	return styles.DocStyle.Render(s)
}

func (m *LoginModel) nextField() {
	switch m.focus {
	case fieldUsername:
		m.focus = fieldPassword
	case fieldPassword:
		if m.showURL {
			m.focus = fieldURL
		}
		// else stay on password
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

	m.username.TextStyle = styles.InputStyle
	m.password.TextStyle = styles.InputStyle
	m.apiURL.TextStyle = styles.InputStyle

	switch m.focus {
	case fieldUsername:
		m.username.Focus()
		m.username.TextStyle = styles.InputFocusedStyle
	case fieldPassword:
		m.password.Focus()
		m.password.TextStyle = styles.InputFocusedStyle
	case fieldURL:
		m.apiURL.Focus()
		m.apiURL.TextStyle = styles.InputFocusedStyle
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

	// Update API URL if changed
	newURL := m.apiURL.Value()
	if newURL != "" && newURL != m.cfg.APIURL {
		m.cfg.APIURL = newURL
		m.client = api.New(newURL)
	}
	m.client.SetToken("") // clear any old token

	return func() tea.Msg {
		resp, err := m.client.Login(username, password, "incidencias-tui")
		if err != nil {
			return LoginErrorMsg(fmt.Sprintf("Error en %s/api/v1/login: %v", m.client.GetBaseURL(), err))
		}

		// Store token and URL
		m.cfg.Token = resp.Token
		m.cfg.APIURL = m.client.GetBaseURL()
		if err := m.cfg.Save(); err != nil {
			return LoginErrorMsg(fmt.Sprintf("Error guardando configuración: %v", err))
		}

		return LoginSuccessMsg{
			Client: m.client,
			Cfg:    m.cfg,
		}
	}
}
