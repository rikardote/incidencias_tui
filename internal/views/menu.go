package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/config"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

type MenuSelectedMsg int

const (
	MenuEmployeeSearch MenuSelectedMsg = iota + 1
	MenuCaptureIncidence
	MenuRecentIncidencias
	MenuBiometric
	MenuLogout
	MenuQuit
)

type MenuModel struct {
	client  *api.Client
	cfg     *config.Config
	user    *models.User
	choices []menuItem
	cursor  int
}

type menuItem struct {
	icon  string
	title string
	desc  string
	msg   MenuSelectedMsg
}

func NewMenuModel(client *api.Client, cfg *config.Config, user *models.User) MenuModel {
	items := []menuItem{
		{icon: "👤", title: "Buscar empleados", desc: "Consulta por número o nombre", msg: MenuEmployeeSearch},
	}
	if user != nil && user.CanCapture {
		items = append(items, menuItem{icon: "📝", title: "Capturar incidencia", desc: "Registra una nueva incidencia", msg: MenuCaptureIncidence})
	}
	items = append(items,
		menuItem{icon: "📊", title: "Incidencias recientes", desc: "Últimos registros capturados", msg: MenuRecentIncidencias},
		menuItem{icon: "🕐", title: "Biométrico", desc: "Checadas del reloj biométrico", msg: MenuBiometric},
		menuItem{icon: "🚪", title: "Cerrar sesión", desc: "Volver al inicio de sesión", msg: MenuLogout},
		menuItem{icon: "⬅️ ", title: "Salir", desc: "Cerrar la aplicación", msg: MenuQuit},
	)

	return MenuModel{
		client:  client,
		cfg:     cfg,
		user:    user,
		choices: items,
		cursor:  0,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor >= 2 {
				m.cursor -= 2
			}

		case tea.KeyDown:
			if m.cursor+2 < len(m.choices) {
				m.cursor += 2
			}

		case tea.KeyLeft:
			if m.cursor%2 > 0 {
				m.cursor--
			}

		case tea.KeyRight:
			if m.cursor%2 < 1 && m.cursor+1 < len(m.choices) {
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

func (m MenuModel) View() string {
	var s string

	s += "\n"
	s += styles.Title.Render("Menú principal")
	s += "\n"
	s += styles.Muted.Render(fmt.Sprintf("Usuario: %s (%s)", m.user.Name, m.user.Type))
	if m.user.CanCapture {
		s += styles.SuccessTxt.Render(" · captura habilitada")
	} else {
		s += styles.Muted.Render(" · solo consulta")
	}
	s += "\n\n"

	// Render cards in 2-column grid
	for i := 0; i < len(m.choices); i += 2 {
		var row []string
		for j := 0; j < 2 && i+j < len(m.choices); j++ {
			idx := i + j
			item := m.choices[idx]
			card := m.renderCard(item, idx == m.cursor)
			row = append(row, card)
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, row...)
		s += "\n"
	}

	return s
}

func (m MenuModel) renderCard(item menuItem, selected bool) string {
	var cardStyle lipgloss.Style
	if selected {
		cardStyle = styles.MenuCardActive
	} else {
		cardStyle = styles.MenuCard
	}

	var content string
	content += styles.Title.Render(item.icon + " " + item.title) + "\n"
	content += styles.Muted.Render(item.desc)

	return cardStyle.Render(content)
}

func (m MenuModel) GetUser() *models.User {
	return m.user
}

func (m MenuModel) GetClient() *api.Client {
	return m.client
}

func (m MenuModel) GetConfig() *config.Config {
	return m.cfg
}
