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

// Code messages
type CodeResultsMsg []models.IncidenceCode
type CodeSelectedMsg models.IncidenceCode
type CodeErrorMsg string

// CodeModel for selecting an incidence code
type CodeModel struct {
	client     *api.Client
	search     textinput.Model
	results    []models.IncidenceCode
	cursor     int
	loading    bool
	errorMsg   string
	searchDone bool
}

// NewCodeModel creates a code selection view
func NewCodeModel(client *api.Client) CodeModel {
	si := textinput.New()
	si.Placeholder = "Código o descripción..."
	si.Prompt = "🔍 "
	si.Focus()
	si.TextStyle = styles.InputFocusedStyle
	si.CharLimit = 128
	si.Width = 60

	return CodeModel{
		client: client,
		search: si,
	}
}

// Init implements tea.Model
func (m CodeModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
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
				sel := m.results[m.cursor]
				return m, func() tea.Msg { return CodeSelectedMsg(sel) }
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
				m.searchDone = false
				m.results = nil
				m.cursor = 0
				m.search.Focus()
				m.search.SetValue("")
			}
		}

	case CodeResultsMsg:
		m.loading = false
		m.results = msg
		m.searchDone = true
		m.cursor = 0
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

// View implements tea.Model
func (m CodeModel) View() string {
	var s string

	s += styles.TitleStyle.Render("📋 Seleccionar Código de Incidencia")
	s += "\n"

	if !m.searchDone {
		s += "\n"
		s += m.search.View()
		s += "\n\n"
		if m.errorMsg != "" {
			s += styles.ErrorStyle.Render("✗ " + m.errorMsg)
			s += "\n\n"
		}
		s += styles.HelpStyle.Render("Enter: buscar · Esc: volver")
		return styles.DocStyle.Render(s)
	}

	s += fmt.Sprintf("\n%s\n\n", styles.InfoStyle.Render("Resultados:"))

	if m.errorMsg != "" {
		s += styles.ErrorStyle.Render("✗ " + m.errorMsg) + "\n\n"
		s += styles.HelpStyle.Render("Backspace: nueva búsqueda")
		return styles.DocStyle.Render(s)
	}

	// Table header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.TableHeaderStyle.Width(8).Render("Código"),
		styles.TableHeaderStyle.Width(50).Render("Descripción"),
		styles.TableHeaderStyle.Width(15).Render("Requiere"),
	)
	s += header + "\n"
	s += styles.TableHeaderStyle.Width(75).Render(strings.Repeat("─", 73)) + "\n"

	for i, code := range m.results {
		style := styles.TableRowStyle
		if i == m.cursor {
			style = styles.MenuItemSelectedStyle.Copy().Width(75)
		} else if i%2 == 0 {
			style = styles.TableRowAltStyle
		}

		req := requirementsSummary(code)
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			style.Width(8).Render(code.Code),
			style.Width(50).Render(truncate(code.Description, 48)),
			style.Width(15).Render(req),
		)
		s += row + "\n"
	}

	s += "\n"
	s += styles.HelpStyle.Render("↑/↓: navegar · Enter: seleccionar · Backspace: buscar · Esc: volver")

	return styles.DocStyle.Render(s)
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
		parts = append(parts, "Incapacidad")
	}
	if c.IsVacacional {
		parts = append(parts, "Vacacional")
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ",")
}
