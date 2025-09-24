package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/config"
	"github.com/peekeah/book-store/handler"
	"github.com/peekeah/book-store/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	addr string
	DB   *gorm.DB
}

func NewSever() *Server {
	return &Server{addr: config.GetConfig().Server.Port}
}

func (s *Server) MigragateDB() {
	dbConfig := config.GetConfig().DB

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port,
	)
	fmt.Println("dsn:", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connection failed")
	}

	model.DBMigrate(db)
	s.DB = db
}

func (s *Server) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from Go Book"))
	})

	// Auth Routes
	authRoutes := router.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/login", s.RequestHandler(handler.UserLogin)).Methods("POST")
	authRoutes.HandleFunc("/signup", s.RequestHandler(handler.CreateUser)).Methods("POST")

	// Authorized routes
	// User Routes
	userRoutes := router.PathPrefix("/users").Subrouter()
	userRoutes.Use(authMiddleware)
	userRoutes.HandleFunc("/", s.RequestHandler(handler.GetUsers)).Methods("GET")
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.GetUserById)).Methods("GET")
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.UpdateUser)).Methods("POST")
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.DeleteUser)).Methods("DELETE")

	// Book Routes
	bookRoutes := router.PathPrefix("/books").Subrouter()
	bookRoutes.Use(authMiddleware)
	bookRoutes.HandleFunc("/purchase", s.RequestHandler(handler.PurchaseBook)).Methods("POST")

	bookRoutes.HandleFunc("/", s.RequestHandler(handler.GetBooks)).Methods("GET")
	bookRoutes.HandleFunc("/", s.RequestHandler(handler.CreateBook)).Methods("POST")
	bookRoutes.HandleFunc("/{id}", s.RequestHandler(handler.GetBookById)).Methods("GET")
	bookRoutes.HandleFunc("/{id}", s.RequestHandler(handler.UpdateBook)).Methods("POST")
	bookRoutes.HandleFunc("/{id}", s.RequestHandler(handler.DeleteBook)).Methods("DELETE")

	// Run Server
	fmt.Println("server starting on port", s.addr)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", s.addr), router); err != nil {
		log.Fatal(err)
	}
}

type RequestHandler func(db *gorm.DB, w http.ResponseWriter, r *http.Request)

func (s *Server) RequestHandler(handler RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(s.DB, w, r)
	}
}
