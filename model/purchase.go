package model

import (
	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model

	UserID   uint `json:"user_id" gorm:"index;not null"`
	BookID   uint `json:"book_id" gorm:"index;not null"`
	Quantity int  `json:"quantity"`
	Amount   int  `json:"amount" gorm:"not null"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Book Book `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE;"`
}

type PurchasePayload struct {
	BookId   int `json:"book_id" validate:"required"`
	Quantity int `json:"quantity" validate:"required"`
}
