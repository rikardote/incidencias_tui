package views

import (
	"reflect"
	"testing"

	"incidencias_tui/internal/models"
)

func testCaptureModel(code models.IncidenceCode) CaptureModel {
	emp := &models.Employee{ID: 10, NumEmpleado: "332618", FullName: "Test Employee"}
	user := &models.User{ID: 1, Name: "Tester"}
	return NewCaptureModel(nil, emp, &code, user)
}

func TestBuildRequestUsesStartDateAsFinalDateForSingleDayCode(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 7, Code: "A1"})
	m.fechaInicio.SetValue("2026-06-05")

	req, err := m.buildRequest()
	if err != nil {
		t.Fatalf("buildRequest returned error: %v", err)
	}

	if req.EmployeeID != 10 {
		t.Fatalf("EmployeeID = %d, want 10", req.EmployeeID)
	}
	if req.Codigo != 7 {
		t.Fatalf("Codigo = %d, want 7", req.Codigo)
	}
	if req.FechaFinal != "2026-06-05" {
		t.Fatalf("FechaFinal = %q, want start date", req.FechaFinal)
	}
}

func TestBuildRequestRequiresValidDate(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 7, Code: "A1"})
	m.fechaInicio.SetValue("05/06/2026")

	if _, err := m.buildRequest(); err == nil {
		t.Fatal("buildRequest returned nil error for invalid date")
	}
}

func TestBuildRequestRejectsEndDateBeforeStartDate(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 8, Code: "RNG", RequiresRange: true})
	m.fechaInicio.SetValue("2026-06-05")
	m.fechaFinal.SetValue("2026-06-04")

	if _, err := m.buildRequest(); err == nil {
		t.Fatal("buildRequest returned nil error for reversed date range")
	}
}

func TestBuildRequestNormalizesCompactDates(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 8, Code: "41", IsLicencia: true})
	m.fechaInicio.SetValue("20260601")
	m.fechaFinal.SetValue("20260605")

	req, err := m.buildRequest()
	if err != nil {
		t.Fatalf("buildRequest returned error: %v", err)
	}
	if req.FechaInicio != "2026-06-01" {
		t.Fatalf("FechaInicio = %q, want 2026-06-01", req.FechaInicio)
	}
	if req.FechaFinal != "2026-06-05" {
		t.Fatalf("FechaFinal = %q, want 2026-06-05", req.FechaFinal)
	}
}

func TestBuildRequestSetsDoctorAndPeriodIDs(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{
		ID:              9,
		Code:            "55",
		RequiresPeriodo: true,
	})
	m.fechaInicio.SetValue("2026-06-05")
	m.fechaFinal.SetValue("2026-06-06")
	m.fechaExpedida.SetValue("20260605")
	m.diagnostico.SetValue("DIAGNOSTICO")
	m.numLicencia.SetValue("LIC-123")
	m.selectedDoctor = &models.Doctor{ID: 22, NumEmpleado: "MED1", FullName: "Doctor Test"}
	m.selectedPeriod = &models.Periodo{ID: 33, Label: "2026 - Periodo 1"}

	req, err := m.buildRequest()
	if err != nil {
		t.Fatalf("buildRequest returned error: %v", err)
	}
	if req.MedicoID == nil || *req.MedicoID != 22 {
		t.Fatalf("MedicoID = %v, want 22", req.MedicoID)
	}
	if req.PeriodoID == nil || *req.PeriodoID != 33 {
		t.Fatalf("PeriodoID = %v, want 33", req.PeriodoID)
	}
	if req.FechaExpedida != "2026-06-05" {
		t.Fatalf("FechaExpedida = %q, want 2026-06-05", req.FechaExpedida)
	}
}

func TestBuildRequestRequiresDoctorSelection(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 9, Code: "55"})
	m.fechaInicio.SetValue("2026-06-05")
	m.medico.SetValue("Doctor sin seleccionar")

	if _, err := m.buildRequest(); err == nil {
		t.Fatal("buildRequest returned nil error without selected doctor")
	}
}

func TestBuildRequestDoesNotSkipValidationsAutomaticallyForAdmin(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 41, Code: "41", IsLicencia: true})
	m.user.IsAdmin = true
	m.fechaInicio.SetValue("2026-06-01")
	m.fechaFinal.SetValue("2026-06-01")

	req, err := m.buildRequest()
	if err != nil {
		t.Fatalf("buildRequest returned error: %v", err)
	}
	if req.SaltarValidacionInca {
		t.Fatal("SaltarValidacionInca should not be enabled automatically")
	}
	if req.SaltarValidacionLic {
		t.Fatal("SaltarValidacionLic should not be enabled automatically")
	}
}

func TestCaptureFieldsDoNotRequireDiagnosisForCode41(t *testing.T) {
	m := testCaptureModel(models.IncidenceCode{ID: 41, Code: "41", IsLicencia: true})

	if m.fieldActive[4] {
		t.Fatal("diagnosis field is active for code 41")
	}
	if m.fieldActive[2] {
		t.Fatal("doctor field is active for code 41")
	}
	if m.fieldActive[3] {
		t.Fatal("issued date field is active for code 41")
	}
	if m.fieldActive[5] {
		t.Fatal("license number field is active for code 41")
	}
}

func TestCaptureFieldsRequireMedicalDetailsForCodes53To55(t *testing.T) {
	for _, code := range []string{"53", "54", "00055"} {
		m := testCaptureModel(models.IncidenceCode{ID: 1, Code: code})
		if !m.fieldActive[2] {
			t.Fatalf("doctor field is inactive for code %s", code)
		}
		if !m.fieldActive[3] {
			t.Fatalf("issued date field is inactive for code %s", code)
		}
		if !m.fieldActive[4] {
			t.Fatalf("diagnosis field is inactive for code %s", code)
		}
		if !m.fieldActive[5] {
			t.Fatalf("license number field is inactive for code %s", code)
		}
	}
}

func TestCaptureFieldMatrixMatchesBackendConcepts(t *testing.T) {
	tests := []struct {
		name string
		code models.IncidenceCode
		want []int
	}{
		{
			name: "licencia 41 only requires date range",
			code: models.IncidenceCode{ID: 41, Code: "41", IsLicencia: true},
			want: []int{0, 1},
		},
		{
			name: "incapacidad 55 requires medical block",
			code: models.IncidenceCode{ID: 55, Code: "55", IsIncapacidad: true},
			want: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name: "vacaciones 60 requires period",
			code: models.IncidenceCode{ID: 60, Code: "60", IsVacacional: true},
			want: []int{0, 1, 6},
		},
		{
			name: "txt 900 requires txt fields",
			code: models.IncidenceCode{ID: 900, Code: "900"},
			want: []int{0, 7, 8},
		},
		{
			name: "comision 61 requires range and reason",
			code: models.IncidenceCode{ID: 61, Code: "61"},
			want: []int{0, 1, 9},
		},
		{
			name: "otorgado 901 requires granted-by field",
			code: models.IncidenceCode{ID: 901, Code: "901"},
			want: []int{0, 10},
		},
		{
			name: "onomastico 14 only requires start date",
			code: models.IncidenceCode{ID: 14, Code: "14"},
			want: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testCaptureModel(tt.code)
			if got := activeFieldIndexes(m); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("active fields = %v, want %v", got, tt.want)
			}
		})
	}
}

func activeFieldIndexes(m CaptureModel) []int {
	var indexes []int
	for i, active := range m.fieldActive {
		if active {
			indexes = append(indexes, i)
		}
	}
	return indexes
}
