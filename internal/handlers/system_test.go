package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"vitalis/internal/handlers/structs"

	"github.com/DATA-DOG/go-sqlmock"
)

// multipartBody builds a multipart/form-data body from the given fields and
// returns it together with the Content-Type header (including boundary).
func multipartBody(t *testing.T, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for name, value := range fields {
		if err := w.WriteField(name, value); err != nil {
			t.Fatalf("failed to write field %q: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return &buf, w.FormDataContentType()
}

func TestGetFiltersWithValues_NoBody(t *testing.T) {
	c, _ := newTestContext(http.MethodPost, "/", nil, "")

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 0 || len(filterValues) != 0 {
		t.Errorf("filters = %v, filterValues = %v, want both empty", filters, filterValues)
	}
}

func TestGetFiltersWithValues_WithFilters(t *testing.T) {
	// getFiltersWithValues matches the posted field name against
	// InfoFilter.Name, which is unprefixed (e.g. "physical_cores_max"),
	// not the InfoFilters map key (e.g. "cpu_physical_cores_max").
	body, contentType := multipartBody(t, map[string]string{
		"physical_cores_max": "12",
		"unknown_field":      "ignored",
	})
	c, _ := newTestContext(http.MethodPost, "/", body, contentType)

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filters) != 1 {
		t.Fatalf("got %d filters, want 1: %+v", len(filters), filters)
	}
	if filters[0].Name != "physical_cores_max" {
		t.Errorf("filters[0].Name = %q, want %q", filters[0].Name, "physical_cores_max")
	}
	if len(filterValues) != 1 || filterValues[0] != "12" {
		t.Errorf("filterValues = %v, want [\"12\"]", filterValues)
	}
}

func TestCpuInformation_Unauthorized(t *testing.T) {
	mockDB(t) // no expectations set: any DB call would fail the test.

	c, w := newTestContext(http.MethodPost, "/", nil, "")

	CpuInformation(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCpuInformation_Success(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	mock.ExpectQuery(`select\s*\*\s*from public\.info_cpu ic`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "name", "physical_cores", "logical_cores",
			"utilization", "current_speed_mhz", "base_speed_mhz",
			"processes_amount", "threads_amount", "handles_amount",
			"uptime", "insertion_datetime",
		}).AddRow(1, 1, "Intel i7", 8, 16, 12.5, 4200.0, 3000.0, 200, 1500, 5000, 3600, time.Now().String()))

	c, w := newTestContext(http.MethodPost, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	CpuInformation(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp structs.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Errorf("resp.Status = false, want true")
	}
	if resp.Data == nil {
		t.Errorf("resp.Data = nil, want CPU rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCpuInformation_QueryError(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	mock.ExpectQuery(`select\s*\*\s*from public\.info_cpu ic`).
		WillReturnError(sqlmock.ErrCancelled)

	c, w := newTestContext(http.MethodPost, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	CpuInformation(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRamInformation_Success(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	mock.ExpectQuery(`select\s*\*\s*from public\.info_ram ir`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "total", "used", "free", "commited", "cached", "insertion_datetime",
		}).AddRow(1, 1, 16000, 8000, 8000, 500, 300, time.Now().String()))

	c, w := newTestContext(http.MethodPost, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	RamInformation(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestNetInformation_Success(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	mock.ExpectQuery(`select\s*\*\s*from public\.info_net nt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "bytes_sent", "bytes_recv", "packets_sent",
			"packets_recv", "err_in", "err_out", "connections", "insertion_datetime",
		}).AddRow(1, 1, 1000, 2000, 10, 20, 0, 0, 5, time.Now().String()))

	c, w := newTestContext(http.MethodPost, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	NetInformation(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFileInformation_Success(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	mock.ExpectQuery(`select\s*\*\s*from public\.info_file if`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "group_id", "path", "total", "used", "free", "used_percent", "insertion_datetime",
		}).AddRow(1, 1, "/", 100000, 40000, 60000, 40.0, time.Now().String()))

	c, w := newTestContext(http.MethodPost, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	FileInformation(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
