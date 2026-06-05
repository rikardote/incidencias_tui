package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

// Biometric messages
type BiometricResultsMsg []models.BiometricRecord
type BiometricErrorMsg string

// BiometricModel shows recent biometric records
type BiometricModel struct {
	client   *api.Client
	results  []models.BiometricRecord
	loading  bool
	errorMsg string
	loaded   bool
}

// NewBiometricModel creates a biometric view
func NewBiometricModel(client *api.Client) BiometricModel {
	return BiometricModel{
		client: client,
	}
}

// Init implements tea.Model
func (m BiometricModel) Init() tea.Cmd {
	return m.doLoad
}

// Update implements tea.Model
func (m BiometricModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuBiometric) }
		default:
			if msg.String() == "r" && !m.loading {
				m.loaded = false
				m.loading = true
				return m, m.doLoad
			}
		}

	case BiometricResultsMsg:
		m.loading = false
		m.loaded = true
		m.results = msg

	case BiometricErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
	}

	return m, nil
}

// View implements tea.Model
func (m BiometricModel) View() string {
	var s string

	s += styles.TitleStyle.Render("👤 Registros Biométricos Recientes")
	s += "\n"

	if m.loading {
		s += "\n" + styles.InfoStyle.Render("Cargando...")
		return styles.DocStyle.Render(s)
	}

	if m.errorMsg != "" {
		s += "\n" + styles.ErrorStyle.Render("✗ "+m.errorMsg)
		return styles.DocStyle.Render(s)
	}

	if !m.loaded {
		s += "\n" + styles.InfoStyle.Render("Preparando...")
		return styles.DocStyle.Render(s)
	}

	if len(m.results) == 0 {
		s += "\n" + styles.InfoStyle.Render("No hay registros biométricos recientes")
		return styles.DocStyle.Render(s)
	}

	s += fmt.Sprintf("\n%s\n\n", styles.SubtitleStyle.Render(fmt.Sprintf("Mostrando %d registros:", len(m.results))))

	// Table header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.TableHeaderStyle.Width(18).Render("Empleado"),
		styles.TableHeaderStyle.Width(30).Render("Nombre"),
		styles.TableHeaderStyle.Width(20).Render("Fecha/Hora"),
		styles.TableHeaderStyle.Width(20).Render("Ubicación"),
	)
	s += header + "\n"
	s += styles.TableHeaderStyle.Width(90).Render(strings.Repeat("─", 88)) + "\n"

	for i, r := range m.results {
		style := styles.TableRowStyle
		if i%2 == 0 {
			style = styles.TableRowAltStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			style.Width(18).Render(r.NumEmpleado),
			style.Width(30).Render(truncate(r.Employee.FullName, 28)),
			style.Width(20).Render(r.Fecha),
			style.Width(20).Render(truncate(r.Location, 18)),
		)
		s += row + "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("R: recargar · Esc: volver al menú")

	return styles.DocStyle.Render(s)
}

func (m *BiometricModel) doLoad() tea.Msg {
	results, err := m.client.GetRecentBiometric(30)
	if err != nil {
		return BiometricErrorMsg(fmt.Sprintf("Error: %v", err))
	}
	return BiometricResultsMsg(results)
}
