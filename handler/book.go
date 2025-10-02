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
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, books, ""}
	res.Dispatch()
}

func GetBookById(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId, ok := vars["id"]

	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid book id"}
		res.Dispatch()
		return
	}

	bookIdInt, err := strconv.Atoi(bookId)
	if err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid book id"}
		res.Dispatch()
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookIdInt).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, "book not found"}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, book, ""}
	res.Dispatch()
}

func CreateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	book := model.Book{}

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	defer r.Body.Close()

	// validate payload
	if err := validate.Struct(&book); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	if err := db.Save(&book).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusCreated, book, ""}
	res.Dispatch()
}

func UpdateBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "url params not passed"}
		res.Dispatch()
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid url params"}
		res.Dispatch()
		return
	}

	book := model.UpdateBook{}
	book.ID = uint(bookId)

	if err := validate.Struct(&book); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	defer r.Body.Close()

	dbBook := model.Book{}

	if err := db.First(&dbBook, book.ID).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, err.Error()}
		res.Dispatch()
		return
	}

	if dbBook.ID == 0 {
		res := ErrorResponse{w, http.StatusNotFound, "book does not exist"}
		res.Dispatch()
		return
	}

	if err := db.Model(&dbBook).Updates(&book).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, dbBook, ""}
	res.Dispatch()
}

func DeleteBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIdStr, ok := vars["id"]

	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "url params not passed"}
		res.Dispatch()
		return
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	book := model.Book{}

	if err := db.First(&book, bookId).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, "book not found"}
		res.Dispatch()
		return
	}

	if err := db.Delete(&book).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, book, ""}
	res.Dispatch()
}

func PurchaseBook(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	payload := model.PurchasePayload{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
	}

	defer r.Body.Close()

	if err := validate.Struct(&payload); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	userId := r.Context().Value("user_id").(uint)

	// Transaction
	tx := db.Begin()

	/*
		if err := tx.Error; err != nil {
			tx.Rollback()
			res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
			res.Dispatch()
			return
		}
	*/

	user := model.User{}
	book := model.Book{}

	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusBadRequest, "user not found"}
		res.Dispatch()
		return
	}

	if user.ID == 0 {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusBadRequest, "user not found"}
		res.Dispatch()
		return
	}

	if err := tx.First(&book, payload.BookId).Error; err != nil {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusBadRequest, "book not found"}
		res.Dispatch()
		return
	}

	if book.ID == 0 {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusBadRequest, "book not found"}
		res.Dispatch()
		return
	}

	if (book.AvailableCopies - payload.Quantity) <= 0 {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusBadRequest, fmt.Sprintf("only %d stock available, can not purchase %d quantities", book.AvailableCopies, payload.Quantity)}
		res.Dispatch()
		return
	}

	book.AvailableCopies = (book.AvailableCopies - payload.Quantity) - 1

	if err := tx.Save(&book).Error; err != nil {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	// purchase
	purchase := model.Purchase{
		UserID:   userId,
		BookID:   uint(payload.BookId),
		Quantity: payload.Quantity,
		Amount:   payload.Quantity * book.Price,
	}

	if err := db.Save(&purchase).Error; err != nil {
		tx.Rollback()
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, nil, "successfully purchased book"}
	res.Dispatch()
}
