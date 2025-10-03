package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		"price":            200,
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
			200,
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

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author", "price"}).
		AddRow(1, "Book1", "Author1", 200)

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WithArgs(1, 1).
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"books\" SET").
		WithArgs(1, "Book2", "Author2", 1).
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

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author", "price"}).
		AddRow(1, "Book1", "Author1", 200)

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
	/*
		req, _ := http.NewRequest(http.MethodGet, "/books/", nil)

		w := httptest.NewRecorder()

			GetBookById(db, w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400")
			}
	*/

	// invalid id
	req, _ := http.NewRequest(http.MethodGet, "/books/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

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

func TestGetBooks_Error(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnError(errors.New("database error"))

	req, _ := http.NewRequest(http.MethodGet, "/books/", nil)
	w := httptest.NewRecorder()

	GetBooks(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCreateBook_DecodeError(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateBook_ValidationError(t *testing.T) {
	db, _ := utils.GetDBMock()

	payload := map[string]any{
		"name": "", // Empty name should fail validation
	}

	payloadByte, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(payloadByte))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateBook_SaveError(t *testing.T) {
	db, mock := utils.GetDBMock()

	payload := map[string]any{
		"name":             "Book1",
		"author":           "Author1",
		"available_copies": 10,
		"published_year":   2000,
		"price":            200,
	}

	payloadByte, _ := json.Marshal(payload)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "books"`).
		WillReturnError(errors.New("save error"))
	mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, "/books", bytes.NewReader(payloadByte))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUpdateBook_MissingID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/books/", nil)
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateBook_InvalidID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/books/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateBook_DecodeError(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodPut, "/books/1", bytes.NewReader([]byte("invalid")))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateBook_BookNotFound(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnError(errors.New("not found"))

	payload := map[string]any{"name": "Book1"}
	payloadByte, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateBook_BookDoesNotExist(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"})
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(mockRows)

	payload := map[string]any{"name": "Book1"}
	payloadByte, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateBook_UpdateError(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	payload := map[string]any{"name": "Book2"}

	payloadByte, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(payloadByte))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestDeleteBook_MissingID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodDelete, "/books/", nil)
	w := httptest.NewRecorder()

	DeleteBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteBook_InvalidID(t *testing.T) {
	db, _ := utils.GetDBMock()

	req, _ := http.NewRequest(http.MethodDelete, "/books/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	DeleteBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteBook_DeleteError(t *testing.T) {
	db, mock := utils.GetDBMock()

	mockRows := sqlmock.NewRows([]string{"ID", "name", "author"}).
		AddRow(1, "Book1", "Author1")

	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(mockRows)

	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodDelete, "/books/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// ============ PURCHASE BOOK TESTS ============

func TestPurchaseBook_InvalidPayload(t *testing.T) {
	db, _ := utils.GetDBMock()

	body := map[string]any{
		"book_id": "a",
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/purchase", bytes.NewReader(bytesData))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPurchaseBook_TransactionBeginError(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	body := map[string]any{
		"book_id":  1,
		"quantity": 2,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPurchaseBook_UserNotFound(t *testing.T) {
	db, mock := utils.GetDBMock()

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnError(errors.New("user not found"))
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 2,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)
	fmt.Println("body:", w.Body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPurchaseBook_BookNotFound(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnError(errors.New("book not found"))
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 2,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPurchaseBook_OutOfStock(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email", "available_copies"}).
		AddRow(1, "user1", "user@example.com", 3)

	bookRows := sqlmock.NewRows([]string{"ID", "name", "available_copies"}).
		AddRow(1, "Book1", 0)

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(bookRows)
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 10,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPurchaseBook_SaveBookError(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	bookRows := sqlmock.NewRows([]string{"ID", "name", "available_copies", "price"}).
		AddRow(1, "Book1", 5, 200)

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(bookRows)
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnError(errors.New("save error"))
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 3,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPurchaseBook_PurchaseError(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	bookRows := sqlmock.NewRows([]string{"ID", "name", "available_copies", "price"}).
		AddRow(1, "Book1", 5, 250)

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(bookRows)
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "purchases"`).
		WillReturnError(errors.New("save error"))
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 3,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPurchaseBook_CommitError(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	bookRows := sqlmock.NewRows([]string{"ID", "name", "available_copies", "price"}).
		AddRow(1, "Book1", 5, 200)

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(bookRows)
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "purchases"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			1,
			1,
			3,
			600,
		).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1))

	mock.ExpectCommit().
		WillReturnError(errors.New("commit error"))
	mock.ExpectRollback()

	body := map[string]any{
		"book_id":  1,
		"quantity": 3,
	}

	bytesData, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)
	fmt.Println("bb:", w.Body)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPurchaseBook_Success(t *testing.T) {
	db, mock := utils.GetDBMock()

	userRows := sqlmock.NewRows([]string{"ID", "name", "email"}).
		AddRow(1, "user1", "user@example.com")

	bookRows := sqlmock.NewRows([]string{"ID", "name", "available_copies", "price"}).
		AddRow(1, "Book1", 5, 200)

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT (.+) FROM \"users\"").
		WillReturnRows(userRows)
	mock.ExpectQuery("^SELECT (.+) FROM \"books\"").
		WillReturnRows(bookRows)
	mock.ExpectExec("^UPDATE \"books\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "purchases"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			1,
			1,
			3,
			600,
		).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1))
	mock.ExpectCommit()
	mock.ExpectCommit()

	body := map[string]any{
		"book_id":  1,
		"quantity": 3,
	}

	bytesData, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/books/1/purchase", bytes.NewReader(bytesData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint(1)))
	w := httptest.NewRecorder()

	PurchaseBook(db, w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
