package books

import "errors"

type Book struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Author        string `json:"author"`
	PublishedYear int    `json:"published_year"`
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

func (bs *BookStore) CreateBook(book Book) {
	bs.books = append(bs.books, book)
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
