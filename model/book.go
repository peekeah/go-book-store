package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model

	Name            string `json:"name" validate:"required"`
	Author          string `json:"author" validate:"required"`
	PublishedYear   int    `json:"published_year" validate:"required"`
	AvailableCopies int    `json:"available_copies" validate:"required"`
	Price           int    `json:"price" validate:"required"`
	Purchases       []Purchase
}

type UpdateBook struct {
	ID              uint   `json:"ID,omitempty"`
	Name            string `json:"name,omitempty"`
	Author          string `json:"author,omitempty"`
	PublishedYear   int    `json:"published_year,omitempty"`
	AvailableCopies int    `json:"available_copies,omitempty"`
}
