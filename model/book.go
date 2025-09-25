package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Name            string `json:"name" validate:"required"`
	Author          string `json:"author" validate:"required"`
	PublishedYear   int    `json:"published_year" validate:"required"`
	AvailableCopies int    `json:"available_copies" validate:"required"`

	// foreign key
	PurchasedCustomers []User `gorm:"many2many:user_books;" json"purchased_customers"`
}

type UpdateBook struct {
	ID              uint   `json:"ID,omitempty"`
	Name            string `json:"name,omitempty"`
	Author          string `json:"author omitempty"`
	PublishedYear   int    `json:"published_year,omitempty"`
	AvailableCopies int    `json:"available_copies,omitempty"`
}
