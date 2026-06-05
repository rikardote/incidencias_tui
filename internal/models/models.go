package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

// User represents the authenticated user
type User struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Username    string       `json:"username"`
	Type        string       `json:"type"`
	CanCapture  bool         `json:"can_capture"`
	IsAdmin     bool         `json:"is_admin"`
	Departments []Department `json:"departments"`
}

// LoginResponse from POST /api/v1/login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Employee from GET /api/v1/employees
type Employee struct {
	ID          int         `json:"id"`
	NumEmpleado string      `json:"num_empleado"`
	FullName    string      `json:"full_name"`
	Department  *Department `json:"department"`
	Puesto      string      `json:"puesto"`
	Horario     string      `json:"horario"`
	Jornada     string      `json:"jornada"`
}

// Department from GET /api/v1/departments
type Department struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// IncidenceCode from GET /api/v1/incidence-codes
type IncidenceCode struct {
	ID               int    `json:"id"`
	Code             string `json:"code"`
	Description      string `json:"description"`
	RequiresRange    bool   `json:"requires_range"`
	RequiresMedico   bool   `json:"requires_medico"`
	RequiresPeriodo  bool   `json:"requires_periodo"`
	RequiresTxt      bool   `json:"requires_txt"`
	RequiresComision bool   `json:"requires_comision"`
	RequiresOtorgado bool   `json:"requires_otorgado"`
	IsIncapacidad    bool   `json:"is_incapacidad"`
	IsLicencia       bool   `json:"is_licencia"`
	IsVacacional     bool   `json:"is_vacacional"`
}

// Doctor from GET /api/v1/doctors
type Doctor struct {
	ID          int    `json:"id"`
	NumEmpleado string `json:"num_empleado"`
	FullName    string `json:"full_name"`
}

// Periodo from GET /api/v1/periodos
type Periodo struct {
	ID      int    `json:"id"`
	Periodo int    `json:"periodo"`
	Year    int    `json:"year"`
	Label   string `json:"label"`
}

// UnmarshalJSON accepts numeric IDs from Laravel as either numbers or strings.
func (p *Periodo) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID      flexibleInt `json:"id"`
		Periodo flexibleInt `json:"periodo"`
		Year    flexibleInt `json:"year"`
		Label   string      `json:"label"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.ID = int(raw.ID)
	p.Periodo = int(raw.Periodo)
	p.Year = int(raw.Year)
	p.Label = raw.Label
	if p.Label == "" {
		p.Label = periodoLabel(p.Periodo, p.Year)
	}
	return nil
}

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*i = flexibleInt(n)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		*i = 0
		return nil
	}
	if slash := strings.Index(s, "/"); slash > 0 {
		s = s[:slash]
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		*i = 0
		return nil
	}
	*i = flexibleInt(parsed)
	return nil
}

func periodoLabel(periodo, year int) string {
	if periodo == 0 && year == 0 {
		return "Periodo sin etiqueta"
	}
	if year == 0 {
		return "Periodo " + strconv.Itoa(periodo)
	}
	if periodo == 0 {
		return strconv.Itoa(year)
	}
	return "Periodo " + strconv.Itoa(periodo) + "/" + strconv.Itoa(year)
}

// QNA from GET /api/v1/qnas
type QNA struct {
	ID          int    `json:"id"`
	Qna         string `json:"qna"`
	Year        int    `json:"year"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Cierre      string `json:"cierre"`
}

// StoreIncidenciaRequest for POST /api/v1/incidencias
type StoreIncidenciaRequest struct {
	EmployeeID           int    `json:"employee_id"`
	Codigo               int    `json:"codigo"`
	FechaInicio          string `json:"fecha_inicio"`
	FechaFinal           string `json:"fecha_final"`
	MedicoID             *int   `json:"medico_id,omitempty"`
	FechaExpedida        string `json:"fecha_expedida,omitempty"`
	Diagnostico          string `json:"diagnostico,omitempty"`
	NumLicencia          string `json:"num_licencia,omitempty"`
	PeriodoID            *int   `json:"periodo_id,omitempty"`
	AutorizaTxt          string `json:"autoriza_txt,omitempty"`
	CoberturaTxt         string `json:"cobertura_txt,omitempty"`
	MotivoComision       string `json:"motivo_comision,omitempty"`
	Otorgado             string `json:"otorgado,omitempty"`
	SaltarValidacionInca bool   `json:"saltar_validacion_inca"`
	SaltarValidacionLic  bool   `json:"saltar_validacion_lic"`
}

// StoreIncidenciaResponse from POST /api/v1/incidencias
type StoreIncidenciaResponse struct {
	Message    string `json:"message"`
	Token      string `json:"token"`
	EmployeeID int    `json:"employee_id"`
}

// APIResponse wrapper for list endpoints
type APIResponse[T any] struct {
	Data []T `json:"data"`
}

// IncidenceRecord for report list
type IncidenceRecord struct {
	ID             int            `json:"id"`
	Token          string         `json:"token"`
	FechaCapturado string         `json:"fecha_capturado"`
	Employee       *Employee      `json:"employee"`
	Codigo         *IncidenceCode `json:"codigo"`
	Qna            string         `json:"qna"`
	FechaInicio    string         `json:"fecha_inicio"`
	FechaFinal     string         `json:"fecha_final"`
	TotalDias      float64        `json:"total_dias"`
	CapturadoPor   string         `json:"capturado_por"`
}

// UnmarshalJSON accepts recent-report QNA as either a string or object.
func (r *IncidenceRecord) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID             int             `json:"id"`
		Token          string          `json:"token"`
		FechaCapturado string          `json:"fecha_capturado"`
		Employee       *Employee       `json:"employee"`
		Codigo         *IncidenceCode  `json:"codigo"`
		Qna            json.RawMessage `json:"qna"`
		FechaInicio    string          `json:"fecha_inicio"`
		FechaFinal     string          `json:"fecha_final"`
		TotalDias      float64         `json:"total_dias"`
		CapturadoPor   string          `json:"capturado_por"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.ID = raw.ID
	r.Token = raw.Token
	r.FechaCapturado = raw.FechaCapturado
	r.Employee = raw.Employee
	r.Codigo = raw.Codigo
	r.FechaInicio = raw.FechaInicio
	r.FechaFinal = raw.FechaFinal
	r.TotalDias = raw.TotalDias
	r.CapturadoPor = raw.CapturadoPor
	r.Qna = qnaLabel(raw.Qna)
	return nil
}

func qnaLabel(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return s
	}
	var q QNA
	if err := json.Unmarshal(data, &q); err == nil {
		// Construir formato Q{número}/{año} en lugar de usar Description
		if q.Qna != "" && q.Year > 0 {
			return q.Qna + "/" + strconv.Itoa(q.Year)
		}
		if q.Qna != "" {
			return q.Qna
		}
	}
	return ""
}

// EmployeeReport for employee report
type EmployeeReport struct {
	ID             int            `json:"id"`
	Token          string         `json:"token"`
	Codigo         *IncidenceCode `json:"codigo"`
	Qna            *QNA           `json:"qna"`
	Periodo        *Periodo       `json:"periodo"`
	FechaInicio    string         `json:"fecha_inicio"`
	FechaFinal     string         `json:"fecha_final"`
	TotalDias      float64        `json:"total_dias"`
	FechaCapturado string         `json:"fecha_capturado"`
	CapturadoPor   string         `json:"capturado_por"`
	Diagnostico    string         `json:"diagnostico"`
	NumLicencia    string         `json:"num_licencia"`
	FechaExpedida  string         `json:"fecha_expedida"`
	Otorgado       string         `json:"otorgado"`
	CoberturaTxt   string         `json:"cobertura_txt"`
	AutorizaTxt    string         `json:"autoriza_txt"`
	MotivoComision string         `json:"motivo_comision"`
}

// QNASummary for qna summary report
type QNASummary struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Registros   int    `json:"registros"`
	Dias        int    `json:"dias"`
}

// VacationResponse for employee vacation balance
type VacationResponse struct {
	Employee     Employee         `json:"employee"`
	Entitlement  float64          `json:"entitlement"`
	TotalPending float64          `json:"total_pending"`
	Periods      []VacationPeriod `json:"periods"`
}

// VacationPeriod represents vacation data for a specific period
type VacationPeriod struct {
	Period      VacationPeriodInfo `json:"period"`
	Entitlement float64            `json:"entitlement"`
	Used        float64            `json:"used"`
	Pending     float64            `json:"pending"`
	Incidencias []EmployeeReport   `json:"incidencias"`
}

// VacationPeriodInfo contains period identification
type VacationPeriodInfo struct {
	ID     int    `json:"id"`
	Period string `json:"period"`
	Year   int    `json:"year"`
	Label  string `json:"label"`
}

// BiometricRecord for recent biometric data
type BiometricRecord struct {
	ID            int       `json:"id"`
	NumEmpleado   string    `json:"num_empleado"`
	Employee      *Employee `json:"employee"`
	Fecha         string    `json:"fecha"`
	Hora          string    `json:"hora"`
	Timestamp     int64     `json:"timestamp"`
	Identificador string    `json:"identificador"`
	Location      string    `json:"location"`
}

// AttendanceDay for employee attendance
type AttendanceDay struct {
	Date              string   `json:"date"`
	PrimeraChecada    string   `json:"primera_checada"`
	UltimaChecada     string   `json:"ultima_checada"`
	HoraEntrada       string   `json:"hora_entrada"`
	HoraSalida        string   `json:"hora_salida"`
	NumChecadas       int      `json:"num_checadas"`
	Retardo           bool     `json:"retardo"`
	Incidencias       []string `json:"incidencias"`
	IncidenciasTokens []string `json:"incidencias_tokens"`
}

// AttendanceResponse for biometric employee attendance
type AttendanceResponse struct {
	Employee Employee        `json:"employee"`
	Start    string          `json:"start"`
	End      string          `json:"end"`
	Data     []AttendanceDay `json:"data"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Message string `json:"message"`
}
