package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name"`
	City     string `json:"city"`
	Email    string `json:"email"`
	Password string `json:"password"`

	Purchase []Book `gorm:"many2many:user_books" json:"purchase"`
}
