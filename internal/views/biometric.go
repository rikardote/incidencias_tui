package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

type BiometricResultsMsg []models.BiometricRecord
type BiometricErrorMsg string

type BiometricModel struct {
	client   *api.Client
	results  []models.BiometricRecord
	cursor   int
	offset   int
	pageSize int
	loading  bool
	errorMsg string
	loaded   bool
}

func NewBiometricModel(client *api.Client) BiometricModel {
	return BiometricModel{
		client:   client,
		pageSize: 15,
	}
}

func (m BiometricModel) Init() tea.Cmd {
	return m.doLoad
}

func (m BiometricModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuBiometric) }
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

	case BiometricResultsMsg:
		m.loading = false
		m.loaded = true
		m.results = msg
		m.cursor = 0
		m.offset = 0

	case BiometricErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
	}

	return m, nil
}

func (m BiometricModel) View() string {
	var s string

	s += "\n"
	s += styles.Breadcrumb([]string{"Menú", "Biométrico"})
	s += "\n\n"
	s += styles.ScreenTitle("Registros Biométricos", "Checadas recientes del reloj biométrico")
	s += "\n\n"

	if m.loading {
		s += "  " + styles.InfoText.Render("● Cargando registros...") + "\n"
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
		s += "  " + styles.Muted.Render("No hay registros biométricos") + "\n"
		return s
	}

	s += fmt.Sprintf("  %s\n\n", styles.Badge.Render(fmt.Sprintf(" %d registros · seleccionado: %d ", len(m.results), m.cursor+1)))

	headers := []string{"Fecha", "Hora", "Empleado", "Nombre", "Ubicación"}
	colWidths := []int{11, 9, 8, 30, 20}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, r := range m.results {
		empName := ""
		if r.Employee != nil {
			empName = r.Employee.FullName
		}
		tbl.Rows = append(tbl.Rows, []string{
			formatDateDMY(r.Fecha),
			r.Hora,
			r.NumEmpleado,
			empName,
			r.Location,
		})
	}

	s += tbl.Render(colWidths)
	s += "\n"

	if len(m.results) > 0 && m.cursor < len(m.results) {
		r := m.results[m.cursor]
		empName := "—"
		if r.Employee != nil {
			empName = r.Employee.FullName
		}

		s += "\n" + styles.Panel.Render(
			styles.Subtitle.Render("🕐 Detalle seleccionado") + "\n\n" +
				styles.Label.Render("Empleado:") + " " + styles.InfoText.Render(fmt.Sprintf("%s - %s", r.NumEmpleado, empName)) + "\n" +
				styles.Label.Render("Fecha:") + " " + styles.InfoText.Render(formatDateDMY(r.Fecha)) + "\n" +
				styles.Label.Render("Hora:") + " " + styles.InfoText.Render(r.Hora) + "\n" +
				styles.Label.Render("Ubicación:") + " " + styles.InfoText.Render(valueOrDash(r.Location)),
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

func (m *BiometricModel) doLoad() tea.Msg {
	results, err := m.client.GetRecentBiometric(100)
	if err != nil {
		return BiometricErrorMsg(fmt.Sprintf("Error: %v", err))
	}
	return BiometricResultsMsg(results)
}

func (m *BiometricModel) moveCursor(delta int) {
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

func (m *BiometricModel) ensureCursorVisible() {
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
