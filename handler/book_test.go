package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
)

func TestGetBooks(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1").
		AddRow(2, "Book2", "Author2")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").WillReturnRows(mockRows)

	req, err := http.NewRequest(http.MethodGet, "/books/", nil)
	if err != nil {
		t.Fatalf("error while making request")
	}

	w := httptest.NewRecorder()

	// Handler call
	GetBooks(db, w, req)

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

	if res.Data[0].Name != "Book1" {
		t.Fatalf("expeced Book1, got %s", res.Data[0].Name)
	}
}

func TestGetBookById(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").WillReturnRows(mockRows)

	req, err := http.NewRequest(http.MethodGet, "/books/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	if err != nil {
		t.Fatalf("error while making request")
	}

	w := httptest.NewRecorder()

	GetBookById(db, w, req)

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
		t.Fatalf("wrong book")
	}
}

func TestCreateBook(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":             "Book1",
		"author":           "Author2",
		"available_copies": 12,
		"published_year":   1999,
	}

	payloadByte, err := json.Marshal(payload)
	if err != nil {
		t.Error("error while parsing payload")
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`^INSERT INTO "books"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"Book1",
			"Author2",
			1999,
			12,
		).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(payloadByte))
	if err != nil {
		t.Fatalf("error while making request")
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	CreateBook(db, w, req)

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
		t.Fatalf("book creation failed")
	}
}

func TestUpdateBook(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WithArgs(1, 1).
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"books\" SET").
		WithArgs(1, "Book2", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payload := struct {
		ID     int    `json:"ID"`
		Name   string `json:"name"`
		Author string `json:"author"`
	}{
		ID:     1,
		Name:   "Book2",
		Author: "Author2",
	}

	bytesData, err := json.Marshal(&payload)
	if err != nil {
		t.Fatalf("error while encoding req")
	}

	req, err := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	if err != nil {
		t.Fatalf("error while making request")
	}

	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

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
		t.Fatalf("book update failed")
	}
}

func TestDeleteBook(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WithArgs(1, 1).
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"books\" SET \"deleted_at\"").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodDelete, "/books/1", nil)
	if err != nil {
		t.Fatalf("error while making req")
	}

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	w := httptest.NewRecorder()

	DeleteBook(db, w, req)

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
		t.Fatalf("Book delete fail")
	}
}

func TestGetBookById_Errors(t *testing.T) {
	db, _ := utils.GetDBMock()

	// missing id
	req, _ := http.NewRequest(http.MethodGet, "/books/", nil)
	w := httptest.NewRecorder()

	GetBookById(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}

	// invalid id
	req, _ = http.NewRequest(http.MethodGet, "/books/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w = httptest.NewRecorder()

	GetBookById(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}

	// not found
	req, _ = http.NewRequest(http.MethodGet, "/books/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w = httptest.NewRecorder()

	GetBookById(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}
