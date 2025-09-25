package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
	"gorm.io/gorm"
)

var validate = validator.New()

func GetUsers(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	users := []model.User{}

	if err := db.Omit("password").Find(&users).Error; err != nil {
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

	if err := db.Omit("password").First(&user, id).Error; err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	payload := model.User{}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
	}

	defer r.Body.Close()

	if validationErr := validate.Struct(&payload); validationErr != nil {
		respondError(w, http.StatusBadRequest, validationErr.Error())
		return
	}

	hashedPwd, err := utils.HashPassword(payload.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
	}

	payload.Password = hashedPwd

	if err := db.Save(&payload).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, payload)
}

func UpdateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	payload := model.UpdateUserPayload{}
	payload.ID = uint(userIdInt)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if validationErr := validate.Struct(&payload); validationErr != nil {
		respondError(w, http.StatusBadRequest, validationErr.Error())
		return
	}

	dbUser := model.User{}

	if err := db.Omit("password").First(&dbUser, payload.ID).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if dbUser.ID == 0 {
		respondError(w, http.StatusNotFound, "user does not exist")
		return
	}

	if err := db.Model(&dbUser).Updates(payload).Error; err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dbUser)
}

func DeleteUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url param not found")
		return
	}

	user := model.User{}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := db.First(&user, userIdInt).Error; err != nil {
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

	type payload struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	var body payload

	user := model.User{}

	if err := decoder.Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// payload validation
	if validationErr := validate.Struct(&body); validationErr != nil {
		respondError(w, http.StatusBadRequest, validationErr.Error())
		return
	}

	// check user id db
	if err := db.First(&user, model.User{Email: body.Email}).Error; err != nil {
		respondError(w, http.StatusNotFound, "user does not exist")
		return
	}

	// Validate password
	if !utils.ComparePassword(body.Password, user.Password) {
		respondError(w, http.StatusNotFound, "password does not match")
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
