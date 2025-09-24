package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"gorm.io/gorm"
)

func GetBooks(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	books := []model.Book{}

	if err := db.Model(&books); err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, books)
}

func GetBookById(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookId); err != nil {
		respondError(w, http.StatusNotFound, "book not found")
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func CreateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	book := model.Book{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&book); err != nil {
		fmt.Println("hre:", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err := db.Save(&book).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, book)
}

func UpdateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url params not passed")
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid url params")
		return
	}

	book := model.Book{}
	book.ID = uint(bookId)

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	if err := decoder.Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbBook := model.Book{}

	if err := db.First(&db, book.ID); err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if dbBook.ID == 0 {
		respondError(w, http.StatusNotFound, "book does not exist")
		return
	}

	if err := db.Save(&book).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func DeleteBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url params not passed")
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookId); err != nil {
		respondError(w, http.StatusNotFound, "book not found")
		return
	}

	if err := db.Delete(&book); err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func PurchaseBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "book id required")
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
	}

	userId := r.Context().Value("user_id")

	err = db.Transaction(func(tx *gorm.DB) error {
		// serial db calls
		book := model.Book{}
		user := model.User{}

		if err := tx.First(&user, userId); err != nil {
			respondError(w, http.StatusBadRequest, "user not found")
			return errors.New("")
		}

		if err := tx.First(&book, bookId); err != nil {
			respondError(w, http.StatusBadRequest, "book not found")
			return errors.New("")
		}

		if book.AvailableCopies == 0 {
			respondError(w, http.StatusBadRequest, "book out of stock")
			return errors.New("")
		}

		book.AvailableCopies -= book.AvailableCopies
		book.PurchasedCustomers = append(book.PurchasedCustomers, user)
		if err := tx.Save(&book); err != nil {
			respondError(w, http.StatusInternalServerError, "intenal server error")
			return errors.New("")
		}

		user.Purchase = append(user.Purchase, book)
		if err := tx.Save(&user); err != nil {
			respondError(w, http.StatusInternalServerError, "internal server error")
			return errors.New("")
		}

		return nil
	})
	if err != nil {
		respondJSON(w, http.StatusOK, struct{ message string }{"successfully purchased book"})
	}
}
