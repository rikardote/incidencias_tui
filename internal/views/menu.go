package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

// Menu selection messages
type MenuSelectedMsg int

const (
	MenuEmployeeSearch MenuSelectedMsg = iota + 1
	MenuCaptureIncidence
	MenuRecentIncidencias
	MenuBiometric
	MenuLogout
	MenuQuit
)

// MenuModel is the main menu screen
type MenuModel struct {
	client  *api.Client
	cfg     *config.Config
	user    *models.User
	choices []menuItem
	cursor  int
}

type menuItem struct {
	title string
	msg   MenuSelectedMsg
}

// NewMenuModel creates the main menu
func NewMenuModel(client *api.Client, cfg *config.Config, user *models.User) MenuModel {
	items := []menuItem{
		{title: "🔍 Buscar Empleados", msg: MenuEmployeeSearch},
		{title: "📝 Capturar Incidencia", msg: MenuCaptureIncidence},
		{title: "📋 Incidencias Recientes", msg: MenuRecentIncidencias},
		{title: "👤 Biométrico", msg: MenuBiometric},
		{title: "🚪 Cerrar Sesión", msg: MenuLogout},
		{title: "❌ Salir", msg: MenuQuit},
	}

	return MenuModel{
		client:  client,
		cfg:     cfg,
		user:    user,
		choices: items,
		cursor:  0,
	}
}

// Init implements tea.Model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case tea.KeyEnter:
			selected := m.choices[m.cursor]
			return m, func() tea.Msg {
				return selected.msg
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m MenuModel) View() string {
	var s string

	// User info header
	userInfo := fmt.Sprintf("👤 %s (%s)", m.user.Name, m.user.Type)
	if m.user.CanCapture {
		userInfo += " 📷 Captura habilitada"
	}
	s += styles.TitleStyle.Render("🏠 Menú Principal")
	s += "\n"
	s += styles.InfoStyle.Render(userInfo)
	s += "\n\n"

	// Menu items
	for i, item := range m.choices {
		if i == m.cursor {
			s += styles.MenuItemSelectedStyle.Render("▸ " + item.title)
		} else {
			s += styles.MenuItemStyle.Render("  " + item.title)
		}
		s += "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("↑/↓: navegar · Enter: seleccionar · Esc/Ctrl+C: salir")

	return styles.DocStyle.Render(s)
}

// GetUser returns the current user
func (m MenuModel) GetUser() *models.User {
	return m.user
}

// GetClient returns the API client
func (m MenuModel) GetClient() *api.Client {
	return m.client
}

// GetConfig returns the config
func (m MenuModel) GetConfig() *config.Config {
	return m.cfg
}


