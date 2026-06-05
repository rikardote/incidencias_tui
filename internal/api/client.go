package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"incidencias_tui/internal/models"
)

// Client is the HTTP API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("La API respondió con estado %d", e.StatusCode)
	}
	return "Error de API"
}

type listResponse[T any] struct {
	Data []T
}

func (r *listResponse[T]) UnmarshalJSON(data []byte) error {
	var direct []T
	if err := json.Unmarshal(data, &direct); err == nil {
		r.Data = direct
		return nil
	}

	var wrapped struct {
		Data []T `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Data != nil {
		r.Data = wrapped.Data
		return nil
	}

	var paginated struct {
		Data struct {
			Data []T `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &paginated); err == nil && paginated.Data.Data != nil {
		r.Data = paginated.Data.Data
		return nil
	}

	return fmt.Errorf("response does not contain a valid data list")
}

// New creates a new API client
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// logAPIError saves API error details to a log file for debugging
func logAPIError(endpoint string, respBody []byte) string {
	logDir := os.TempDir()
	logFile := filepath.Join(logDir, "incidencias-tui-api-debug.log")
	
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Sprintf("(no se pudo guardar log: %v)", err)
	}
	defer f.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("\n=== %s ===\nEndpoint: %s\nResponse length: %d bytes\nResponse body:\n%s\n", 
		timestamp, endpoint, len(respBody), string(respBody))
	
	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Sprintf("(no se pudo escribir log: %v)", err)
	}
	
	return logFile
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current token
func (c *Client) GetToken() string {
	return c.token
}

// GetBaseURL returns the base API URL
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// doRequestRaw performs an HTTP request and returns the raw response body
func (c *Client) doRequestRaw(method, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	fullURL := c.baseURL + path
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error en %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %w", fullURL, err)
	}

	if resp.StatusCode >= 400 {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    errorMessageFromBody(respBody),
		}
	}

	return respBody, nil
}

// doRequest performs an HTTP request with optional auth
func (c *Client) doRequest(method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	fullURL := c.baseURL + path
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error en %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response from %s: %w", fullURL, err)
	}

	if resp.StatusCode >= 400 {
		return APIError{
			StatusCode: resp.StatusCode,
			Message:    errorMessageFromBody(respBody),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("error parsing response from %s: %w; body: %s", fullURL, err, previewBody(respBody))
		}
	}

	return nil
}

func errorMessageFromBody(body []byte) string {
	var payload struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		var parts []string
		if strings.TrimSpace(payload.Message) != "" {
			parts = append(parts, payload.Message)
		}
		for _, messages := range payload.Errors {
			parts = append(parts, messages...)
		}
		if len(parts) > 0 {
			return strings.Join(uniqueStrings(parts), " · ")
		}
	}

	text := strings.TrimSpace(string(body))
	if text == "" {
		return "La API no devolvió detalle del error"
	}
	return previewBody([]byte(text))
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func previewBody(body []byte) string {
	const maxLen = 300
	text := string(body)
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// --- Auth ---

// Login authenticates and returns a token
func (c *Client) Login(username, password, deviceName string) (*models.LoginResponse, error) {
	body := map[string]string{
		"username":    username,
		"password":    password,
		"device_name": deviceName,
	}
	var result models.LoginResponse
	if err := c.doRequest("POST", "/api/v1/login", body, &result); err != nil {
		return nil, err
	}
	c.token = result.Token
	return &result, nil
}

// Logout invalidates the current token
func (c *Client) Logout() error {
	return c.doRequest("POST", "/api/v1/logout", nil, nil)
}

// Me returns the current user info
func (c *Client) Me() (*models.User, error) {
	var result models.User
	if err := c.doRequest("GET", "/api/v1/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Employees ---

// SearchEmployees searches employees by query
func (c *Client) SearchEmployees(query string) ([]models.Employee, error) {
	path := "/api/v1/employees?search=" + url.QueryEscape(query)
	var result listResponse[models.Employee]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// --- Catalogs ---

// GetIncidenceCodes returns incidence codes, optionally filtered
func (c *Client) GetIncidenceCodes(query string) ([]models.IncidenceCode, error) {
	path := "/api/v1/incidence-codes"
	if query != "" {
		path += "?search=" + url.QueryEscape(query)
	}
	var result listResponse[models.IncidenceCode]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetDoctors searches doctors by query
func (c *Client) GetDoctors(query string) ([]models.Doctor, error) {
	path := "/api/v1/doctors?search=" + url.QueryEscape(query)
	var result listResponse[models.Doctor]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetPeriodos returns all periods
func (c *Client) GetPeriodos() ([]models.Periodo, error) {
	var result listResponse[models.Periodo]
	if err := c.doRequest("GET", "/api/v1/periodos", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetQNAs returns all quincenas
func (c *Client) GetQNAs() ([]models.QNA, error) {
	var result listResponse[models.QNA]
	if err := c.doRequest("GET", "/api/v1/qnas", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetDepartments returns departments
func (c *Client) GetDepartments() ([]models.Department, error) {
	var result listResponse[models.Department]
	if err := c.doRequest("GET", "/api/v1/departments", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// --- Incidencias ---

// StoreIncidencia creates a new incidence record
func (c *Client) StoreIncidencia(req models.StoreIncidenciaRequest) (*models.StoreIncidenciaResponse, error) {
	var result models.StoreIncidenciaResponse
	if err := c.doRequest("POST", "/api/v1/incidencias", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteIncidencia deletes incidences by token
func (c *Client) DeleteIncidencia(token string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.doRequest("DELETE", "/api/v1/incidencias/"+url.PathEscape(token), nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Reports ---

// GetRecentIncidencias returns recent incidence records
func (c *Client) GetRecentIncidencias(limit int) ([]models.IncidenceRecord, error) {
	path := "/api/v1/reports/recent"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var result listResponse[models.IncidenceRecord]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEmployeeReport returns incidence history for an employee
func (c *Client) GetEmployeeReport(employeeID int, start, end string) ([]models.EmployeeReport, error) {
	path := fmt.Sprintf("/api/v1/reports/employee/%d?start=%s&end=%s", employeeID, start, end)
	respBody, err := c.doRequestRaw("GET", path)
	if err != nil {
		return nil, err
	}

	// Try multiple response formats
	// Format 1: {"data": [...]}
	var wrapped struct {
		Data []models.EmployeeReport `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}

	// Format 2: direct array [...]
	var direct []models.EmployeeReport
	if err := json.Unmarshal(respBody, &direct); err == nil {
		return direct, nil
	}

	// Format 3: paginated {"data": {"data": [...]}}
	var paginated struct {
		Data struct {
			Data []models.EmployeeReport `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &paginated); err == nil && paginated.Data.Data != nil {
		return paginated.Data.Data, nil
	}

	// Format 4: {"employee": {...}, "incidencias": [...]}
	var withIncidencias struct {
		Incidencias []models.EmployeeReport `json:"incidencias"`
		Records     []models.EmployeeReport `json:"records"`
		Report      []models.EmployeeReport `json:"report"`
	}
	if err := json.Unmarshal(respBody, &withIncidencias); err == nil {
		if withIncidencias.Incidencias != nil {
			return withIncidencias.Incidencias, nil
		}
		if withIncidencias.Records != nil {
			return withIncidencias.Records, nil
		}
		if withIncidencias.Report != nil {
			return withIncidencias.Report, nil
		}
	}

	logFile := logAPIError(path, respBody)
	return nil, fmt.Errorf("no se pudo interpretar la respuesta de la API. Detalles guardados en: %s", logFile)
}

// GetQNASummary returns QNA summary report
func (c *Client) GetQNASummary(qnaID, departmentID int) ([]models.QNASummary, error) {
	path := fmt.Sprintf("/api/v1/reports/qna-summary?qna_id=%d&department_id=%d", qnaID, departmentID)
	var result listResponse[models.QNASummary]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// --- Biometric ---

// GetRecentBiometric returns recent biometric records
func (c *Client) GetRecentBiometric(limit int) ([]models.BiometricRecord, error) {
	path := "/api/v1/biometrico/recent"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var result listResponse[models.BiometricRecord]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEmployeeAttendance returns attendance for an employee
func (c *Client) GetEmployeeAttendance(employeeID int, start, end string) (*models.AttendanceResponse, error) {
	path := fmt.Sprintf("/api/v1/biometrico/employee/%d/attendance?start=%s&end=%s", employeeID, start, end)
	respBody, err := c.doRequestRaw("GET", path)
	if err != nil {
		return nil, err
	}

	// Try multiple response formats
	// Format 1: direct AttendanceResponse
	var direct models.AttendanceResponse
	if err := json.Unmarshal(respBody, &direct); err == nil && direct.Employee.ID > 0 {
		return &direct, nil
	}

	// Format 2: {"data": {...}}
	var wrapped struct {
		Data models.AttendanceResponse `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapped); err == nil && wrapped.Data.Employee.ID > 0 {
		return &wrapped.Data, nil
	}

	// Format 3: {"employee": {...}, "data": [...]}
	var split struct {
		Employee models.Employee       `json:"employee"`
		Data     []models.AttendanceDay `json:"data"`
		Start    string                `json:"start"`
		End      string                `json:"end"`
	}
	if err := json.Unmarshal(respBody, &split); err == nil && split.Employee.ID > 0 {
		return &models.AttendanceResponse{
			Employee: split.Employee,
			Start:    split.Start,
			End:      split.End,
			Data:     split.Data,
		}, nil
	}

	logFile := logAPIError(path, respBody)
	return nil, fmt.Errorf("no se pudo interpretar la respuesta de asistencia. Detalles guardados en: %s", logFile)
}
