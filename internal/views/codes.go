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

type CodeResultsMsg []models.IncidenceCode
type CodeSelectedMsg models.IncidenceCode
type CodeErrorMsg string

type CodeModel struct {
	client     *api.Client
	title      string
	context    string
	search     textinput.Model
	results    []models.IncidenceCode
	cursor     int
	offset     int
	pageSize   int
	loading    bool
	errorMsg   string
	searchDone bool
}

func NewCodeModel(client *api.Client) CodeModel {
	return NewCodeModelFor(client, "Seleccionar Código de Incidencia", "")
}

func NewCodeModelFor(client *api.Client, title, context string) CodeModel {
	si := textinput.New()
	si.Placeholder = "Código o descripción..."
	si.Prompt = ""
	si.Focus()
	si.TextStyle = styles.InputFocused
	si.CharLimit = 128
	si.Width = 50

	return CodeModel{
		client:   client,
		title:    title,
		context:  context,
		search:   si,
		pageSize: 15,
	}
}

func (m CodeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CodeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuCaptureIncidence) }

		case tea.KeyEnter:
			if !m.searchDone {
				query := m.search.Value()
				m.loading = true
				m.errorMsg = ""
				return m, m.doSearch(query)
			}
			if len(m.results) > 0 {
				return m, func() tea.Msg { return CodeSelectedMsg(m.results[m.cursor]) }
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
				m.offset = 0
				m.search.Focus()
				m.search.SetValue("")
			}
		}

	case CodeResultsMsg:
		m.loading = false
		m.results = msg
		m.searchDone = true
		m.cursor = 0
		m.offset = 0
		if len(msg) == 0 {
			m.errorMsg = "No se encontraron códigos"
		}

	case CodeErrorMsg:
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

func (m CodeModel) View() string {
	var s string

	s += "\n"
	s += styles.Breadcrumb([]string{"Menú", "Capturar", m.title})
	s += "\n"

	if strings.Contains(m.context, "Paso") {
		s += styles.Stepper([]string{"Empleado", "Código", "Datos"}, 1)
		s += "\n"
	}

	s += "\n"
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

	s += fmt.Sprintf("  Resultados: %d códigos\n\n", len(m.results))

	if m.errorMsg != "" {
		s += "  " + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n\n"
		return s
	}

	headers := []string{"Código", "Descripción", "Requiere"}
	colWidths := []int{8, 50, 18}

	tbl := NewTable(headers)
	tbl.Cursor = m.cursor
	tbl.Offset = m.offset
	tbl.PageSz = m.pageSize

	for _, code := range m.results {
		req := requirementsSummary(code)
		tbl.Rows = append(tbl.Rows, []string{code.Code, code.Description, req})
	}

	s += tbl.Render(colWidths)
	s += "\n"

	end := m.offset + m.pageSize
	if end > len(m.results) {
		end = len(m.results)
	}
	s += fmt.Sprintf("  Mostrando %d-%d de %d", m.offset+1, end, len(m.results))

	return s
}

func (m *CodeModel) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.GetIncidenceCodes(query)
		if err != nil {
			return CodeErrorMsg(fmt.Sprintf("Error: %v", err))
		}
		return CodeResultsMsg(results)
	}
}

func requirementsSummary(c models.IncidenceCode) string {
	var parts []string
	if c.RequiresRange {
		parts = append(parts, "Rango")
	}
	if c.RequiresMedico {
		parts = append(parts, "Médico")
	}
	if c.RequiresPeriodo {
		parts = append(parts, "Periodo")
	}
	if c.RequiresTxt {
		parts = append(parts, "TXT")
	}
	if c.IsIncapacidad {
		parts = append(parts, "Incap.")
	}
	if c.IsVacacional {
		parts = append(parts, "Vac.")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func (m *CodeModel) ensureCursorVisible() {
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
