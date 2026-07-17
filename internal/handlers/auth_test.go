package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"vitalis/internal/enviroment"
	"vitalis/internal/handlers/structs"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCheckToken_NoAuthHeader(t *testing.T) {
	c, w := newTestContext(http.MethodGet, "/", nil, "")

	if got := CheckToken(c); got != false {
		t.Fatalf("CheckToken() = %v, want false", got)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp structs.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status {
		t.Errorf("resp.Status = true, want false")
	}
}

func TestCheckToken_MalformedHeader(t *testing.T) {
	c, w := newTestContext(http.MethodGet, "/", nil, "")
	c.Request.Header.Set("Authorization", "no-bearer-prefix")

	if got := CheckToken(c); got != false {
		t.Fatalf("CheckToken() = %v, want false", got)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCheckToken_DatabaseError(t *testing.T) {
	mock := mockDB(t)
	mock.ExpectQuery(`select now\(\);`).WillReturnError(errors.New("connection lost"))

	c, w := newTestContext(http.MethodGet, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	if got := CheckToken(c); got != false {
		t.Fatalf("CheckToken() = %v, want false", got)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCheckToken_InvalidToken(t *testing.T) {
	mock := mockDB(t)
	expectInvalidToken(mock)

	c, w := newTestContext(http.MethodGet, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	if got := CheckToken(c); got != false {
		t.Fatalf("CheckToken() = %v, want false", got)
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCheckToken_ValidToken(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	c, w := newTestContext(http.MethodGet, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	if got := CheckToken(c); got != true {
		t.Fatalf("CheckToken() = %v, want true", got)
	}

	// CheckToken itself writes no response body on success; that is left
	// to the caller.
	if w.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAccessTokenCheck(t *testing.T) {
	mock := mockDB(t)
	expectValidToken(mock)

	c, w := newTestContext(http.MethodGet, "/", nil, "")
	c.Request.Header.Set("Authorization", "Bearer sometoken")

	AccessTokenCheck(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp structs.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Errorf("resp.Status = false, want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAccessToken_MissingSecretKey(t *testing.T) {
	c, w := newTestContext(http.MethodPost, "/", strings.NewReader(""), "application/x-www-form-urlencoded")

	AccessToken(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAccessToken_InvalidSecretKey(t *testing.T) {
	enviroment.ENV.SECRET_KEY = "correct-secret"

	form := url.Values{"secret_key": {"wrong-secret"}}
	c, w := newTestContext(http.MethodPost, "/", strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")

	AccessToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAccessToken_Success(t *testing.T) {
	enviroment.ENV.SECRET_KEY = "correct-secret"
	enviroment.ENV.MAX_SESSION_KEYS_AMOUNT = 10
	enviroment.ENV.ACCESS_TOKEN_LIFETIME_MINUTES = 60

	mock := mockDB(t)

	// GenerateSessionKey -> CheckSessionKeyUnique: no existing rows means unique.
	mock.ExpectQuery(`select \* \s*from auth_session_keys`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_key", "creation_datetime", "valid_until"}))

	// GenerateSessionKey -> SaveSessionKey: current timestamp, cleanup, insert.
	mock.ExpectQuery(`select now\(\);`).
		WillReturnRows(sqlmock.NewRows([]string{"now"}).AddRow(time.Now()))
	mock.ExpectExec(`delete from auth_session_keys`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`insert into auth_session_keys`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"secret_key": {"correct-secret"}}
	c, w := newTestContext(http.MethodPost, "/", strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")

	AccessToken(c)

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
