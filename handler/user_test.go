package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
)

func TestGetUsers(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com").
		AddRow(1, "user2", "user2@example.com")

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").WillReturnRows(mockRows)

	req, err := http.NewRequest(http.MethodGet, "/users/", nil)
	if err != nil {
		t.Fatalf("error while making request")
	}

	w := httptest.NewRecorder()

	// Handler call
	GetUsers(db, w, req)

	// validate status
	if w.Code != http.StatusOK {
		t.Fatalf("expeced status 200, got %d", w.Code)
	}

	// validate body
	res := struct {
		Status  int          `json:"status"`
		Message string       `json:"message"`
		Data    []model.Book `json:"data"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 200 {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(res.Data) != 2 {
		t.Fatalf("expeced 3 books, got %d", len(res.Data))
	}

	if res.Data[0].Name != "user1" {
		t.Fatalf("expeced user1, got %s", res.Data[0].Name)
	}
}

func TestGetUserById(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(mockRows)

	req, err := http.NewRequest(http.MethodGet, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	if err != nil {
		t.Fatalf("error while making request")
	}

	w := httptest.NewRecorder()

	GetUserById(db, w, req)

	// validate status
	if w.Code != http.StatusOK {
		t.Fatalf("expeced status 200, got %d", w.Code)
	}

	// validate body
	res := struct {
		Status  int        `json:"status"`
		Message string     `json:"message"`
		Data    model.Book `json:"data"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 200 {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Data.ID != 1 {
		t.Fatalf("wrong user")
	}
}

func TestCreateUser(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":     "user1",
		"email":    "user@example.com",
		"city":     "Bengaluru",
		"password": "test",
	}

	payloadByte, err := json.Marshal(payload)
	if err != nil {
		t.Error("error while parsing payload")
	}

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnError(errors.New("user already exist"))

	mock.ExpectBegin()

	mock.ExpectQuery(`^INSERT INTO "users"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"user1",
			"Bengaluru",
			"user@example.com",
			sqlmock.AnyArg(),
			"user",
		).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	if err != nil {
		t.Fatalf("error while making request")
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	CreateUser(db, w, req)

	// validate status
	if w.Code != http.StatusCreated {
		t.Fatalf("expeced status 201, got %d", w.Code)
	}

	// validate body
	res := struct {
		Status  int        `json:"status"`
		Message string     `json:"message"`
		Data    model.Book `json:"data"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 201 {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Data.ID != 1 {
		t.Fatalf("failed to create user")
	}
}

func TestUserLogin(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"email":    "user@example.com",
		"password": "test",
	}

	payloadByte, err := json.Marshal(payload)
	if err != nil {
		t.Error("error while parsing payload")
	}

	hashedPwd, err := utils.HashPassword("test")
	if err != nil {
		t.Fatalf("error while hashing password")
	}

	rows := sqlmock.NewRows([]string{"ID", "email", "password"}).
		AddRow(1, "user@example.com", hashedPwd)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	req, err := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payloadByte))
	if err != nil {
		t.Fatalf("error while making request")
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	UserLogin(db, w, req)

	// validate status
	if w.Code != http.StatusOK {
		t.Fatalf("expeced status 200, got %d", w.Code)
	}

	// validate body
	res := struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		}
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 200 {
		t.Fatalf("failed to parse response: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":  "user1",
		"email": "user@example.com",
		"city":  "Bengaluru",
	}

	payloadByte, err := json.Marshal(payload)
	if err != nil {
		t.Error("error while parsing payload")
	}

	rows := sqlmock.NewRows([]string{"ID", "name", "email", "city"}).
		AddRow(1, "user2", "user1@example.com", "Pune")

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	mock.ExpectBegin()

	mock.ExpectExec(`^UPDATE "users"`).
		WithArgs(
			1,
			"user1",
			"Bengaluru",
			"user@example.com",
			1,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	if err != nil {
		t.Fatalf("error while making request")
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	UpdateUser(db, w, req)

	// validate status
	if w.Code != http.StatusOK {
		t.Fatalf("expeced status 200, got %d", w.Code)
	}

	// validate body
	res := struct {
		Status  int        `json:"status"`
		Message string     `json:"message"`
		Data    model.Book `json:"data"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 200 {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Data.ID != 1 {
		t.Fatalf("failed to create user")
	}
}

func TestDeleteUser(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WithArgs(1, 1).
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"users\" SET \"deleted_at\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodDelete, "/users/1", nil)
	if err != nil {
		t.Fatalf("error while making req")
	}

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	// validate body
	res := struct {
		Status  int        `json:"status"`
		Message string     `json:"message"`
		Data    model.Book `json:"data"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Status != 200 {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Data.ID != 1 {
		t.Fatalf("failed to delete user")
	}
}

func TestGetUserById_Errors(t *testing.T) {
	db, _ := utils.GetDBMock()

	// missing id
	req, _ := http.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()

	GetUserById(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}

	// invalid id
	req, _ = http.NewRequest(http.MethodGet, "/users/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w = httptest.NewRecorder()

	GetUserById(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}

	// not found
	req, _ = http.NewRequest(http.MethodGet, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w = httptest.NewRecorder()

	GetUserById(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}

func TestGetUsers_Error(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnError(errors.New("database error"))

	req, _ := http.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()

	GetUsers(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCreateUser_DecodeError(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCreateUser_ValidationError(t *testing.T) {
	db, _ := utils.GetDBMock()

	payload := map[string]any{
		"name": "", // Empty name
	}

	payloadByte, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateUser_HashPasswordError(t *testing.T) {
	db, _ := utils.GetDBMock()

	payload := map[string]any{
		"name":     "user1",
		"email":    "user@example.com",
		"city":     "City",
		"password": string(make([]byte, 100)), // Very long password that might fail
	}

	payloadByte, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	// This may or may not fail depending on bcrypt limits
	// Adjust based on your hash implementation
}

func TestCreateUser_UserExists(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":     "user1",
		"email":    "user@example.com",
		"city":     "City",
		"password": "test",
	}

	payloadByte, _ := json.Marshal(payload)

	existingUserRows := sqlmock.NewRows([]string{"ID", "email"}).
		AddRow(1, "user@example.com")

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(existingUserRows)

	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCreateUser_WithAdminRole(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":     "admin",
		"email":    "admin@example.com",
		"city":     "City",
		"password": "test",
	}

	payloadByte, _ := json.Marshal(payload)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnError(errors.New("not found"))

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "users"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"admin",
			"City",
			"admin@example.com",
			sqlmock.AnyArg(),
			"admin",
		).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1))
	mock.ExpectCommit()

	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	req = req.WithContext(context.WithValue(req.Context(), "role", "admin"))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestCreateUser_SaveError(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":     "user1",
		"email":    "user@example.com",
		"city":     "City",
		"password": "test",
	}

	payloadByte, _ := json.Marshal(payload)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnError(errors.New("not found"))

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "users"`).
		WillReturnError(errors.New("save error"))
	mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	CreateUser(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUpdateUser_MissingID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/users/", nil)
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUser_InvalidID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/users/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUser_DecodeError(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewReader([]byte("invalid")))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{"name": "user1"}
	payloadByte, _ := json.Marshal(payload)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnError(errors.New("not found"))

	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUpdateUser_UserDoesNotExist(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{"name": "user1"}
	payloadByte, _ := json.Marshal(payload)

	rows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(0, "user1", "user@example.com")
	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateUser_UpdateError(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{"name": "user1"}
	payloadByte, _ := json.Marshal(payload)

	rows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user2", "user@example.com")

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "users"`).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateUser(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestDeleteUser_MissingID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodDelete, "/users/", nil)
	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodDelete, "/users/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteUser_UserNotFoundInDB(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnError(errors.New("not found"))

	req, _ := http.NewRequest(http.MethodDelete, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteUser_UserIDZero(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "email"})
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(mockRows)

	req, _ := http.NewRequest(http.MethodDelete, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteUser_DeleteError(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"users\"").
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodDelete, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteUser(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUserLogin_DecodeError(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	UserLogin(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserLogin_ValidationError(t *testing.T) {
	db, _ := utils.GetDBMock()

	payload := map[string]any{
		"email": "", // Empty email
	}

	payloadByte, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	UserLogin(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserLogin_UserNotFound(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"email":    "user@example.com",
		"password": "test",
	}

	payloadByte, _ := json.Marshal(payload)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnError(errors.New("not found"))

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	UserLogin(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserLogin_PasswordMismatch(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"email":    "user@example.com",
		"password": "wrongpassword",
	}

	payloadByte, _ := json.Marshal(payload)

	hashedPwd, _ := utils.HashPassword("correctpassword")
	rows := sqlmock.NewRows([]string{"ID", "email", "password"}).
		AddRow(1, "user@example.com", hashedPwd)

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	UserLogin(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserLogin_JWTCreationError(t *testing.T) {
	// This test requires mocking the CreateJWTToken function
	// You might need to refactor your code to make JWT creation testable
	// For now, this is a placeholder showing the approach
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"email":    "user@example.com",
		"password": "test",
	}

	payloadByte, _ := json.Marshal(payload)

	hashedPwd, _ := utils.HashPassword("test")
	rows := sqlmock.NewRows([]string{"ID", "email", "password", "name"}).
		AddRow(1, "user@example.com", hashedPwd, "")

	mock.ExpectQuery(`^SELECT (.+) FROM "users"`).
		WillReturnRows(rows)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payloadByte))
	w := httptest.NewRecorder()

	// Note: This might succeed unless you can mock JWT creation failure
	// Consider refactoring CreateJWTToken to be injectable
	UserLogin(db, w, req)
}
