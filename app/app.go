package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/handler"
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

	users := handler.NewUserStore()
	books := handler.NewBookStore()

	// Auth Routes
	authRoutes := router.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/login", users.UserLogin)
	authRoutes.HandleFunc("/signup", users.CreateUser).Methods("POST")

	// Authorized routes
	// User Routes
	userRoutes := router.PathPrefix("/users").Subrouter()
	userRoutes.Use(authMiddleware)
	userRoutes.HandleFunc("/", users.GetUsers).Methods("GET")
	userRoutes.HandleFunc("/{id}", users.GetUserById).Methods("GET")
	userRoutes.HandleFunc("/{id}", users.UpdateUser).Methods("POST")
	userRoutes.HandleFunc("/{id}", users.DeleteUser).Methods("DELETE")

	// Book Routes
	bookRoutes := router.PathPrefix("/books").Subrouter()
	bookRoutes.Use(authMiddleware)
	bookRoutes.HandleFunc("/purchase", func(w http.ResponseWriter, r *http.Request) {
		books.PurchaseBook(w, r, users)
	}).Methods("POST")

	bookRoutes.HandleFunc("/", books.GetBooks).Methods("GET")
	bookRoutes.HandleFunc("/", books.CreateBook).Methods("POST")
	bookRoutes.HandleFunc("/{id}", books.GetBookById).Methods("GET")
	bookRoutes.HandleFunc("/{id}", books.UpdateBook).Methods("POST")
	bookRoutes.HandleFunc("/{id}", books.DeleteBook).Methods("DELETE")

	// Run Server
	fmt.Println("server starting on port", s.addr)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.addr), router); err != nil {
		log.Fatal(err)
	}
}
