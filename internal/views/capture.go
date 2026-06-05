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

// Capture messages
type CaptureSuccessMsg struct {
	Token      string
	EmployeeID int
	Message    string
}
type CaptureErrorMsg string

// CaptureModel for the incidence capture form
type CaptureModel struct {
	client        *api.Client
	employee      *models.Employee
	code          *models.IncidenceCode
	user          *models.User

	// Form fields
	fechaInicio   textinput.Model
	fechaFinal    textinput.Model
	medico        textinput.Model
	fechaExpedida textinput.Model
	diagnostico   textinput.Model
	numLicencia   textinput.Model
	periodo       textinput.Model
	autorizaTxt   textinput.Model
	coberturaTxt  textinput.Model
	motivoComision textinput.Model
	otorgado      textinput.Model

	focusIndex  int
	fieldLabels []string
	fieldActive []bool

	loading    bool
	errorMsg   string
	successMsg string
	Done       bool
}

// fieldAt returns a pointer to the textinput at the given index.
// This ensures we always get a pointer to the current model's field.
func (m *CaptureModel) fieldAt(idx int) *textinput.Model {
	switch idx {
	case 0:
		return &m.fechaInicio
	case 1:
		return &m.fechaFinal
	case 2:
		return &m.medico
	case 3:
		return &m.fechaExpedida
	case 4:
		return &m.diagnostico
	case 5:
		return &m.numLicencia
	case 6:
		return &m.periodo
	case 7:
		return &m.autorizaTxt
	case 8:
		return &m.coberturaTxt
	case 9:
		return &m.motivoComision
	case 10:
		return &m.otorgado
	default:
		return nil
	}
}

// totalFields returns the total number of fields
func (m *CaptureModel) totalFields() int {
	return 11
}

// NewCaptureModel creates a capture form for a specific employee + code
func NewCaptureModel(client *api.Client, emp *models.Employee, code *models.IncidenceCode, user *models.User) CaptureModel {
	m := CaptureModel{
		client:   client,
		employee: emp,
		code:     code,
		user:     user,
	}

	// Initialize fields
	fi := textinput.New()
	fi.Placeholder = "YYYY-MM-DD"
	fi.CharLimit = 10
	fi.Width = 18
	fi.Prompt = ""

	ff := textinput.New()
	ff.Placeholder = "YYYY-MM-DD"
	ff.CharLimit = 10
	ff.Width = 18
	ff.Prompt = ""

	med := textinput.New()
	med.Placeholder = "Nombre o ID del médico"
	med.CharLimit = 128
	med.Width = 40
	med.Prompt = ""

	fexp := textinput.New()
	fexp.Placeholder = "YYYY-MM-DD"
	fexp.CharLimit = 10
	fexp.Width = 18
	fexp.Prompt = ""

	diag := textinput.New()
	diag.Placeholder = "Diagnóstico"
	diag.CharLimit = 255
	diag.Width = 40
	diag.Prompt = ""

	nlic := textinput.New()
	nlic.Placeholder = "Número de licencia"
	nlic.CharLimit = 50
	nlic.Width = 30
	nlic.Prompt = ""

	per := textinput.New()
	per.Placeholder = "ID del periodo"
	per.CharLimit = 10
	per.Width = 18
	per.Prompt = ""

	atxt := textinput.New()
	atxt.Placeholder = "Autoriza TXT"
	atxt.CharLimit = 100
	atxt.Width = 30
	atxt.Prompt = ""

	ctxt := textinput.New()
	ctxt.Placeholder = "Cobertura TXT"
	ctxt.CharLimit = 100
	ctxt.Width = 30
	ctxt.Prompt = ""

	mcom := textinput.New()
	mcom.Placeholder = "Motivo de comisión"
	mcom.CharLimit = 255
	mcom.Width = 40
	mcom.Prompt = ""

	oto := textinput.New()
	oto.Placeholder = "Otorgado por"
	oto.CharLimit = 100
	oto.Width = 30
	oto.Prompt = ""

	m.fechaInicio = fi
	m.fechaFinal = ff
	m.medico = med
	m.fechaExpedida = fexp
	m.diagnostico = diag
	m.numLicencia = nlic
	m.periodo = per
	m.autorizaTxt = atxt
	m.coberturaTxt = ctxt
	m.motivoComision = mcom
	m.otorgado = oto

	// Build labels and active flags
	m.buildFields()

	return m
}

// buildFields sets up labels and active flags based on the code's requirements
func (m *CaptureModel) buildFields() {
	m.fieldLabels = []string{
		"Fecha Inicio",
		"Fecha Final",
		"Médico",
		"Fecha Expedida",
		"Diagnóstico",
		"Núm. Licencia",
		"Periodo ID",
		"Autoriza TXT",
		"Cobertura TXT",
		"Motivo Comisión",
		"Otorgado",
	}

	m.fieldActive = []bool{
		true, // fecha_inicio always active
		m.code.RequiresRange,
		m.code.RequiresMedico || m.code.IsIncapacidad,
		m.code.IsIncapacidad,
		m.code.IsIncapacidad || m.code.IsLicencia,
		m.code.IsLicencia || m.code.IsIncapacidad,
		m.code.RequiresPeriodo || m.code.IsVacacional,
		m.code.RequiresTxt,
		m.code.RequiresTxt,
		m.code.RequiresComision,
		m.code.RequiresOtorgado,
	}

	// Focus the first active field
	for i, active := range m.fieldActive {
		if active {
			m.focusIndex = i
			f := m.fieldAt(i)
			f.Focus()
			f.TextStyle = styles.InputFocusedStyle
			break
		}
	}
}

// IsDone returns true if the capture has been completed
func (m CaptureModel) IsDone() bool {
	return m.Done
}

// Init implements tea.Model
func (m CaptureModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m CaptureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading || m.Done {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m, func() tea.Msg { return MenuSelectedMsg(MenuCaptureIncidence) }

		case tea.KeyTab, tea.KeyDown:
			m.nextField()
			return m, nil

		case tea.KeyShiftTab, tea.KeyUp:
			m.prevField()
			return m, nil

		case tea.KeyEnter:
			// Submit the form
			return m, m.doCapture()
		}

	case CaptureSuccessMsg:
		m.loading = false
		m.successMsg = msg.Message
		m.Done = true
		return m, nil

	case CaptureErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update the focused field directly by calling Update on the value
	// We use fieldAt() to get a pointer to the CURRENT model's field
	if m.focusIndex >= 0 && m.focusIndex < m.totalFields() && m.fieldActive[m.focusIndex] {
		f := m.fieldAt(m.focusIndex)
		if f != nil {
			var cmd tea.Cmd
			*f, cmd = f.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m CaptureModel) View() string {
	var s string

	s += styles.TitleStyle.Render("📝 Capturar Incidencia")
	s += "\n\n"

	// Employee summary
	s += styles.BoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styles.InfoStyle.Render(fmt.Sprintf("Empleado: %s - %s", m.employee.NumEmpleado, m.employee.FullName)),
			styles.SubtitleStyle.Render(fmt.Sprintf("Código: %s - %s", m.code.Code, m.code.Description)),
			styles.SubtitleStyle.Render(fmt.Sprintf("Departamento: %s", m.employee.Department.Description)),
		),
	)
	s += "\n\n"

	if m.Done {
		s += styles.SuccessStyle.Render("✓ " + m.successMsg)
		s += "\n\n"
		s += styles.HelpStyle.Render("Presiona Enter para volver al menú")
		return styles.DocStyle.Render(s)
	}

	// Render active fields
	for i := 0; i < m.totalFields(); i++ {
		if !m.fieldActive[i] {
			continue
		}

		f := m.fieldAt(i)
		if f == nil {
			continue
		}

		label := styles.LabelStyle.Render(m.fieldLabels[i] + ":")
		field := f.View()

		if i == m.focusIndex {
			s += lipgloss.JoinHorizontal(lipgloss.Top,
				label,
				field,
			) + "\n"
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Top,
				label,
				field,
			) + "\n"
		}
	}

	s += "\n"

	if m.errorMsg != "" {
		s += styles.ErrorStyle.Render("✗ " + m.errorMsg)
		s += "\n"
	}

	if m.loading {
		s += styles.InfoStyle.Render("Capturando incidencia...")
	} else {
		s += styles.HelpStyle.Render("Enter: capturar · Tab/↓: siguiente · ↑: anterior · Esc: cancelar")
	}

	return styles.DocStyle.Render(s)
}

func (m *CaptureModel) nextField() {
	// Blur current
	cur := m.fieldAt(m.focusIndex)
	if cur != nil {
		cur.Blur()
		cur.TextStyle = styles.InputStyle
	}

	// Find next active field
	total := m.totalFields()
	for i := 1; i <= total; i++ {
		idx := (m.focusIndex + i) % total
		if m.fieldActive[idx] {
			m.focusIndex = idx
			break
		}
	}

	// Focus new
	next := m.fieldAt(m.focusIndex)
	if next != nil {
		next.Focus()
		next.TextStyle = styles.InputFocusedStyle
	}
}

func (m *CaptureModel) prevField() {
	// Blur current
	cur := m.fieldAt(m.focusIndex)
	if cur != nil {
		cur.Blur()
		cur.TextStyle = styles.InputStyle
	}

	// Find previous active field
	total := m.totalFields()
	for i := 1; i <= total; i++ {
		idx := (m.focusIndex - i + total) % total
		if m.fieldActive[idx] {
			m.focusIndex = idx
			break
		}
	}

	// Focus new
	next := m.fieldAt(m.focusIndex)
	if next != nil {
		next.Focus()
		next.TextStyle = styles.InputFocusedStyle
	}
}

func (m *CaptureModel) doCapture() tea.Cmd {
	// Validate required fields
	m.errorMsg = ""

	if strings.TrimSpace(m.fechaInicio.Value()) == "" {
		m.errorMsg = "Fecha de inicio es requerida"
		return nil
	}

	m.loading = true

	// Build the request
	req := models.StoreIncidenciaRequest{
		EmployeeID:  m.employee.ID,
		Codigo:      m.code.ID,
		FechaInicio: m.fechaInicio.Value(),
	}

	// Set fecha_final (use fecha_inicio if not provided)
	if m.fieldActive[1] {
		req.FechaFinal = m.fechaFinal.Value()
	}
	if req.FechaFinal == "" {
		req.FechaFinal = req.FechaInicio
	}

	// Optional fields based on active flags
	if m.fieldActive[3] { // fecha_expedida
		req.FechaExpedida = m.fechaExpedida.Value()
	}
	if m.fieldActive[4] { // diagnostico
		req.Diagnostico = m.diagnostico.Value()
	}
	if m.fieldActive[5] { // num_licencia
		req.NumLicencia = m.numLicencia.Value()
	}
	if m.fieldActive[7] { // autoriza_txt
		req.AutorizaTxt = m.autorizaTxt.Value()
	}
	if m.fieldActive[8] { // cobertura_txt
		req.CoberturaTxt = m.coberturaTxt.Value()
	}
	if m.fieldActive[9] { // motivo_comision
		req.MotivoComision = m.motivoComision.Value()
	}
	if m.fieldActive[10] { // otorgado
		req.Otorgado = m.otorgado.Value()
	}

	// Admin can skip validations
	if m.user.IsAdmin {
		req.SaltarValidacionInca = true
		req.SaltarValidacionLic = true
	}

	return func() tea.Msg {
		resp, err := m.client.StoreIncidencia(req)
		if err != nil {
			return CaptureErrorMsg(fmt.Sprintf("%v", err))
		}
		return CaptureSuccessMsg{
			Token:      resp.Token,
			EmployeeID: resp.EmployeeID,
			Message:    resp.Message,
		}
	}
}
