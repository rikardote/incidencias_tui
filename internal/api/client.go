package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"incidencias_tui/internal/models"
)

// Client is the HTTP API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New creates a new API client
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
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
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("%s respondió %d: %s", fullURL, resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("%s respondió %d: %s", fullURL, resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}
	}

	return nil
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
	var result models.APIResponse[models.Employee]
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
	var result models.APIResponse[models.IncidenceCode]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetDoctors searches doctors by query
func (c *Client) GetDoctors(query string) ([]models.Doctor, error) {
	path := "/api/v1/doctors?search=" + url.QueryEscape(query)
	var result models.APIResponse[models.Doctor]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetPeriodos returns all periods
func (c *Client) GetPeriodos() ([]models.Periodo, error) {
	var result models.APIResponse[models.Periodo]
	if err := c.doRequest("GET", "/api/v1/periodos", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetQNAs returns all quincenas
func (c *Client) GetQNAs() ([]models.QNA, error) {
	var result models.APIResponse[models.QNA]
	if err := c.doRequest("GET", "/api/v1/qnas", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetDepartments returns departments
func (c *Client) GetDepartments() ([]models.Department, error) {
	var result models.APIResponse[models.Department]
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
	var result models.APIResponse[models.IncidenceRecord]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEmployeeReport returns incidence history for an employee
func (c *Client) GetEmployeeReport(employeeID int, start, end string) ([]models.EmployeeReport, error) {
	path := fmt.Sprintf("/api/v1/reports/employee/%d?start=%s&end=%s", employeeID, start, end)
	var result struct {
		Employee models.Employee      `json:"employee"`
		Data     []models.EmployeeReport `json:"data"`
	}
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetQNASummary returns QNA summary report
func (c *Client) GetQNASummary(qnaID, departmentID int) ([]models.QNASummary, error) {
	path := fmt.Sprintf("/api/v1/reports/qna-summary?qna_id=%d&department_id=%d", qnaID, departmentID)
	var result models.APIResponse[models.QNASummary]
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
	var result models.APIResponse[models.BiometricRecord]
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEmployeeAttendance returns attendance for an employee
func (c *Client) GetEmployeeAttendance(employeeID int, start, end string) (*models.AttendanceResponse, error) {
	path := fmt.Sprintf("/api/v1/biometrico/employee/%d/attendance?start=%s&end=%s", employeeID, start, end)
	var result models.AttendanceResponse
	if err := c.doRequest("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
