package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `json:"name" validate:"required"`
	City     string `json:"city,omitempty" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`

	Purchase []Book `gorm:"many2many:user_books" json:"purchase"`
}

type UpdateUserPayload struct {
	ID    uint   `json:"ID,omitempty"`
	Name  string `json:"name,omitempty"`
	City  string `json:"city,omitempty"`
	Email string `json:"email,omitempty"`
}
