package handlers

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"vitalis/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockDB swaps database.Database for a sqlmock-backed *sql.DB for the
// duration of the test and restores the original afterwards.
func mockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	original := database.Database
	database.Database = db

	t.Cleanup(func() {
		database.Database = original
		db.Close()
	})

	return mock
}

// newTestContext builds a gin.Context/ResponseRecorder pair for a request
// with the given method, path, body and headers, without going through a
// real HTTP server.
func newTestContext(method, target string, body io.Reader, contentType string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, target, body)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	c.Request = req

	return c, w
}

// expectValidToken sets up the sqlmock expectations that CheckToken's
// underlying services.CheckSessionKey call needs to see in order to
// consider a session token valid.
func expectValidToken(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`select now\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"now"}).AddRow(time.Now()))

	mock.ExpectQuery(`select id from auth_session_keys`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
}

// expectInvalidToken sets up the sqlmock expectations for a session token
// that does not correspond to any valid, non-expired row.
func expectInvalidToken(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`select now\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"now"}).AddRow(time.Now()))

	mock.ExpectQuery(`select id from auth_session_keys`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
}
