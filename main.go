package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
	"incidencias_tui/internal/views"
)

type screen int

const (
	screenLogin screen = iota
	screenMenu
	screenEmployees
	screenCodes
	screenCapture
	screenReports
	screenBiometric
	screenEmployeeDetail
)

type employeePurpose int

const (
	employeePurposeSearch employeePurpose = iota
	employeePurposeCapture
)

type mainModel struct {
	current         screen
	employeePurpose employeePurpose

	client *api.Client
	cfg    *config.Config
	user   *models.User

	login          views.LoginModel
	menu           views.MenuModel
	employees      views.EmployeeModel
	codes          views.CodeModel
	capture        views.CaptureModel
	reports        views.ReportModel
	biometric      views.BiometricModel
	employeeDetail views.EmployeeDetailModel

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

	return mainModel{
		current: screenLogin,
		cfg:     cfg,
		client:  api.New(cfg.APIURL),
		login:   views.NewLoginModel(cfg),
	}
}

func (m mainModel) Init() tea.Cmd {
	if m.cfg.HasToken() {
		m.client.SetToken(m.cfg.Token)
		return func() tea.Msg {
			user, err := m.client.Me()
			if err != nil {
				m.cfg.ClearToken()
				return nil
			}
			return loginSuccessMsg{client: m.client, cfg: m.cfg, user: user}
		}
	}
	return m.login.Init()
}

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

	case views.LoginSuccessMsg:
		m.client = msg.Client
		m.cfg = msg.Cfg
		return m, func() tea.Msg {
			user, err := m.client.Me()
			if err != nil {
				return views.LoginErrorMsg(fmt.Sprintf("Error: %v", err))
			}
			return loginSuccessMsg{client: m.client, cfg: m.cfg, user: user}
		}

	case loginSuccessMsg:
		m.client = msg.client
		m.cfg = msg.cfg
		m.user = msg.user
		m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
		m.current = screenMenu
		return m, nil

	case views.MenuSelectedMsg:
		switch msg {
		case views.MenuEmployeeSearch:
			m.employeePurpose = employeePurposeSearch
			m.employees = views.NewEmployeeModel(m.client)
			m.current = screenEmployees
			return m, m.employees.Init()

		case views.MenuCaptureIncidence:
			m.employeePurpose = employeePurposeCapture
			m.captureEmployee = nil
			m.captureCode = nil
			m.employees = views.NewEmployeeModelFor(m.client, "Seleccionar Empleado", "Paso 1 de 3")
			m.current = screenEmployees
			return m, m.employees.Init()

		case views.MenuRecentIncidencias:
			m.reports = views.NewReportModel(m.client)
			m.current = screenReports
			return m, m.reports.Init()

		case views.MenuBiometric:
			m.biometric = views.NewBiometricModel(m.client)
			m.current = screenBiometric
			return m, m.biometric.Init()

		case views.MenuLogout:
			m.client.Logout()
			m.cfg.ClearToken()
			m.user = nil
			m.client.SetToken("")
			m.login = views.NewLoginModel(m.cfg)
			m.current = screenLogin
			return m, m.login.Init()

		case views.MenuQuit:
			return m, tea.Quit
		}

	case views.EmployeeSelectedMsg:
		emp := models.Employee(msg)
		if m.employeePurpose == employeePurposeCapture {
			m.captureEmployee = &emp
			m.codes = views.NewCodeModelFor(m.client, "Seleccionar Código",
				fmt.Sprintf("Paso 2 de 3 · %s", emp.FullName))
			m.current = screenCodes
			return m, m.codes.Init()
		}
		// Search mode: show employee detail
		m.employeeDetail = views.NewEmployeeDetailModel(m.client, emp)
		m.current = screenEmployeeDetail
		return m, m.employeeDetail.Init()

	case views.CodeSelectedMsg:
		code := models.IncidenceCode(msg)
		if m.captureEmployee != nil {
			m.captureCode = &code
			m.capture = views.NewCaptureModel(m.client, m.captureEmployee, m.captureCode, m.user)
			m.current = screenCapture
			return m, m.capture.Init()
		}
		return m, nil
	}

	switch m.current {
	case screenLogin:
		model, cmd := m.login.Update(msg)
		m.login = model.(views.LoginModel)
		return m, cmd

	case screenMenu:
		model, cmd := m.menu.Update(msg)
		m.menu = model.(views.MenuModel)
		return m, cmd

	case screenEmployees:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
			m.current = screenMenu
			return m, nil
		}
		model, cmd := m.employees.Update(msg)
		m.employees = model.(views.EmployeeModel)
		return m, cmd

	case screenCodes:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.employees = views.NewEmployeeModelFor(m.client, "Seleccionar Empleado", "Paso 1 de 3")
			m.current = screenEmployees
			return m, m.employees.Init()
		}
		model, cmd := m.codes.Update(msg)
		m.codes = model.(views.CodeModel)
		return m, cmd

	case screenCapture:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc && !m.capture.IsSelectingPicker() {
			m.codes = views.NewCodeModelFor(m.client, "Seleccionar Código", "Paso 2 de 3")
			m.current = screenCodes
			return m, m.codes.Init()
		}
		model, cmd := m.capture.Update(msg)
		m.capture = model.(views.CaptureModel)

		if m.capture.IsDone() {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				if keyMsg.Type == tea.KeyEnter {
					m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
					m.current = screenMenu
					return m, nil
				}
				if keyMsg.String() == "r" {
					m.reports = views.NewReportModel(m.client)
					m.current = screenReports
					return m, m.reports.Init()
				}
			}
		}
		return m, cmd

	case screenReports:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
			m.current = screenMenu
			return m, nil
		}
		model, cmd := m.reports.Update(msg)
		m.reports = model.(views.ReportModel)
		return m, cmd

	case screenBiometric:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.menu = views.NewMenuModel(m.client, m.cfg, m.user)
			m.current = screenMenu
			return m, nil
		}
		model, cmd := m.biometric.Update(msg)
		m.biometric = model.(views.BiometricModel)
		return m, cmd

	case screenEmployeeDetail:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.employees = views.NewEmployeeModel(m.client)
			m.current = screenEmployees
			return m, m.employees.Init()
		}
		model, cmd := m.employeeDetail.Update(msg)
		m.employeeDetail = model.(views.EmployeeDetailModel)
		return m, cmd
	}

	return m, nil
}

func (m mainModel) View() string {
	var content string
	var footer string

	switch m.current {
	case screenLogin:
		content = m.login.View()
		footer = "Tab: siguiente · Enter: entrar · Ctrl+L: API · Ctrl+C: salir"
	case screenMenu:
		content = m.menu.View()
		footer = "↑↓←→: navegar · Enter: seleccionar · Ctrl+C: salir"
	case screenEmployees:
		content = m.employees.View()
		footer = "Enter: buscar · ↑↓: navegar · Esc: menú"
	case screenCodes:
		content = m.codes.View()
		footer = "Enter: buscar · ↑↓: navegar · Esc: atrás"
	case screenCapture:
		content = m.capture.View()
		footer = "Tab/Enter: siguiente · Ctrl+S: capturar · Esc: atrás"
	case screenReports:
		content = m.reports.View()
		footer = "↑↓: navegar · R: recargar · Esc: menú"
	case screenBiometric:
		content = m.biometric.View()
		footer = "↑↓: navegar · R: recargar · Esc: menú"
	case screenEmployeeDetail:
		content = m.employeeDetail.View()
		footer = "Tab: pestañas · ↑↓: navegar · R: recargar · Esc: volver"
	default:
		content = "Saliendo..."
		footer = ""
	}

	// Header
	userInfo := ""
	if m.user != nil {
		userInfo = fmt.Sprintf("  %s (%s)", m.user.Name, m.user.Type)
	}
	headerLeft := styles.HeaderBar.Render("⚡ Incidencias TUI")
	headerRight := styles.HeaderBar.Copy().Align(lipgloss.Right).Render(userInfo)
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerLeft, headerRight)

	// Footer
	ft := styles.FooterBar.Render(footer)

	// Compose
	return lipgloss.JoinVertical(lipgloss.Left, header, content, ft)
}

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

func padRight(s string, width int) string {
	if lipgloss.Width(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}
