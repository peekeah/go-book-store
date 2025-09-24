package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Name            string `json:"name"`
	Author          string `json:"author"`
	PublishedYear   int    `json:"published_year"`
	AvailableCopies int    `json:"available_copies"`

	// foreign key
	PurchasedCustomers []User `gorm:"many2many:user_books;" json"purchased_customers"`
}
