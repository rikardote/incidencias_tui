package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

// Messages
type EmployeeReportLoadedMsg []models.EmployeeReport
type EmployeeAttendanceLoadedMsg *models.AttendanceResponse
type EmployeeDetailErrorMsg struct {
	Field   string
	Message string
}

// Tab types
type employeeDetailTab int

const (
	tabOverview employeeDetailTab = iota
	tabIncidencias
	tabAsistencia
)

// EmployeeDetailModel shows complete employee information
type EmployeeDetailModel struct {
	client     *api.Client
	employee   models.Employee
	report     []models.EmployeeReport
	attendance *models.AttendanceResponse
	activeTab  employeeDetailTab
	loading    bool
	errorMsg   string
	cursor     int
	offset     int
	pageSize   int
}

// NewEmployeeDetailModel creates a new employee detail view
func NewEmployeeDetailModel(client *api.Client, emp models.Employee) EmployeeDetailModel {
	return EmployeeDetailModel{
		client:   client,
		employee: emp,
		activeTab: tabOverview,
		pageSize: 12,
	}
}

func (m EmployeeDetailModel) Init() tea.Cmd {
	return m.loadAllData()
}

func (m EmployeeDetailModel) loadAllData() tea.Cmd {
	return tea.Batch(m.loadReport(), m.loadAttendance())
}

func (m EmployeeDetailModel) loadReport() tea.Cmd {
	return func() tea.Msg {
		// Last 6 months
		end := time.Now()
		start := end.AddDate(0, -6, 0)
		report, err := m.client.GetEmployeeReport(m.employee.ID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			return EmployeeDetailErrorMsg{Field: "incidencias", Message: err.Error()}
		}
		return EmployeeReportLoadedMsg(report)
	}
}

func (m EmployeeDetailModel) loadAttendance() tea.Cmd {
	return func() tea.Msg {
		// Last 15 days
		end := time.Now()
		start := end.AddDate(0, 0, -15)
		attendance, err := m.client.GetEmployeeAttendance(m.employee.ID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		if err != nil {
			return EmployeeDetailErrorMsg{Field: "asistencia", Message: err.Error()}
		}
		return EmployeeAttendanceLoadedMsg(attendance)
	}
}

func (m EmployeeDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuEmployeeSearch) }
		case tea.KeyTab:
			m.activeTab = (m.activeTab + 1) % 3
			m.cursor = 0
			m.offset = 0
			return m, nil
		case tea.KeyShiftTab:
			if m.activeTab == 0 {
				m.activeTab = 2
			} else {
				m.activeTab--
			}
			m.cursor = 0
			m.offset = 0
			return m, nil
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
			m.cursor = m.maxCursor()
			m.ensureCursorVisible()
		default:
			// Handle character keys
			switch msg.String() {
			case "1":
				m.activeTab = tabOverview
				m.cursor = 0
				m.offset = 0
			case "2":
				m.activeTab = tabIncidencias
				m.cursor = 0
				m.offset = 0
			case "3":
				m.activeTab = tabAsistencia
				m.cursor = 0
				m.offset = 0
			case "r":
				m.loading = true
				m.errorMsg = ""
				return m, m.loadAllData()
			}
		}

	case EmployeeReportLoadedMsg:
		m.report = msg
		m.loading = false

	case EmployeeAttendanceLoadedMsg:
		m.attendance = msg
		m.loading = false

	case EmployeeDetailErrorMsg:
		m.errorMsg = fmt.Sprintf("Error cargando %s: %s", msg.Field, msg.Message)
		m.loading = false
	}

	return m, nil
}

func (m *EmployeeDetailModel) moveCursor(delta int) {
	max := m.maxCursor()
	if max < 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor > max {
		m.cursor = max
	}
	m.ensureCursorVisible()
}

func (m *EmployeeDetailModel) maxCursor() int {
	switch m.activeTab {
	case tabIncidencias:
		return len(m.report) - 1
	case tabAsistencia:
		if m.attendance != nil {
			return len(m.attendance.Data) - 1
		}
	}
	return 0
}

func (m *EmployeeDetailModel) ensureCursorVisible() {
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

func (m EmployeeDetailModel) View() string {
	var s string

	s += "\n"
	s += styles.Breadcrumb([]string{"Menú", "Empleado", m.employee.FullName})
	s += "\n\n"

	// Employee header
	s += m.renderEmployeeHeader()
	s += "\n"

	// Tabs
	s += m.renderTabs()
	s += "\n\n"

	if m.loading {
		s += "  " + styles.InfoText.Render("● Cargando datos...") + "\n"
		return s
	}

	if m.errorMsg != "" {
		s += "  " + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n\n"
	}

	// Tab content
	switch m.activeTab {
	case tabOverview:
		s += m.renderOverview()
	case tabIncidencias:
		s += m.renderIncidencias()
	case tabAsistencia:
		s += m.renderAsistencia()
	}

	s += "\n"
	s += styles.Muted.Render("  Tab/1-2-3: cambiar pestaña · R: recargar · Esc: volver")

	return s
}

func (m EmployeeDetailModel) renderEmployeeHeader() string {
	deptName := "—"
	if m.employee.Department != nil {
		deptName = m.employee.Department.Description
	}

	var lines []string
	lines = append(lines, styles.Title.Render("👤 "+m.employee.FullName))
	lines = append(lines, "")
	lines = append(lines, styles.Label.Render("No. Empleado:")+" "+styles.InfoText.Render(m.employee.NumEmpleado))
	lines = append(lines, styles.Label.Render("Departamento:")+" "+styles.InfoText.Render(deptName))
	lines = append(lines, styles.Label.Render("Puesto:")+" "+styles.InfoText.Render(m.employee.Puesto))
	lines = append(lines, styles.Label.Render("Horario:")+" "+styles.InfoText.Render(valueOrDash(m.employee.Horario)))
	lines = append(lines, styles.Label.Render("Jornada:")+" "+styles.InfoText.Render(valueOrDash(m.employee.Jornada)))

	return styles.Panel.Render(strings.Join(lines, "\n"))
}

func (m EmployeeDetailModel) renderTabs() string {
	var tabs []string

	tabNames := []string{"📋 Resumen", "📊 Incidencias", "🕐 Asistencia"}
	for i, name := range tabNames {
		if employeeDetailTab(i) == m.activeTab {
			tabs = append(tabs, styles.MenuCardActive.Render(fmt.Sprintf(" %s ", name)))
		} else {
			tabs = append(tabs, styles.MenuCard.Render(fmt.Sprintf(" %s ", name)))
		}
	}

	return strings.Join(tabs, " ")
}

func (m EmployeeDetailModel) renderOverview() string {
	var s string

	// Summary stats
	totalIncidencias := len(m.report)
	totalDias := 0.0
	for _, r := range m.report {
		totalDias += r.TotalDias
	}

	s += styles.Panel.Render(
		styles.Subtitle.Render("📈 Resumen últimos 6 meses") + "\n\n" +
			styles.Label.Render("Total incidencias:")+" "+styles.InfoText.Render(fmt.Sprintf("%d", totalIncidencias)) + "\n" +
			styles.Label.Render("Total días:")+" "+styles.InfoText.Render(formatDias(totalDias))+"\n",
	)

	if len(m.report) > 0 {
		s += "\n"
		s += styles.Muted.Render("  Últimas incidencias:")
		s += "\n\n"

		headers := []string{"Código", "Descripción", "Inicio", "Final", "Días", "QNA"}
		colWidths := []int{6, 30, 10, 10, 5, 8}

		tbl := NewTable(headers)
		tbl.PageSz = 5

		count := 0
		for _, r := range m.report {
			if count >= 5 {
				break
			}
			tbl.Rows = append(tbl.Rows, []string{
				r.Codigo,
				truncate(r.Description, 28),
				r.FechaInicio,
				r.FechaFinal,
				formatDias(r.TotalDias),
				r.Qna,
			})
			count++
		}

		s += tbl.Render(colWidths)
	}

	return s
}

func (m EmployeeDetailModel) renderIncidencias() string {
	if len(m.report) == 0 {
		return styles.Muted.Render("  No hay incidencias en los últimos 6 meses")
	}

	headers := []string{"Código", "Descripción", "Inicio", "Final", "Días", "QNA", "Periodo"}
	colWidths := []int{6, 25, 10, 10, 5, 8, 15}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, r := range m.report {
		tbl.Rows = append(tbl.Rows, []string{
			r.Codigo,
			truncate(r.Description, 23),
			r.FechaInicio,
			r.FechaFinal,
			formatDias(r.TotalDias),
			r.Qna,
			valueOrDash(r.Periodo),
		})
	}

	s := tbl.Render(colWidths)
	s += "\n"

	end := m.offset + m.pageSize
	if end > len(m.report) {
		end = len(m.report)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d", m.offset+1, end, len(m.report))

	return s
}

func (m EmployeeDetailModel) renderAsistencia() string {
	if m.attendance == nil || len(m.attendance.Data) == 0 {
		return styles.Muted.Render("  No hay registros de asistencia")
	}

	headers := []string{"Fecha", "Entrada", "Salida", "Checadas", "Retardo", "Incidencias"}
	colWidths := []int{10, 8, 8, 8, 7, 20}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, day := range m.attendance.Data {
		retardo := "No"
		if day.Retardo {
			retardo = styles.ErrorTxt.Render("Sí")
		}

		incidencias := "—"
		if len(day.Incidencias) > 0 {
			incidencias = strings.Join(day.Incidencias, ", ")
		}

		tbl.Rows = append(tbl.Rows, []string{
			day.Date,
			valueOrDash(day.PrimeraChecada),
			valueOrDash(day.UltimaChecada),
			fmt.Sprintf("%d", day.NumChecadas),
			retardo,
			truncate(incidencias, 18),
		})
	}

	s := tbl.Render(colWidths)
	s += "\n"

	end := m.offset + m.pageSize
	if end > len(m.attendance.Data) {
		end = len(m.attendance.Data)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d", m.offset+1, end, len(m.attendance.Data))

	return s
}
