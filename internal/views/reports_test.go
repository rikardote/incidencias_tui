package views

import (
	"testing"

	"incidencias_tui/internal/models"
)

func TestReportModelCursorScrollsThroughLoadedResults(t *testing.T) {
	m := NewReportModel(nil)
	m.results = make([]models.IncidenceRecord, 30)
	m.pageSize = 12

	m.moveCursor(15)

	if m.cursor != 15 {
		t.Fatalf("cursor = %d, want 15", m.cursor)
	}
	if m.offset != 4 {
		t.Fatalf("offset = %d, want 4", m.offset)
	}
}

func TestReportModelCursorClampsAtEnd(t *testing.T) {
	m := NewReportModel(nil)
	m.results = make([]models.IncidenceRecord, 5)
	m.pageSize = 12

	m.moveCursor(100)

	if m.cursor != 4 {
		t.Fatalf("cursor = %d, want 4", m.cursor)
	}
}
