package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

// Employee messages
type EmployeeResultsMsg []models.Employee
type EmployeeSelectedMsg models.Employee
type EmployeeErrorMsg string

// EmployeeModel handles employee search and selection
type EmployeeModel struct {
	client     *api.Client
	search     textinput.Model
	results    []models.Employee
	selected   int
	cursor     int
	loading    bool
	errorMsg   string
	searchDone bool
	pageSize   int
	page       int
}

// NewEmployeeModel creates an employee search view
func NewEmployeeModel(client *api.Client) EmployeeModel {
	si := textinput.New()
	si.Placeholder = "Nombre, apellido o número de empleado..."
	si.Prompt = "🔍 "
	si.Focus()
	si.TextStyle = styles.InputFocusedStyle
	si.CharLimit = 128
	si.Width = 60

	return EmployeeModel{
		client:   client,
		search:   si,
		pageSize: 15,
		page:     0,
	}
}

// Init implements tea.Model
func (m EmployeeModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m EmployeeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuEmployeeSearch) } // go back to menu

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

			// If search is done and we have results, select the current item
			if len(m.results) > 0 {
				selected := m.results[m.cursor]
				return m, func() tea.Msg {
					return EmployeeSelectedMsg(selected)
				}
			}

		case tea.KeyUp:
			if m.searchDone && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.searchDone && m.cursor < len(m.results)-1 {
				m.cursor++
			}

		case tea.KeyBackspace:
			if m.searchDone {
				// Go back to search
				m.searchDone = false
				m.results = nil
				m.cursor = 0
				m.search.Focus()
				m.search.SetValue("")
			}

		case tea.KeyRunes, tea.KeySpace:
			if !m.searchDone {
				// Let the input handle it
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

	// Update search input if not done
	if !m.searchDone {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m EmployeeModel) View() string {
	var s string

	s += styles.TitleStyle.Render("🔍 Búsqueda de Empleados")
	s += "\n"

	if !m.searchDone {
		s += "\n"
		s += m.search.View()
		s += "\n\n"
		if m.errorMsg != "" {
			s += styles.ErrorStyle.Render("✗ " + m.errorMsg)
			s += "\n\n"
		}
		s += styles.HelpStyle.Render("Enter: buscar · Esc: volver al menú")
		return styles.DocStyle.Render(s)
	}

	// Results display
	s += fmt.Sprintf("\n%s %s\n\n",
		styles.SubtitleStyle.Render("Resultados para:"),
		styles.InfoStyle.Render(m.search.Value()),
	)

	if m.errorMsg != "" {
		s += styles.ErrorStyle.Render("✗ " + m.errorMsg) + "\n\n"
		s += styles.HelpStyle.Render("Presiona Backspace para buscar de nuevo")
		return styles.DocStyle.Render(s)
	}

	// Render results as a table
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.TableHeaderStyle.Width(12).Render("Empleado"),
		styles.TableHeaderStyle.Width(40).Render("Nombre"),
		styles.TableHeaderStyle.Width(30).Render("Departamento"),
		styles.TableHeaderStyle.Width(20).Render("Puesto"),
	)
	s += header + "\n"
	s += styles.TableHeaderStyle.Width(102).Render(strings.Repeat("─", 100)) + "\n"

	for i, emp := range m.results {
		style := styles.TableRowStyle
		if i == m.cursor {
			style = styles.MenuItemSelectedStyle.Copy().Width(102)
		} else if i%2 == 0 {
			style = styles.TableRowAltStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			style.Width(12).Render(emp.NumEmpleado),
			style.Width(40).Render(truncate(emp.FullName, 38)),
			style.Width(30).Render(truncate(emp.Department.Description, 28)),
			style.Width(20).Render(truncate(emp.Puesto, 18)),
		)
		s += row + "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("↑/↓: navegar · Enter: seleccionar · Backspace: nueva búsqueda · Esc: menú")

	return styles.DocStyle.Render(s)
}

func (m *EmployeeModel) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.SearchEmployees(query)
		if err != nil {
			return EmployeeErrorMsg(fmt.Sprintf("Error de búsqueda: %v", err))
		}
		return EmployeeResultsMsg(results)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
