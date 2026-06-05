package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/api"
	"incidencias_tui/internal/models"
	"incidencias_tui/internal/styles"
)

type CaptureSuccessMsg struct {
	Token      string
	EmployeeID int
	Message    string
}
type CaptureErrorMsg string

type CaptureModel struct {
	client   *api.Client
	employee *models.Employee
	code     *models.IncidenceCode
	user     *models.User

	fechaInicio    textinput.Model
	fechaFinal     textinput.Model
	medico         textinput.Model
	fechaExpedida  textinput.Model
	diagnostico    textinput.Model
	numLicencia    textinput.Model
	periodo        textinput.Model
	autorizaTxt    textinput.Model
	coberturaTxt   textinput.Model
	motivoComision textinput.Model
	otorgado       textinput.Model

	focusIndex  int
	fieldLabels []string
	fieldActive []bool

	periods         []models.Periodo
	selectingPeriod bool
	periodCursor    int
	selectedPeriod  *models.Periodo

	doctors         []models.Doctor
	selectingDoctor bool
	doctorCursor    int
	selectedDoctor  *models.Doctor

	loading    bool
	errorMsg   string
	successMsg string
	Done       bool
}

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

func (m *CaptureModel) totalFields() int {
	return 11
}

func NewCaptureModel(client *api.Client, emp *models.Employee, code *models.IncidenceCode, user *models.User) CaptureModel {
	m := CaptureModel{
		client:   client,
		employee: emp,
		code:     code,
		user:     user,
	}

	m.fechaInicio = m.newInput("YYYYMMDD o YYYY-MM-DD", 10, 24)
	m.fechaFinal = m.newInput("YYYYMMDD o YYYY-MM-DD", 10, 24)
	m.medico = m.newInput("Escribe nombre o #Empleado y presiona Enter", 128, 50)
	m.fechaExpedida = m.newInput("YYYYMMDD o YYYY-MM-DD", 10, 24)
	m.diagnostico = m.newInput("Diagnóstico", 255, 40)
	m.numLicencia = m.newInput("Número de licencia", 50, 30)
	m.periodo = m.newInput("Presiona Enter para seleccionar", 10, 30)
	m.autorizaTxt = m.newInput("Autoriza TXT", 100, 30)
	m.coberturaTxt = m.newInput("Cobertura TXT", 100, 30)
	m.motivoComision = m.newInput("Motivo de comisión", 255, 40)
	m.otorgado = m.newInput("Otorgado por", 100, 30)

	m.buildFields()
	return m
}

func (m *CaptureModel) newInput(placeholder string, charLimit, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.Width = width
	ti.Prompt = ""
	return ti
}

func (m *CaptureModel) buildFields() {
	m.fieldLabels = []string{
		"Fecha Inicio",
		"Fecha Final",
		"Médico",
		"Fecha Expedida",
		"Diagnóstico",
		"Núm. Licencia",
		"Periodo",
		"Autoriza TXT",
		"Cobertura TXT",
		"Motivo Comisión",
		"Otorgado",
	}

	m.fieldActive = []bool{
		true,
		m.requiresDateRange(),
		m.requiresIncapacityDetails(),
		m.requiresIncapacityDetails(),
		m.requiresDiagnosis(),
		m.requiresIncapacityDetails(),
		m.requiresPeriod(),
		m.requiresTXTFields(),
		m.requiresTXTFields(),
		m.requiresCommissionReason(),
		m.requiresGrantedBy(),
	}

	for i, active := range m.fieldActive {
		if active {
			m.focusIndex = i
			f := m.fieldAt(i)
			f.Focus()
			f.TextStyle = styles.InputFocused
			break
		}
	}
}

func (m *CaptureModel) requiresDateRange() bool {
	if m.code.RequiresRange {
		return true
	}
	switch normalizedCode(m.code.Code) {
	case "40", "41", "47", "48", "49", "53", "54", "55", "60", "61", "62", "63":
		return true
	default:
		return false
	}
}

func (m *CaptureModel) requiresIncapacityDetails() bool {
	switch normalizedCode(m.code.Code) {
	case "53", "54", "55":
		return true
	default:
		return false
	}
}

func (m *CaptureModel) requiresDiagnosis() bool {
	return m.requiresIncapacityDetails()
}

func (m *CaptureModel) requiresPeriod() bool {
	if m.code.RequiresPeriodo || m.code.IsVacacional {
		return true
	}
	switch normalizedCode(m.code.Code) {
	case "60", "62", "63":
		return true
	default:
		return false
	}
}

func (m *CaptureModel) requiresTXTFields() bool {
	return m.code.RequiresTxt || normalizedCode(m.code.Code) == "900"
}

func (m *CaptureModel) requiresCommissionReason() bool {
	return m.code.RequiresComision || normalizedCode(m.code.Code) == "61"
}

func (m *CaptureModel) requiresGrantedBy() bool {
	return m.code.RequiresOtorgado || normalizedCode(m.code.Code) == "901"
}

func normalizedCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.TrimLeft(code, "0")
	if code == "" {
		return "0"
	}
	return code
}

func (m CaptureModel) IsDone() bool {
	return m.Done
}

func (m CaptureModel) IsSelectingPicker() bool {
	return m.selectingPeriod || m.selectingDoctor
}

func (m CaptureModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CaptureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.selectingPeriod {
		return m.updatePeriodSelection(msg)
	}
	if m.selectingDoctor {
		return m.updateDoctorSelection(msg)
	}

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
			if m.focusIndex == 2 && m.fieldActive[2] {
				return m, m.loadDoctors()
			}
			if m.focusIndex == 6 && m.fieldActive[6] {
				return m, m.loadPeriods()
			}
			m.nextField()
			return m, nil
		case tea.KeyCtrlS:
			return m, m.doCapture()
		case tea.KeyRunes, tea.KeyBackspace, tea.KeyDelete:
			if m.focusIndex == 2 {
				m.selectedDoctor = nil
			}
			if m.focusIndex == 6 {
				m.selectedPeriod = nil
			}
		}

	case DoctorsLoadedMsg:
		m.loading = false
		m.doctors = msg
		m.selectingDoctor = true
		m.doctorCursor = 0
		if len(msg) == 0 {
			m.errorMsg = "No se encontraron médicos"
			m.selectingDoctor = false
		}
		return m, nil

	case DoctorsErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil

	case PeriodsLoadedMsg:
		m.loading = false
		m.periods = msg
		m.selectingPeriod = true
		m.periodCursor = 0
		if len(msg) == 0 {
			m.errorMsg = "No hay periodos disponibles"
			m.selectingPeriod = false
		}
		return m, nil

	case PeriodsErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil

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

type PeriodsLoadedMsg []models.Periodo
type PeriodsErrorMsg string
type DoctorsLoadedMsg []models.Doctor
type DoctorsErrorMsg string

func (m *CaptureModel) loadDoctors() tea.Cmd {
	query := strings.TrimSpace(m.medico.Value())
	if query == "" {
		m.errorMsg = "Escribe parte del nombre o número del médico"
		return nil
	}
	m.loading = true
	m.errorMsg = ""
	return func() tea.Msg {
		doctors, err := m.client.GetDoctors(query)
		if err != nil {
			return DoctorsErrorMsg(fmt.Sprintf("Error: %v", err))
		}
		return DoctorsLoadedMsg(doctors)
	}
}

func (m *CaptureModel) loadPeriods() tea.Cmd {
	m.loading = true
	m.errorMsg = ""
	return func() tea.Msg {
		periods, err := m.client.GetPeriodos()
		if err != nil {
			return PeriodsErrorMsg(fmt.Sprintf("Error: %v", err))
		}
		return PeriodsLoadedMsg(periods)
	}
}

func (m CaptureModel) updatePeriodSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.selectingPeriod = false
			return m, nil
		case tea.KeyUp:
			if m.periodCursor > 0 {
				m.periodCursor--
			}
			return m, nil
		case tea.KeyDown:
			if m.periodCursor < len(m.periods)-1 {
				m.periodCursor++
			}
			return m, nil
		case tea.KeyEnter:
			if len(m.periods) > 0 {
				m.selectedPeriod = &m.periods[m.periodCursor]
				m.periodo.SetValue(fmt.Sprintf("%d", m.selectedPeriod.ID))
				m.selectingPeriod = false
				m.nextField()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m CaptureModel) updateDoctorSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.selectingDoctor = false
			return m, nil
		case tea.KeyUp:
			if m.doctorCursor > 0 {
				m.doctorCursor--
			}
			return m, nil
		case tea.KeyDown:
			if m.doctorCursor < len(m.doctors)-1 {
				m.doctorCursor++
			}
			return m, nil
		case tea.KeyEnter:
			if len(m.doctors) > 0 {
				m.selectedDoctor = &m.doctors[m.doctorCursor]
				m.medico.SetValue(fmt.Sprintf("%s - %s", m.selectedDoctor.NumEmpleado, m.selectedDoctor.FullName))
				m.selectingDoctor = false
				m.nextField()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m CaptureModel) View() string {
	var s string

	// Breadcrumb
	s += "\n" + styles.Breadcrumb([]string{"Menú", "Capturar", "Datos"}) + "\n"

	// Stepper
	s += styles.Stepper([]string{"Empleado", "Código", "Datos"}, 2) + "\n\n"

	// Title
	s += styles.ScreenTitle("Capturar Incidencia", "Completa los datos y presiona Ctrl+S")
	s += "\n\n"

	if m.Done {
		s += "  " + styles.SuccessTxt.Render("✓ "+m.successMsg) + "\n\n"
		s += "  " + styles.Muted.Render("R: ver incidencias · Enter: menú")
		return s
	}

	if m.selectingDoctor {
		s += m.renderDoctorSelection()
		return s
	}
	if m.selectingPeriod {
		s += m.renderPeriodSelection()
		return s
	}

	// Context box
	deptName := ""
	if m.employee.Department != nil {
		deptName = m.employee.Department.Description
	}

	var ctx strings.Builder
	ctx.WriteString(styles.Label.Render("Empleado:") + " ")
	ctx.WriteString(styles.InfoText.Render(fmt.Sprintf("%s - %s", m.employee.NumEmpleado, m.employee.FullName)))
	ctx.WriteString("\n")
	ctx.WriteString(styles.Label.Render("Departamento:") + " ")
	ctx.WriteString(styles.InfoText.Render(deptName))
	ctx.WriteString("\n")
	ctx.WriteString(styles.Label.Render("Código:") + " ")
	ctx.WriteString(styles.InfoText.Render(fmt.Sprintf("%s - %s", m.code.Code, m.code.Description)))

	s += styles.Panel.Render(ctx.String())
	s += "\n\n"

	// Group fields
	var groups []string

	// Dates
	if m.fieldActive[0] || m.fieldActive[1] {
		var dateFields []string
		dateFields = append(dateFields, styles.Subtitle.Render("📅 Fechas"))
		dateFields = append(dateFields, "")
		if m.fieldActive[0] {
			dateFields = append(dateFields, m.renderField(0))
		}
		if m.fieldActive[1] {
			dateFields = append(dateFields, m.renderField(1))
		}
		groups = append(groups, styles.Panel.Render(strings.Join(dateFields, "\n")))
	}

	// Medical
	if m.fieldActive[2] || m.fieldActive[3] || m.fieldActive[4] || m.fieldActive[5] {
		var medFields []string
		medFields = append(medFields, styles.Subtitle.Render("🏥 Información médica"))
		medFields = append(medFields, "")

		if m.selectedDoctor != nil && m.fieldActive[2] {
			medFields = append(medFields, styles.Label.Render("Médico:")+" "+
				styles.SuccessTxt.Render(fmt.Sprintf("✓ %s - %s", m.selectedDoctor.NumEmpleado, m.selectedDoctor.FullName)))
		} else if m.fieldActive[2] {
			medFields = append(medFields, m.renderField(2))
		}

		if m.fieldActive[3] {
			medFields = append(medFields, m.renderField(3))
		}
		if m.fieldActive[4] {
			medFields = append(medFields, m.renderField(4))
		}
		if m.fieldActive[5] {
			medFields = append(medFields, m.renderField(5))
		}
		groups = append(groups, styles.Panel.Render(strings.Join(medFields, "\n")))
	}

	// Period
	if m.fieldActive[6] {
		var periodFields []string
		periodFields = append(periodFields, styles.Subtitle.Render("📆 Periodo"))
		periodFields = append(periodFields, "")

		if m.selectedPeriod != nil {
			periodFields = append(periodFields, styles.Label.Render("Periodo:")+" "+
				styles.SuccessTxt.Render("✓ "+m.selectedPeriod.Label))
		} else {
			periodFields = append(periodFields, m.renderField(6))
		}
		groups = append(groups, styles.Panel.Render(strings.Join(periodFields, "\n")))
	}

	// Additional
	var addFields []string
	addFields = append(addFields, styles.Subtitle.Render("📝 Información adicional"))
	addFields = append(addFields, "")

	hasAdditionalFields := false
	for i := 7; i <= 10; i++ {
		if m.fieldActive[i] {
			addFields = append(addFields, m.renderField(i))
			hasAdditionalFields = true
		}
	}
	if hasAdditionalFields {
		groups = append(groups, styles.Panel.Render(strings.Join(addFields, "\n")))
	}

	// Render groups
	for _, group := range groups {
		s += group + "\n"
	}

	// Error
	if m.errorMsg != "" {
		s += "\n" + styles.ErrorTxt.Render("✗ "+m.errorMsg) + "\n"
	}

	if m.loading {
		s += "\n" + styles.InfoText.Render("● Procesando...") + "\n"
	} else {
		var helpText string
		if m.focusIndex == 2 && m.fieldActive[2] {
			helpText = "💡 Escribe nombre o #Empleado → Enter para buscar · Tab: siguiente · Ctrl+S: capturar"
		} else if m.focusIndex == 6 && m.fieldActive[6] {
			helpText = "Enter: seleccionar periodo · Tab: siguiente · Ctrl+S: capturar"
		} else {
			helpText = "Tab/Enter: siguiente · Shift+Tab: anterior · Ctrl+S: capturar"
		}
		s += "\n" + styles.Muted.Render(helpText)
	}

	return s
}

func (m CaptureModel) renderField(idx int) string {
	f := m.fieldAt(idx)
	if f == nil {
		return ""
	}

	label := styles.Label.Render(m.fieldLabels[idx] + ":")

	if idx == 2 && m.selectedDoctor != nil {
		return label + " " + styles.InfoText.Render(fmt.Sprintf("%s - %s", m.selectedDoctor.NumEmpleado, m.selectedDoctor.FullName))
	} else if idx == 6 && m.selectedPeriod != nil {
		return label + " " + styles.InfoText.Render(m.selectedPeriod.Label)
	}

	return label + " " + f.View()
}

func (m CaptureModel) renderDoctorSelection() string {
	var s string
	s += styles.ScreenTitle("Seleccionar médico", "Elige el médico que respalda la incidencia")
	s += "\n\n"

	headers := []string{"Empleado", "Nombre"}
	colWidths := []int{10, 50}

	tbl := NewTable(headers)
	tbl.Cursor = m.doctorCursor
	tbl.Offset = 0
	tbl.PageSz = 10

	for _, d := range m.doctors {
		tbl.Rows = append(tbl.Rows, []string{d.NumEmpleado, d.FullName})
	}

	s += tbl.Render(colWidths)
	s += "\n" + styles.Muted.Render("↑↓ navegar · Enter seleccionar · Esc cancelar")

	return s
}

func (m CaptureModel) renderPeriodSelection() string {
	var s string
	s += styles.ScreenTitle("Seleccionar periodo", "Elige el periodo vacacional")
	s += "\n\n"

	headers := []string{"Periodo"}
	colWidths := []int{60}

	tbl := NewTable(headers)
	tbl.Cursor = m.periodCursor
	tbl.Offset = 0
	tbl.PageSz = 10

	for _, p := range m.periods {
		tbl.Rows = append(tbl.Rows, []string{p.Label})
	}

	s += tbl.Render(colWidths)
	s += "\n" + styles.Muted.Render("↑↓ navegar · Enter seleccionar · Esc cancelar")

	return s
}

func (m *CaptureModel) nextField() {
	cur := m.fieldAt(m.focusIndex)
	if cur != nil {
		cur.Blur()
		cur.TextStyle = styles.InputBox
	}

	total := m.totalFields()
	for i := 1; i <= total; i++ {
		idx := (m.focusIndex + i) % total
		if m.fieldActive[idx] {
			m.focusIndex = idx
			break
		}
	}

	next := m.fieldAt(m.focusIndex)
	if next != nil {
		next.Focus()
		next.TextStyle = styles.InputFocused
	}
}

func (m *CaptureModel) prevField() {
	cur := m.fieldAt(m.focusIndex)
	if cur != nil {
		cur.Blur()
		cur.TextStyle = styles.InputBox
	}

	total := m.totalFields()
	for i := 1; i <= total; i++ {
		idx := (m.focusIndex - i + total) % total
		if m.fieldActive[idx] {
			m.focusIndex = idx
			break
		}
	}

	next := m.fieldAt(m.focusIndex)
	if next != nil {
		next.Focus()
		next.TextStyle = styles.InputFocused
	}
}

func (m *CaptureModel) doCapture() tea.Cmd {
	m.errorMsg = ""
	req, err := m.buildRequest()
	if err != nil {
		m.errorMsg = err.Error()
		return nil
	}

	m.loading = true

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

func (m *CaptureModel) buildRequest() (models.StoreIncidenciaRequest, error) {
	if strings.TrimSpace(m.fechaInicio.Value()) == "" {
		return models.StoreIncidenciaRequest{}, fmt.Errorf("Fecha de inicio es requerida")
	}
	startDate, fechaInicio, err := parseDate("Fecha de inicio", m.fechaInicio.Value())
	if err != nil {
		return models.StoreIncidenciaRequest{}, err
	}

	req := models.StoreIncidenciaRequest{
		EmployeeID:  m.employee.ID,
		Codigo:      m.code.ID,
		FechaInicio: fechaInicio,
	}

	if m.fieldActive[1] {
		if strings.TrimSpace(m.fechaFinal.Value()) == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Fecha final es requerida")
		}
		endDate, fechaFinal, err := parseDate("Fecha final", m.fechaFinal.Value())
		if err != nil {
			return models.StoreIncidenciaRequest{}, err
		}
		if endDate.Before(startDate) {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Fecha final no puede ser anterior a fecha de inicio")
		}
		req.FechaFinal = fechaFinal
	} else {
		req.FechaFinal = req.FechaInicio
	}

	if m.fieldActive[2] {
		medicoID, err := m.selectedOrTypedDoctorID()
		if err != nil {
			return models.StoreIncidenciaRequest{}, err
		}
		req.MedicoID = &medicoID
	}
	if m.fieldActive[3] {
		_, fechaExpedida, err := parseDate("Fecha expedida", m.fechaExpedida.Value())
		if err != nil {
			return models.StoreIncidenciaRequest{}, err
		}
		req.FechaExpedida = fechaExpedida
	}
	if m.fieldActive[4] {
		req.Diagnostico = strings.TrimSpace(m.diagnostico.Value())
		if req.Diagnostico == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Diagnóstico es requerido")
		}
	}
	if m.fieldActive[5] {
		req.NumLicencia = strings.TrimSpace(m.numLicencia.Value())
		if req.NumLicencia == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Núm. licencia es requerido")
		}
	}
	if m.fieldActive[6] {
		periodoID, err := m.selectedOrTypedPeriodID()
		if err != nil {
			return models.StoreIncidenciaRequest{}, err
		}
		req.PeriodoID = &periodoID
	}
	if m.fieldActive[7] {
		req.AutorizaTxt = strings.TrimSpace(m.autorizaTxt.Value())
		if req.AutorizaTxt == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Autoriza TXT es requerido")
		}
	}
	if m.fieldActive[8] {
		req.CoberturaTxt = strings.TrimSpace(m.coberturaTxt.Value())
		if req.CoberturaTxt == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Cobertura TXT es requerido")
		}
	}
	if m.fieldActive[9] {
		req.MotivoComision = strings.TrimSpace(m.motivoComision.Value())
		if req.MotivoComision == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Motivo comisión es requerido")
		}
	}
	if m.fieldActive[10] {
		req.Otorgado = strings.TrimSpace(m.otorgado.Value())
		if req.Otorgado == "" {
			return models.StoreIncidenciaRequest{}, fmt.Errorf("Otorgado es requerido")
		}
	}

	return req, nil
}

func (m *CaptureModel) selectedOrTypedDoctorID() (int, error) {
	if m.selectedDoctor != nil {
		return m.selectedDoctor.ID, nil
	}
	return 0, fmt.Errorf("Busca y selecciona un médico válido con Enter")
}

func (m *CaptureModel) selectedOrTypedPeriodID() (int, error) {
	if m.selectedPeriod != nil {
		return m.selectedPeriod.ID, nil
	}
	return 0, fmt.Errorf("Selecciona un periodo válido con Enter")
}

func parseDate(label, value string) (time.Time, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, "", fmt.Errorf("%s es requerida", label)
	}

	normalized := value
	if len(value) == 8 && !strings.Contains(value, "-") {
		normalized = value[:4] + "-" + value[4:6] + "-" + value[6:8]
	}

	parsed, err := time.Parse("2006-01-02", normalized)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("%s debe tener formato YYYYMMDD o YYYY-MM-DD", label)
	}
	return parsed, normalized, nil
}

// Unused but needed for compilation
var _ = lipgloss.NewStyle
