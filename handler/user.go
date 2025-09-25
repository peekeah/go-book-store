package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
	"gorm.io/gorm"
)

func GetUsers(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	users := model.User{}

	if err := db.Model(&users).Error; err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func GetUserById(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	user := model.User{}

	if err := db.First(&user, id); err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	user := model.User{}
	if err := decoder.Decode(&user); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
	}

	hashedPwd, err := utils.HashPassword(user.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
	}

	user.Password = hashedPwd

	if err := db.Save(&user).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func UpdateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	user := model.User{}

	if err := decoder.Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbUser := model.User{}

	if err := db.First(&db, dbUser.ID).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if dbUser.ID == 0 {
		respondError(w, http.StatusNotFound, "user does not exist")
		return
	}

	if err := db.Save(&user).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func DeleteUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url param not found")
		return
	}

	user := model.User{}

	if err := db.First(&user, userId); err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.ID == 0 {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := db.Delete(&user, userId).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func UserLogin(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	body := struct {
		Email    string
		Password string
	}{}

	user := model.User{}

	if err := decoder.Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// check user id db
	if err := db.First(&user).Error; err != nil {
		respondError(w, http.StatusNotFound, "user does not exist")
		return
	}

	token, err := utils.CreateJWTToken(utils.JWTTokenBody{ID: user.ID, Email: user.Email, Name: user.Name})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{token})
}
