package api

import (
	"encoding/json"
	"testing"

	"incidencias_tui/internal/models"
)

func TestListResponseAcceptsWrappedData(t *testing.T) {
	body := []byte(`{"data":[{"id":1,"periodo":1,"year":2026,"label":"Periodo 1/2026"}]}`)
	var result listResponse[models.Periodo]

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ID != 1 {
		t.Fatalf("unexpected result: %+v", result.Data)
	}
}

func TestListResponseAcceptsDirectArray(t *testing.T) {
	body := []byte(`[{"id":2,"periodo":"2","year":"2026","label":"Periodo 2/2026"}]`)
	var result listResponse[models.Periodo]

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].Periodo != 2 {
		t.Fatalf("unexpected result: %+v", result.Data)
	}
}

func TestListResponseAcceptsLaravelPaginatedData(t *testing.T) {
	body := []byte(`{"data":{"data":[{"id":3,"periodo":3,"year":2026,"label":"Periodo 3/2026"}]}}`)
	var result listResponse[models.Periodo]

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ID != 3 {
		t.Fatalf("unexpected result: %+v", result.Data)
	}
}

func TestErrorMessageFromBodyUsesBusinessMessageOnly(t *testing.T) {
	body := []byte(`{"message":"Incidencia Duplicada"}`)

	got := errorMessageFromBody(body)
	if got != "Incidencia Duplicada" {
		t.Fatalf("message = %q, want Incidencia Duplicada", got)
	}
}

func TestErrorMessageFromBodyIncludesLaravelValidationErrors(t *testing.T) {
	body := []byte(`{
		"message": "The given data was invalid.",
		"errors": {
			"fecha_inicio": ["Fecha de inicio es requerida"],
			"periodo_id": ["Debe seleccionar un periodo vacacional"]
		}
	}`)

	got := errorMessageFromBody(body)
	want := "The given data was invalid. · Fecha de inicio es requerida · Debe seleccionar un periodo vacacional"
	if got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestAPIErrorStringReturnsCleanMessage(t *testing.T) {
	err := APIError{StatusCode: 422, Message: "Trabajador ya tomó 3 días económicos"}

	if err.Error() != "Trabajador ya tomó 3 días económicos" {
		t.Fatalf("error = %q", err.Error())
	}
}
