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

// Report messages
type RecentIncidenciasMsg []models.IncidenceRecord
type ReportErrorMsg string

// ReportModel shows recent incidencias
type ReportModel struct {
	client   *api.Client
	results  []models.IncidenceRecord
	loading  bool
	errorMsg string
	loaded   bool
}

// NewReportModel creates a recent incidencias view
func NewReportModel(client *api.Client) ReportModel {
	return ReportModel{
		client: client,
	}
}

// Init implements tea.Model
func (m ReportModel) Init() tea.Cmd {
	return m.doLoad
}

// Update implements tea.Model
func (m ReportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuRecentIncidencias) }
		default:
			if msg.String() == "r" && !m.loading {
				m.loaded = false
				m.loading = true
				return m, m.doLoad
			}
		}

	case RecentIncidenciasMsg:
		m.loading = false
		m.loaded = true
		m.results = msg

	case ReportErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
	}

	return m, nil
}

// View implements tea.Model
func (m ReportModel) View() string {
	var s string

	s += styles.TitleStyle.Render("📋 Incidencias Recientes")
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
		s += "\n" + styles.InfoStyle.Render("No hay incidencias recientes")
		return styles.DocStyle.Render(s)
	}

	s += fmt.Sprintf("\n%s\n\n", styles.SubtitleStyle.Render(fmt.Sprintf("Mostrando %d registros:", len(m.results))))

	// Table header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.TableHeaderStyle.Width(18).Render("Empleado"),
		styles.TableHeaderStyle.Width(30).Render("Nombre"),
		styles.TableHeaderStyle.Width(10).Render("Código"),
		styles.TableHeaderStyle.Width(12).Render("Inicio"),
		styles.TableHeaderStyle.Width(12).Render("Final"),
		styles.TableHeaderStyle.Width(6).Render("Días"),
		styles.TableHeaderStyle.Width(14).Render("QNA"),
	)
	s += header + "\n"
	s += styles.TableHeaderStyle.Width(102).Render(strings.Repeat("─", 100)) + "\n"

	for i, r := range m.results {
		style := styles.TableRowStyle
		if i%2 == 0 {
			style = styles.TableRowAltStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			style.Width(18).Render(r.Employee.NumEmpleado),
			style.Width(30).Render(truncate(r.Employee.FullName, 28)),
			style.Width(10).Render(r.Codigo.Code),
			style.Width(12).Render(r.FechaInicio),
			style.Width(12).Render(r.FechaFinal),
			style.Width(6).Render(fmt.Sprintf("%d", r.TotalDias)),
			style.Width(14).Render(r.Qna),
		)
		s += row + "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("R: recargar · Esc: volver al menú")

	return styles.DocStyle.Render(s)
}

func (m *ReportModel) doLoad() tea.Msg {
	results, err := m.client.GetRecentIncidencias(30)
	if err != nil {
		return ReportErrorMsg(fmt.Sprintf("Error: %v", err))
	}
	return RecentIncidenciasMsg(results)
}
