package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
	"incidencias_tui/internal/views"
)

// Screen identifiers
type screen int

const (
	screenLogin screen = iota
	screenMenu
	screenEmployees
	screenCodes
	screenCapture
	screenReports
	screenBiometric
	screenQuit
)

// mainModel is the root model that routes between screens
type mainModel struct {
	current  screen
	previous screen

	// Shared state
	client *api.Client
	cfg    *config.Config
	user   *models.User

	// Sub-models (only one active at a time)
	login     views.LoginModel
	menu      views.MenuModel
	employees views.EmployeeModel
	codes     views.CodeModel
	capture   views.CaptureModel
	reports   views.ReportModel
	biometric views.BiometricModel

	// Capture flow state
	captureEmployee *models.Employee
	captureCode     *models.IncidenceCode

	width  int
	height int
}

func newMainModel() mainModel {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = &config.Config{APIURL: config.DefaultAPI}
	}

	loginModel := views.NewLoginModel(cfg)

	m := mainModel{
		current: screenLogin,
		cfg:     cfg,
		client:  api.New(cfg.APIURL),
		login:   loginModel,
	}
	return m
}

// Init implements tea.Model
func (m mainModel) Init() tea.Cmd {
	if m.cfg.HasToken() {
		m.client.SetToken(m.cfg.Token)
		return func() tea.Msg {
			user, err := m.client.Me()
			if err != nil {
				m.cfg.ClearToken()
				return nil
			}
			return loginSuccessMsg{
				client: m.client,
				cfg:    m.cfg,
				user:   user,
			}
		}
	}
	return m.login.Init()
}

// Update implements tea.Model
func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	// --- Login messages ---
	case views.LoginSuccessMsg:
		m.client = msg.Client
		m.cfg = msg.Cfg
		return m, func() tea.Msg {
			user, err := m.client.Me()
			if err != nil {
				return views.LoginErrorMsg(fmt.Sprintf("Error obteniendo usuario: %v", err))
			}
			return loginSuccessMsg{
				client: m.client,
				cfg:    m.cfg,
				user:   user,
			}
		}

	case loginSuccessMsg:
		m.client = msg.client
		m.cfg = msg.cfg
		m.user = msg.user
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		m.previous = screenLogin
		return m, nil

	// --- Menu messages ---
	case views.MenuSelectedMsg:
		switch msg {
		case views.MenuEmployeeSearch:
			m.employees = views.NewEmployeeModel(m.client)
			m.current = screenEmployees
			m.previous = screenMenu
			return m, m.employees.Init()

		case views.MenuCaptureIncidence:
			m.employees = views.NewEmployeeModel(m.client)
			m.current = screenEmployees
			m.previous = screenCodes
			return m, m.employees.Init()

		case views.MenuRecentIncidencias:
			m.reports = views.NewReportModel(m.client)
			m.current = screenReports
			m.previous = screenMenu
			return m, m.reports.Init()

		case views.MenuBiometric:
			m.biometric = views.NewBiometricModel(m.client)
			m.current = screenBiometric
			m.previous = screenMenu
			return m, m.biometric.Init()

		case views.MenuLogout:
			m.client.Logout() //nolint:errcheck
			m.cfg.ClearToken()
			m.user = nil
			m.client.SetToken("")
			m.login = views.NewLoginModel(m.cfg)
			m.current = screenLogin
			return m, m.login.Init()

		case views.MenuQuit:
			return m, tea.Quit
		}

	// --- Employee messages ---
	case views.EmployeeSelectedMsg:
		emp := models.Employee(msg)
		if m.previous == screenCodes {
			m.captureEmployee = &emp
			m.codes = views.NewCodeModel(m.client)
			m.current = screenCodes
			m.previous = screenCapture
			return m, m.codes.Init()
		}
		return m, nil

	// --- Code messages ---
	case views.CodeSelectedMsg:
		code := models.IncidenceCode(msg)
		if m.previous == screenCapture && m.captureEmployee != nil {
			m.captureCode = &code
			m.capture = views.NewCaptureModel(m.client, m.captureEmployee, m.captureCode, m.user)
			m.current = screenCapture
			m.previous = screenCodes
			return m, m.capture.Init()
		}
		return m, nil

	// --- Capture messages ---
	case views.CaptureSuccessMsg:
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		m.captureEmployee = nil
		m.captureCode = nil
		return m, nil
	}

	// Route messages to the active sub-model
	switch m.current {
	case screenLogin:
		return m.forwardToLogin(msg)
	case screenMenu:
		return m.forwardToMenu(msg)
	case screenEmployees:
		return m.forwardToEmployees(msg)
	case screenCodes:
		return m.forwardToCodes(msg)
	case screenCapture:
		return m.forwardToCapture(msg)
	case screenReports:
		return m.forwardToReports(msg)
	case screenBiometric:
		return m.forwardToBiometric(msg)
	}

	return m, nil
}

// View implements tea.Model
func (m mainModel) View() string {
	var content string
	switch m.current {
	case screenLogin:
		content = m.login.View()
	case screenMenu:
		content = m.menu.View()
	case screenEmployees:
		content = m.employees.View()
	case screenCodes:
		content = m.codes.View()
	case screenCapture:
		content = m.capture.View()
	case screenReports:
		content = m.reports.View()
	case screenBiometric:
		content = m.biometric.View()
	default:
		content = "Saliendo..."
	}
	return styles.AppStyle.Render(content)
}

// --- Forwarding methods ---

func (m *mainModel) forwardToLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.login.Update(msg)
	m.login = model.(views.LoginModel)
	return m, cmd
}

func (m *mainModel) forwardToMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.menu.Update(msg)
	m.menu = model.(views.MenuModel)
	return m, cmd
}

func (m *mainModel) forwardToEmployees(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for escape to go back
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		m.captureEmployee = nil
		return m, nil
	}

	model, cmd := m.employees.Update(msg)
	m.employees = model.(views.EmployeeModel)
	return m, cmd
}

func (m *mainModel) forwardToCodes(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for escape to go back
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		m.captureEmployee = nil
		return m, nil
	}

	model, cmd := m.codes.Update(msg)
	m.codes = model.(views.CodeModel)
	return m, cmd
}

func (m *mainModel) forwardToCapture(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for escape to go back
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.codes = views.NewCodeModel(m.client)
		m.current = screenCodes
		m.previous = screenCapture
		return m, m.codes.Init()
	}

	model, cmd := m.capture.Update(msg)
	m.capture = model.(views.CaptureModel)

	// If capture done and user pressed Enter, go to menu
	if m.capture.IsDone() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter {
			m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
			m.current = screenMenu
			m.captureEmployee = nil
			m.captureCode = nil
			return m, nil
		}
	}

	return m, cmd
}

func (m *mainModel) forwardToReports(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		return m, nil
	}

	model, cmd := m.reports.Update(msg)
	m.reports = model.(views.ReportModel)
	return m, cmd
}

func (m *mainModel) forwardToBiometric(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		return m, nil
	}

	model, cmd := m.biometric.Update(msg)
	m.biometric = model.(views.BiometricModel)
	return m, cmd
}

// loginSuccessMsg is an internal message after successful token validation
type loginSuccessMsg struct {
	client *api.Client
	cfg    *config.Config
	user   *models.User
}

func main() {
	p := tea.NewProgram(
		newMainModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
