package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name      string `json:"name" validate:"required"`
	City      string `json:"city,omitempty"`
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password,omitempty" validate:"required"`
	Role      string `json:"role"`
	Purchases []Purchase
}

type UpdateUserPayload struct {
	ID    uint   `json:"ID,omitempty"`
	Name  string `json:"name,omitempty"`
	City  string `json:"city,omitempty"`
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}
