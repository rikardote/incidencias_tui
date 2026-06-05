package models

import (
	"encoding/json"
	"testing"
)

func TestIncidenceRecordAcceptsDocumentedRecentReportShape(t *testing.T) {
	body := []byte(`{
		"fecha_capturado": "2026-06-01 10:30",
		"employee": {
			"id": 123,
			"num_empleado": "001234",
			"full_name": "NOMBRE APELLIDOS"
		},
		"codigo": {
			"id": 10,
			"code": "60",
			"description": "VACACIONES"
		},
		"fecha_inicio": "2026-06-01",
		"fecha_final": "2026-06-05",
		"total_dias": 5,
		"qna": "11/2026",
		"capturado_por": "U."
	}`)

	var record IncidenceRecord
	if err := json.Unmarshal(body, &record); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if record.Employee == nil || record.Employee.NumEmpleado != "001234" {
		t.Fatalf("unexpected employee: %+v", record.Employee)
	}
	if record.Codigo == nil || record.Codigo.Code != "60" {
		t.Fatalf("unexpected code: %+v", record.Codigo)
	}
	if record.Qna != "11/2026" {
		t.Fatalf("Qna = %q, want 11/2026", record.Qna)
	}
}

func TestIncidenceRecordAcceptsQNAObject(t *testing.T) {
	body := []byte(`{
		"qna": {"qna": "11", "year": 2026, "description": "1ra quincena junio"},
		"total_dias": 1
	}`)

	var record IncidenceRecord
	if err := json.Unmarshal(body, &record); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if record.Qna != "11/2026" {
		t.Fatalf("Qna = %q, want 11/2026", record.Qna)
	}
}
