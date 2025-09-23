package handler

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Purchase []*Book `json:"purchase"`
}

type UserStore struct {
	users []User
}

func InitilizeUserStore() *UserStore {
	return &UserStore{}
}

func (us *UserStore) GetUsers() []User {
	return us.users
}

func (us *UserStore) GetUserById(id string) (User, error) {
	for _, user := range us.users {
		if user.Id == id {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (us *UserStore) CreateUser(name string, city string) {
	us.users = append(us.users, User{
		Id:   uuid.NewString(),
		Name: name,
		City: city,
	})
}

func (us *UserStore) UpdateUser(user User) (User, error) {
	for id, crruser := range us.users {
		if crruser.Id == user.Id {
			us.users[id] = user
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (us *UserStore) DeleteUser(user User) error {
	for id, crruser := range us.users {
		if crruser.Id == user.Id {
			us.users = append(us.users[:id], us.users[id+1:]...)
			return nil
		}
	}
	return errors.New("user not found")
}
