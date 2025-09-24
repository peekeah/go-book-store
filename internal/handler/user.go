package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/utils"
)

type User struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Purchase []*Book `json:"purchase"`
}

type createUser struct {
	Name     string `json:"name"`
	City     string `json:"city"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserStore struct {
	users []*User
}

func InitilizeUserStore() *UserStore {
	return &UserStore{}
}

func (us *UserStore) GetUsers(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, us.users)
}

func (us *UserStore) GetUserById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		respondError(w, http.StatusForbidden, "user not found")
		return
	}

	for _, user := range us.users {
		if user.Id == id {
			respondJSON(w, http.StatusOK, *user)
			return
		}
	}
	respondError(w, http.StatusForbidden, "user not found")
}

func (us *UserStore) CreateUser(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var userBody createUser

	if err = json.Unmarshal(bytes, &userBody); err != nil {
		respondError(w, http.StatusBadRequest, "incorrect body")
		return
	}

	hashedPwd, err := utils.HashPassword(userBody.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	newUser := &User{
		Id:       uuid.NewString(),
		Name:     userBody.Name,
		City:     userBody.City,
		Email:    userBody.Email,
		Password: hashedPwd,
	}

	us.users = append(us.users, newUser)

	respondJSON(w, http.StatusCreated, newUser)
}

func (us *UserStore) UpdateUser(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var user User

	err = json.Unmarshal(bytes, &user)
	if err != nil {
		respondError(w, http.StatusBadRequest, "incorrect body")
		return
	}

	for id, crrUser := range us.users {
		if crrUser.Id == user.Id {
			us.users[id] = &user
			respondJSON(w, http.StatusOK, user)
			return
		}
	}
	respondError(w, http.StatusForbidden, "user not found")
}

func (us *UserStore) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, ok := vars["id"]

	if !ok {
		respondError(w, http.StatusBadRequest, "url param not found")
		return
	}

	for id, crruser := range us.users {
		if crruser.Id == userId {
			us.users = append(us.users[:id], us.users[id+1:]...)
			respondJSON(w, http.StatusOK, "successfully deleted user")
			return
		}
	}
	respondError(w, http.StatusForbidden, "user not found")
}

func (us *UserStore) UserLogin(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var body struct {
		Email    string
		Password string
	}

	if err = json.Unmarshal(bytes, &body); err != nil {
		respondError(w, http.StatusForbidden, "incorrect payload")
		return
	}

	for id := range us.users {
		if us.users[id].Email == body.Email {
			if !utils.ComparePassword(body.Password, us.users[id].Password) {
				respondError(w, http.StatusUnauthorized, "password does not match")
				return
			}

			token, err := utils.CreateJWTToken(struct {
				Id    string
				Email string
				Name  string
			}{
				us.users[id].Id,
				us.users[id].Name,
				us.users[id].Email,
			})
			if err != nil {
				respondError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			respondJSON(w, http.StatusOK, struct {
				Token string `json:"token"`
			}{token})
			return
		}
	}
	respondError(w, http.StatusForbidden, "user not found")
}
