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
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, users, ""}
	res.Dispatch()
}

func GetUserById(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "id is required"}
		res.Dispatch()
		return
	}

	user := model.User{}

	if err := db.Omit("password").First(&user, id).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, "user not found"}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, user, ""}
	res.Dispatch()
}

func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	payload := model.User{}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	defer r.Body.Close()

	if validationErr := validate.Struct(&payload); validationErr != nil {
		res := ErrorResponse{w, http.StatusBadRequest, validationErr.Error()}
		res.Dispatch()
		return
	}

	hashedPwd, err := utils.HashPassword(payload.Password)
	if err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	existUser := model.User{}
	existUser.Email = payload.Email

	if err := db.First(&existUser, model.User{Email: payload.Email}).Error; err == nil {
		res := ErrorResponse{w, http.StatusNotFound, "user already exist"}
		res.Dispatch()
		return
	}

	payload.Password = hashedPwd
	payload.Role = "user"

	// default role user
	if r.Context().Value("role") == "admin" {
		payload.Role = "admin"
	}

	if err := db.Save(&payload).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusCreated, payload, ""}
	res.Dispatch()
}

func UpdateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid user id"}
		res.Dispatch()
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid user id"}
		res.Dispatch()
		return
	}

	payload := model.UpdateUserPayload{}
	payload.ID = uint(userIdInt)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	defer r.Body.Close()

	if validationErr := validate.Struct(&payload); validationErr != nil {
		res := ErrorResponse{w, http.StatusBadRequest, validationErr.Error()}
		res.Dispatch()
		return
	}

	dbUser := model.User{}

	if err := db.Omit("password").First(&dbUser, payload.ID).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	if dbUser.ID == 0 {
		res := ErrorResponse{w, http.StatusNotFound, "user does not exist"}
		res.Dispatch()
		return
	}

	if err := db.Model(&dbUser).Updates(payload).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, dbUser, ""}
	res.Dispatch()
}

func DeleteUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		res := ErrorResponse{w, http.StatusBadRequest, "url param not found"}
		res.Dispatch()
		return
	}

	user := model.User{}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		res := ErrorResponse{w, http.StatusBadRequest, "invalid user id"}
		res.Dispatch()
		return
	}

	if err := db.First(&user, userIdInt).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, "user not found"}
		res.Dispatch()
		return
	}

	if user.ID == 0 {
		res := ErrorResponse{w, http.StatusNotFound, "user not found"}
		res.Dispatch()
		return
	}

	if err := db.Delete(&user, userId).Error; err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, user, ""}
	res.Dispatch()
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
		res := ErrorResponse{w, http.StatusBadRequest, err.Error()}
		res.Dispatch()
		return
	}

	// payload validation
	if validationErr := validate.Struct(&body); validationErr != nil {
		res := ErrorResponse{w, http.StatusBadRequest, validationErr.Error()}
		res.Dispatch()
		return
	}

	// check user id db
	if err := db.First(&user, model.User{Email: body.Email}).Error; err != nil {
		res := ErrorResponse{w, http.StatusNotFound, "user does not exist"}
		res.Dispatch()
		return
	}

	// Validate password
	if !utils.ComparePassword(body.Password, user.Password) {
		res := ErrorResponse{w, http.StatusNotFound, "password does not match"}
		res.Dispatch()
		return
	}

	token, err := utils.CreateJWTToken(utils.JWTTokenBody{ID: user.ID, Email: user.Email, Name: user.Name})
	if err != nil {
		res := ErrorResponse{w, http.StatusInternalServerError, err.Error()}
		res.Dispatch()
		return
	}

	res := SuccessResponse{w, http.StatusOK, struct {
		Token string `json:"token"`
	}{token}, ""}

	res.Dispatch()
}
