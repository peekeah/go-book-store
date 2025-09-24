package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Book struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Author          string `json:"author"`
	PublishedYear   int    `json:"published_year"`
	AvailableCopies int    `json:"available_copies"`
}

type creatBook struct {
	Name            string `json:"name"`
	Author          string `json:"author"`
	PublishedYear   int    `json:"published_year"`
	AvailableCopies int    `json:"available_copies"`
}

type BookStore struct {
	books []Book
}

func NewBookStore() *BookStore {
	return &BookStore{}
}

func (bs *BookStore) GetBooks(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, &bs.books)
}

func (bs *BookStore) GetBookById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	for _, book := range bs.books {
		if book.Id == bookId {
			respondJSON(w, http.StatusOK, book)
			return
		}
	}
	respondError(w, http.StatusBadRequest, "book not found")
}

func (bs *BookStore) CreateBook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body creatBook

	byte, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = json.Unmarshal(byte, &body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newBook := Book{
		uuid.NewString(),
		body.Name,
		body.Author,
		body.PublishedYear,
		body.AvailableCopies,
	}

	bs.books = append(bs.books, newBook)
	respondJSON(w, http.StatusCreated, newBook)
}

func (bs *BookStore) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = json.Unmarshal(bytes, &book)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for id, crrBook := range bs.books {
		if crrBook.Id == book.Id {
			bs.books[id] = book
			respondJSON(w, http.StatusOK, book)
			return
		}
	}
	respondError(w, http.StatusForbidden, "book not found")
}

func (bs *BookStore) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url params not passed")
		return
	}

	for id, book := range bs.books {
		if book.Id == bookId {
			bs.books = append(bs.books[:id], bs.books[id+1:]...)
			respondJSON(w, http.StatusOK, "successfully deleted book")
			return
		}
	}
	respondError(w, http.StatusBadRequest, "book not found")
}

func (bs *BookStore) PurchaseBook(w http.ResponseWriter, r *http.Request, us *UserStore) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	userId := r.Context().Value("user_id")

	var body struct {
		BookId string `json:"book_id"`
	}

	err = json.Unmarshal(bytes, &body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// validate user
	var existUser bool
	var user User
	for _, crrUser := range us.users {
		if crrUser.Id == userId {
			existUser = true
			crrUser = &user
			break
		}
	}

	if !existUser {
		respondError(w, http.StatusForbidden, "user not found")
		return
	}

	for id, book := range bs.books {
		if book.Id == body.BookId {
			if book.AvailableCopies != 0 {
				bs.books[id].AvailableCopies -= 1

				// update user
				user.Purchase = append(user.Purchase, &book)
				respondJSON(w, http.StatusOK, bs.books[id])
				return
			}
			respondError(w, http.StatusForbidden, "book out of stock")
			return
		}
	}
	respondError(w, http.StatusBadRequest, "book not found")
}
