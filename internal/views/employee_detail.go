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
type EmployeeVacationsLoadedMsg *models.VacationResponse
type EmployeeDetailErrorMsg struct {
	Field   string
	Message string
}

// Tab types
type employeeDetailTab int

const (
	tabIncidencias employeeDetailTab = iota
	tabAsistencia
	tabVacaciones
)

// EmployeeDetailModel shows complete employee information
type EmployeeDetailModel struct {
	client     *api.Client
	employee   models.Employee
	report     []models.EmployeeReport
	attendance *models.AttendanceResponse
	vacations  *models.VacationResponse
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
		activeTab: tabIncidencias,
		pageSize: 12,
	}
}

func (m EmployeeDetailModel) Init() tea.Cmd {
	return m.loadAllData()
}

func (m EmployeeDetailModel) loadAllData() tea.Cmd {
	return tea.Batch(m.loadReport(), m.loadAttendance(), m.loadVacations())
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

func (m EmployeeDetailModel) loadVacations() tea.Cmd {
	return func() tea.Msg {
		vacations, err := m.client.GetEmployeeVacaciones(m.employee.ID)
		if err != nil {
			return EmployeeDetailErrorMsg{Field: "vacaciones", Message: err.Error()}
		}
		return EmployeeVacationsLoadedMsg(vacations)
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
				m.activeTab = tabIncidencias
				m.cursor = 0
				m.offset = 0
			case "2":
				m.activeTab = tabAsistencia
				m.cursor = 0
				m.offset = 0
			case "3":
				m.activeTab = tabVacaciones
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

	case EmployeeVacationsLoadedMsg:
		m.vacations = msg
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
	case tabVacaciones:
		if m.vacations != nil {
			return len(m.vacations.Periods) - 1
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
	case tabIncidencias:
		s += m.renderIncidencias()
	case tabAsistencia:
		s += m.renderAsistencia()
	case tabVacaciones:
		s += m.renderVacaciones()
	}

	s += "\n"
	s += styles.Muted.Render("  Tab/1-3: cambiar pestaña · R: recargar · Esc: volver")

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

	tabNames := []string{"📊 Incidencias", "🕐 Asistencia", "🌴 Vacaciones"}
	for i, name := range tabNames {
		if employeeDetailTab(i) == m.activeTab {
			tabs = append(tabs, styles.MenuCardActive.Render(fmt.Sprintf(" %s ", name)))
		} else {
			tabs = append(tabs, styles.MenuCard.Render(fmt.Sprintf(" %s ", name)))
		}
	}

	return strings.Join(tabs, " ")
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
		codigo := ""
		descripcion := ""
		if r.Codigo != nil {
			codigo = r.Codigo.Code
			descripcion = r.Codigo.Description
		}
		qna := ""
		if r.Qna != nil {
			qna = fmt.Sprintf("%s/%d", r.Qna.Qna, r.Qna.Year)
		}
		periodo := ""
		if r.Periodo != nil {
			periodo = fmt.Sprintf("%02d/%d", r.Periodo.Periodo, r.Periodo.Year)
		}
		tbl.Rows = append(tbl.Rows, []string{
			codigo,
			truncate(descripcion, 23),
			formatDateDMY(r.FechaInicio),
			formatDateDMY(r.FechaFinal),
			formatDias(r.TotalDias),
			qna,
			valueOrDash(periodo),
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

	headers := []string{"Fecha", "1ra Entrada", "Última Salida", "Checadas", "Retardo", "Incidencias"}
	colWidths := []int{10, 11, 13, 8, 7, 20}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for i := len(m.attendance.Data) - 1; i >= 0; i-- {
		day := m.attendance.Data[i]
		retardo := "No"
		if day.Retardo {
			retardo = styles.ErrorTxt.Render("Sí")
		}

		incidencias := "—"
		if len(day.Incidencias) > 0 {
			incidencias = strings.Join(day.Incidencias, ", ")
		}

		tbl.Rows = append(tbl.Rows, []string{
			formatDateDMY(day.Date),
			extractTime(day.PrimeraChecada),
			extractTime(day.UltimaChecada),
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

func (m EmployeeDetailModel) renderVacaciones() string {
	if m.vacations == nil || len(m.vacations.Periods) == 0 {
		return styles.Muted.Render("  No hay información de vacaciones")
	}

	// Summary panel
	summary := []string{
		styles.Subtitle.Render("🌴 Resumen de Vacaciones"),
		"",
		styles.Label.Render("Derecho por periodo:") + " " + styles.InfoText.Render(formatDias(m.vacations.Entitlement) + " días"),
		styles.Label.Render("Total pendiente:") + " " + styles.InfoText.Render(formatDias(m.vacations.TotalPending) + " días"),
	}

	s := styles.Panel.Render(strings.Join(summary, "\n"))
	s += "\n\n"

	// Periods table
	headers := []string{"Periodo", "Derecho", "Usados", "Pendientes"}
	colWidths := []int{12, 10, 10, 12}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, period := range m.vacations.Periods {
		tbl.Rows = append(tbl.Rows, []string{
			period.Period.Label,
			formatDias(period.Entitlement),
			formatDias(period.Used),
			formatDias(period.Pending),
		})
	}

	s += tbl.Render(colWidths)
	s += "\n"

	end := m.offset + m.pageSize
	if end > len(m.vacations.Periods) {
		end = len(m.vacations.Periods)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d periodos", m.offset+1, end, len(m.vacations.Periods))

	return s
}

// extractTime extracts HH:MM from a datetime string like "2024-01-15 14:30:00"
// Returns "—" if empty
func extractTime(s string) string {
	if s == "" {
		return "—"
	}
	// If it contains a space, it's a full datetime - extract the time part
	if strings.Contains(s, " ") {
		parts := strings.Split(s, " ")
		if len(parts) >= 2 {
			timePart := parts[1]
			// Extract HH:MM from HH:MM:SS
			if len(timePart) >= 5 {
				return timePart[:5]
			}
			return timePart
		}
	}
	// If it's already just a time like "14:30", return as-is
	return s
}

// formatDateDMY formats a date from YYYY-MM-DD to DD-MM-YYYY for display
// Returns the original string if it doesn't match the expected format
func formatDateDMY(s string) string {
	if s == "" {
		return "—"
	}
	// Try to parse as YYYY-MM-DD
	parts := strings.Split(s, "-")
	if len(parts) == 3 && len(parts[0]) == 4 {
		// YYYY-MM-DD format, convert to DD-MM-YYYY
		return parts[2] + "-" + parts[1] + "-" + parts[0]
	}
	// Return as-is if not in expected format
	return s
}
