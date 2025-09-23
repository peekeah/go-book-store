package handler

import (
	"errors"

	"github.com/google/uuid"
)

type Book struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Author          string `json:"author"`
	PublishedYear   int    `json:"published_year"`
	AvailableCopies int    `json:"available_copies"`
}

type BookStore struct {
	books []Book
}

func InitilizeBookStore() *BookStore {
	return &BookStore{}
}

func (bs *BookStore) GetBooks() []Book {
	return bs.books
}

func (bs *BookStore) GetBookById(id string) (Book, error) {
	for _, book := range bs.books {
		if book.Id == id {
			return book, nil
		}
	}
	return Book{}, errors.New("book not found")
}

func (bs *BookStore) CreateBook(bookName string, author string, published_year int) {
	bs.books = append(bs.books, Book{
		Id:            uuid.NewString(),
		Name:          bookName,
		Author:        author,
		PublishedYear: published_year,
	})
}

func (bs *BookStore) UpdateBook(book Book) (Book, error) {
	for id, crrBook := range bs.books {
		if crrBook.Id == book.Id {
			bs.books[id] = book
			return book, nil
		}
	}
	return Book{}, errors.New("book not found")
}

func (bs *BookStore) DeleteBook(book Book) error {
	for id, crrBook := range bs.books {
		if crrBook.Id == book.Id {
			bs.books = append(bs.books[:id], bs.books[id+1:]...)
			return nil
		}
	}
	return errors.New("book not found")
}

func (bs *BookStore) PurchaseBook(bookId string, user *User) (Book, error) {
	for id, book := range bs.books {
		if book.Id == bookId {
			if book.AvailableCopies != 0 {
				bs.books[id].AvailableCopies -= 1
				// update user
				user.Purchase = append(user.Purchase, &book)
				return book, nil
			}
			return Book{}, errors.New("book out of stock")
		}
	}
	return Book{}, errors.New("book not found")
}
