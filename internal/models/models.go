package models

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
	ID          int        `json:"id"`
	NumEmpleado string     `json:"num_empleado"`
	FullName    string     `json:"full_name"`
	Department  Department `json:"department"`
	Puesto      string     `json:"puesto"`
	Horario     string     `json:"horario"`
	Jornada     string     `json:"jornada"`
}

// Department from GET /api/v1/departments
type Department struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// IncidenceCode from GET /api/v1/incidence-codes
type IncidenceCode struct {
	ID                int    `json:"id"`
	Code              string `json:"code"`
	Description       string `json:"description"`
	RequiresRange     bool   `json:"requires_range"`
	RequiresMedico    bool   `json:"requires_medico"`
	RequiresPeriodo   bool   `json:"requires_periodo"`
	RequiresTxt       bool   `json:"requires_txt"`
	RequiresComision  bool   `json:"requires_comision"`
	RequiresOtorgado  bool   `json:"requires_otorgado"`
	IsIncapacidad     bool   `json:"is_incapacidad"`
	IsLicencia        bool   `json:"is_licencia"`
	IsVacacional      bool   `json:"is_vacacional"`
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
	FechaCapturado string   `json:"fecha_capturado"`
	Employee       Employee `json:"employee"`
	Codigo         struct {
		ID          int    `json:"id"`
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"codigo"`
	FechaInicio  string `json:"fecha_inicio"`
	FechaFinal   string `json:"fecha_final"`
	TotalDias    int    `json:"total_dias"`
	Qna          string `json:"qna"`
	CapturadoPor string `json:"capturado_por"`
}

// EmployeeReport for employee report
type EmployeeReport struct {
	Codigo      string `json:"codigo"`
	Description string `json:"description"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFinal  string `json:"fecha_final"`
	TotalDias   int    `json:"total_dias"`
	Qna         string `json:"qna"`
	Periodo     string `json:"periodo"`
}

// QNASummary for qna summary report
type QNASummary struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Registros   int    `json:"registros"`
	Dias        int    `json:"dias"`
}

// BiometricRecord for recent biometric data
type BiometricRecord struct {
	ID           int      `json:"id"`
	NumEmpleado  string   `json:"num_empleado"`
	Employee     Employee `json:"employee"`
	Fecha        string   `json:"fecha"`
	Hora         string   `json:"hora"`
	Timestamp    int64    `json:"timestamp"`
	Identificador string  `json:"identificador"`
	Location     string   `json:"location"`
}

// AttendanceDay for employee attendance
type AttendanceDay struct {
	Date           string   `json:"date"`
	PrimeraChecada string   `json:"primera_checada"`
	UltimaChecada  string   `json:"ultima_checada"`
	HoraEntrada    string   `json:"hora_entrada"`
	HoraSalida     string   `json:"hora_salida"`
	NumChecadas    int      `json:"num_checadas"`
	Retardo        bool     `json:"retardo"`
	Incidencias    []string `json:"incidencias"`
	IncidenciasTokens []string `json:"incidencias_tokens"`
}

// AttendanceResponse for biometric employee attendance
type AttendanceResponse struct {
	Employee Employee         `json:"employee"`
	Start    string           `json:"start"`
	End      string           `json:"end"`
	Data     []AttendanceDay  `json:"data"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Message string `json:"message"`
}
