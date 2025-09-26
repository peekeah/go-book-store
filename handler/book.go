package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"gorm.io/gorm"
)

func GetBooks(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	books := []model.Book{}

	if err := db.Find(&books).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	RespondJSON(w, http.StatusOK, books)
}

func GetBookById(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId, ok := vars["id"]

	if !ok {
		RespondError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	bookIdInt, err := strconv.Atoi(bookId)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookIdInt).Error; err != nil {
		RespondError(w, http.StatusNotFound, "book not found")
		return
	}

	RespondJSON(w, http.StatusOK, book)
}

func CreateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	book := model.Book{}

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	// validate payload
	if err := validate.Struct(&book); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.Save(&book).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Printf("%+v", book)

	RespondJSON(w, http.StatusCreated, book)
}

func UpdateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		RespondError(w, http.StatusBadRequest, "url params not passed")
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid url params")
		return
	}

	book := model.UpdateBook{}
	book.ID = uint(bookId)

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	dbBook := model.Book{}

	if err := db.First(&dbBook, book.ID).Error; err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if dbBook.ID == 0 {
		RespondError(w, http.StatusNotFound, "book does not exist")
		return
	}

	if err := db.Model(&dbBook).Updates(&book).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, dbBook)
}

func DeleteBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		RespondError(w, http.StatusBadRequest, "url params not passed")
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookId).Error; err != nil {
		RespondError(w, http.StatusNotFound, "book not found")
		return
	}

	if err := db.Delete(&book).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	RespondJSON(w, http.StatusOK, book)
}

func PurchaseBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		RespondError(w, http.StatusBadRequest, "book id required")
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
	}

	userId := r.Context().Value("user_id")

	// Transaction
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
		}
		tx.Rollback()
	}()

	if err := tx.Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusInternalServerError, "intenal server error")
		return
	}

	user := model.User{}
	book := model.Book{}

	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusBadRequest, "user not found")
		return
	}

	if err := tx.First(&book, bookId).Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusBadRequest, "book not found")
		return
	}

	if book.AvailableCopies == 0 {
		tx.Rollback()
		RespondError(w, http.StatusBadRequest, "book out of stock")
		return
	}

	book.AvailableCopies = book.AvailableCopies - 1

	book.PurchasedCustomers = append(book.PurchasedCustomers, user)
	if err := tx.Save(&book).Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusInternalServerError, "intenal server error")
		return
	}

	user.Purchase = append(user.Purchase, book)
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "successfully purchased book"})
}
