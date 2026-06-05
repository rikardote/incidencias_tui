package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

type EmployeeResultsMsg []models.Employee
type EmployeeSelectedMsg models.Employee
type EmployeeErrorMsg string

type EmployeeModel struct {
	client     *api.Client
	title      string
	context    string
	search     textinput.Model
	results    []models.Employee
	cursor     int
	loading    bool
	errorMsg   string
	searchDone bool
	pageSize   int
	page       int
}

func NewEmployeeModel(client *api.Client) EmployeeModel {
	return NewEmployeeModelFor(client, "Búsqueda de Empleados", "")
}

func NewEmployeeModelFor(client *api.Client, title, context string) EmployeeModel {
	si := textinput.New()
	si.Placeholder = "Nombre, apellido o número..."
	si.Prompt = ""
	si.Focus()
	si.TextStyle = styles.InputFocused
	si.CharLimit = 128
	si.Width = 50

	return EmployeeModel{
		client:   client,
		title:    title,
		context:  context,
		search:   si,
		pageSize: 15,
	}
}

func (m EmployeeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EmployeeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuEmployeeSearch) }

		case tea.KeyEnter:
			if !m.searchDone {
				query := m.search.Value()
				if strings.TrimSpace(query) == "" {
					m.errorMsg = "Ingresa un término de búsqueda"
					return m, nil
				}
				m.loading = true
				m.errorMsg = ""
				return m, m.doSearch(query)
			}
			if len(m.results) > 0 {
				return m, func() tea.Msg {
					return EmployeeSelectedMsg(m.results[m.cursor])
				}
			}

		case tea.KeyUp:
			if m.searchDone && m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}

		case tea.KeyDown:
			if m.searchDone && m.cursor < len(m.results)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}

		case tea.KeyPgUp:
			if m.searchDone {
				m.cursor -= m.pageSize
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.ensureCursorVisible()
			}

		case tea.KeyPgDown:
			if m.searchDone && len(m.results) > 0 {
				m.cursor += m.pageSize
				if m.cursor >= len(m.results) {
					m.cursor = len(m.results) - 1
				}
				m.ensureCursorVisible()
			}

		case tea.KeyBackspace:
			if m.searchDone {
				m.searchDone = false
				m.results = nil
				m.cursor = 0
				m.page = 0
				m.search.Focus()
				m.search.SetValue("")
			}
		}

	case EmployeeResultsMsg:
		m.loading = false
		m.results = msg
		m.searchDone = true
		m.cursor = 0
		m.errorMsg = ""
		if len(msg) == 0 {
			m.errorMsg = "No se encontraron empleados"
		}

	case EmployeeErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
	}

	if !m.searchDone {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m EmployeeModel) View() string {
	var s string

	s += "\n"
	s += styles.Breadcrumb([]string{"Menú", m.title})
	s += "\n\n"
	s += styles.ScreenTitle(m.title, m.context)
	s += "\n\n"

	if !m.searchDone {
		s += "  🔍 " + m.search.View()
		s += "\n\n"
		if m.errorMsg != "" {
			s += "  " + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n\n"
		}
		if m.loading {
			s += "  " + styles.InfoText.Render("● Buscando...") + "\n\n"
		}
		return s
	}

	s += fmt.Sprintf("  Resultados: %s · %d encontrados\n\n",
		styles.InfoText.Render(m.search.Value()),
		len(m.results),
	)

	if m.errorMsg != "" {
		s += "  " + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n\n"
		return s
	}

	// Table
	headers := []string{"Empleado", "Nombre", "Departamento", "Puesto"}
	colWidths := []int{10, 28, 22, 16}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.page
	tbl.PageSz = m.pageSize

	for _, emp := range m.results {
		deptName := ""
		if emp.Department != nil {
			deptName = emp.Department.Description
		}
		tbl.Rows = append(tbl.Rows, []string{emp.NumEmpleado, emp.FullName, deptName, emp.Puesto})
	}

	s += tbl.Render(colWidths)
	s += "\n"

	end := m.page + m.pageSize
	if end > len(m.results) {
		end = len(m.results)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d", m.page+1, end, len(m.results))

	return s
}

func (m *EmployeeModel) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.SearchEmployees(query)
		if err != nil {
			return EmployeeErrorMsg(fmt.Sprintf("Error: %v", err))
		}
		return EmployeeResultsMsg(results)
	}
}

func (m *EmployeeModel) ensureCursorVisible() {
	if m.cursor < m.page {
		m.page = m.cursor
	}
	if m.cursor >= m.page+m.pageSize {
		m.page = m.cursor - m.pageSize + 1
	}
	if m.page < 0 {
		m.page = 0
	}
}
