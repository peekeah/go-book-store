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

func (bs *UserStore) GetUsers() []User {
	return bs.users
}

func (bs *UserStore) GetUserById(id string) (User, error) {
	for _, user := range bs.users {
		if user.Id == id {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (bs *UserStore) CreateUser(name string, city string) {
	bs.users = append(bs.users, User{
		Id:   uuid.NewString(),
		Name: name,
		City: city,
	})
}

func (bs *UserStore) Updateuser(user User) (User, error) {
	for id, crruser := range bs.users {
		if crruser.Id == user.Id {
			bs.users[id] = user
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (bs *UserStore) Deleteuser(user User) error {
	for id, crruser := range bs.users {
		if crruser.Id == user.Id {
			bs.users = append(bs.users[:id], bs.users[id+1:]...)
			return nil
		}
	}
	return errors.New("user not found")
}
