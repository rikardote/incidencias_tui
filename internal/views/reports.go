package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

type RecentIncidenciasMsg []models.IncidenceRecord
type ReportErrorMsg string

type ReportModel struct {
	client   *api.Client
	results  []models.IncidenceRecord
	cursor   int
	offset   int
	pageSize int
	loading  bool
	errorMsg string
	loaded   bool
}

func NewReportModel(client *api.Client) ReportModel {
	return ReportModel{
		client:   client,
		pageSize: 15,
	}
}

func (m ReportModel) Init() tea.Cmd {
	return m.doLoad
}

func (m ReportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuRecentIncidencias) }
		case tea.KeyUp:
			m.moveCursor(-1)
		case tea.KeyDown:
			m.moveCursor(1)
		case tea.KeyPgUp:
			m.moveCursor(-m.pageSize)
		case tea.KeyPgDown:
			m.moveCursor(m.pageSize)
		case tea.KeyHome:
			m.cursor = 0
			m.offset = 0
		case tea.KeyEnd:
			if len(m.results) > 0 {
				m.cursor = len(m.results) - 1
				m.ensureCursorVisible()
			}
		default:
			if msg.String() == "r" && !m.loading {
				m.loaded = false
				m.loading = true
				m.errorMsg = ""
				return m, m.doLoad
			}
		}

	case RecentIncidenciasMsg:
		m.loading = false
		m.loaded = true
		m.results = msg
		m.cursor = 0
		m.offset = 0

	case ReportErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
	}

	return m, nil
}

func (m ReportModel) View() string {
	var s string

	s += "\n"
	s += styles.Breadcrumb([]string{"Menú", "Incidencias Recientes"})
	s += "\n\n"
	s += styles.ScreenTitle("Incidencias Recientes", "Últimos registros capturados")
	s += "\n\n"

	if m.loading {
		s += "  " + styles.InfoText.Render("● Cargando incidencias...") + "\n"
		return s
	}

	if m.errorMsg != "" {
		s += "  " + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n"
		return s
	}

	if !m.loaded {
		s += "  " + styles.InfoText.Render("● Preparando consulta...") + "\n"
		return s
	}

	if len(m.results) == 0 {
		s += "  " + styles.Muted.Render("No hay incidencias recientes") + "\n"
		return s
	}

	s += fmt.Sprintf("  %s\n\n", styles.Badge.Render(fmt.Sprintf(" %d registros · seleccionado: %d ", len(m.results), m.cursor+1)))

	headers := []string{"Fecha", "Empleado", "Nombre", "Código", "Inicio", "Final", "Días", "QNA"}
	colWidths := []int{11, 8, 25, 6, 10, 10, 5, 8}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, r := range m.results {
		empNum := ""
		empName := ""
		codeStr := ""
		if r.Employee != nil {
			empNum = r.Employee.NumEmpleado
			empName = r.Employee.FullName
		}
		if r.Codigo != nil {
			codeStr = r.Codigo.Code
		}
		tbl.Rows = append(tbl.Rows, []string{
			formatDateDMY(r.FechaCapturado),
			empNum,
			empName,
			codeStr,
			formatDateDMY(r.FechaInicio),
			formatDateDMY(r.FechaFinal),
			formatDias(r.TotalDias),
			valueOrDash(r.Qna),
		})
	}

	s += tbl.Render(colWidths)
	s += "\n"

	if len(m.results) > 0 && m.cursor < len(m.results) {
		r := m.results[m.cursor]
		emp := "Sin empleado"
		if r.Employee != nil {
			emp = fmt.Sprintf("%s - %s", r.Employee.NumEmpleado, r.Employee.FullName)
		}
		code := "Sin código"
		if r.Codigo != nil {
			code = fmt.Sprintf("%s - %s", r.Codigo.Code, r.Codigo.Description)
		}

		s += "\n" + styles.Panel.Render(
			styles.Subtitle.Render("📋 Detalle seleccionado") + "\n\n" +
				styles.Label.Render("Empleado:") + " " + styles.InfoText.Render(emp) + "\n" +
				styles.Label.Render("Código:") + " " + styles.InfoText.Render(code) + "\n" +
				styles.Label.Render("Periodo:") + " " + styles.InfoText.Render(formatDateDMY(r.FechaInicio)+" a "+formatDateDMY(r.FechaFinal)+" · "+formatDias(r.TotalDias)+" días"),
		)
	}

	s += "\n"
	end := m.offset + m.pageSize
	if end > len(m.results) {
		end = len(m.results)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d", m.offset+1, end, len(m.results))

	return s
}

func (m *ReportModel) doLoad() tea.Msg {
	results, err := m.client.GetRecentIncidencias(100)
	if err != nil {
		return ReportErrorMsg(fmt.Sprintf("Error: %v", err))
	}
	return RecentIncidenciasMsg(results)
}

func (m *ReportModel) moveCursor(delta int) {
	if len(m.results) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	m.ensureCursorVisible()
}

func (m *ReportModel) ensureCursorVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.pageSize {
		m.offset = m.cursor - m.pageSize + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func valueOrDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func formatDias(d float64) string {
	if d == float64(int(d)) {
		return fmt.Sprintf("%d", int(d))
	}
	return fmt.Sprintf("%.1f", d)
}
