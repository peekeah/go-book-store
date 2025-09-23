package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/internal/handler"
)

type Server struct {
	addr int
}

func InitilizeServer(addr int) *Server {
	return &Server{addr: addr}
}

func (s *Server) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from Go Book"))
	})

	// User Routes
	users := handler.InitilizeUserStore()
	router.HandleFunc("/users", users.GetUsers).Methods("GET")
	router.HandleFunc("/users", users.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", users.GetUserById).Methods("GET")
	router.HandleFunc("/users/{id}", users.UpdateUser).Methods("POST")
	router.HandleFunc("/users/{id}", users.DeleteUser).Methods("DELETE")

	// Book Routes
	books := handler.InitilizeBookStore()

	router.HandleFunc("/books/purchase", func(w http.ResponseWriter, r *http.Request) {
		books.PurchaseBook(w, r, users)
	}).Methods("POST")

	router.HandleFunc("/books", books.GetBooks).Methods("GET")
	router.HandleFunc("/books", books.CreateBook).Methods("POST")
	router.HandleFunc("/books/{id}", books.GetBookById).Methods("GET")
	router.HandleFunc("/books/{id}", books.UpdateBook).Methods("POST")
	router.HandleFunc("/books/{id}", books.DeleteBook).Methods("DELETE")

	// Run Server
	fmt.Println("server starting on port", s.addr)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.addr), router); err != nil {
		log.Fatal(err)
	}
}
