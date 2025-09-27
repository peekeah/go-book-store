package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
)

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

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
