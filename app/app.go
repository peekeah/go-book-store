package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peekeah/book-store/config"
	"github.com/peekeah/book-store/handler"
	"github.com/peekeah/book-store/logger"
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
	l := logger.Get()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from Go Book"))
	})

	// Logger
	router.Use(logger.ReqMiddleware)

	// Health
	router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("."))
	})

	// Auth Routes
	authRoutes := router.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/login", s.RequestHandler(handler.UserLogin)).Methods("POST")
	authRoutes.HandleFunc("/signup", s.RequestHandler(handler.CreateUser)).Methods("POST")

	// Authorized routes
	// User Routes
	userRoutes := router.PathPrefix("/users").Subrouter()
	userRoutes.Use(s.MiddlewareHandler(authenticate))
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.GetUserById)).Methods("GET")
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.UpdateUser)).Methods("PUT")
	userRoutes.HandleFunc("/{id}", s.RequestHandler(handler.DeleteUser)).Methods("DELETE")

	// Admin user routes
	userAdminRoutes := userRoutes.PathPrefix("/").Subrouter()
	userAdminRoutes.Use(s.MiddlewareHandler(authorizeAdmin))
	userAdminRoutes.HandleFunc("/", s.RequestHandler(handler.GetUsers)).Methods("GET")

	// Book Routes
	bookRoutes := router.PathPrefix("/books").Subrouter()
	bookRoutes.Use(s.MiddlewareHandler(authenticate))
	bookRoutes.HandleFunc("/purchase/{id}", s.RequestHandler(handler.PurchaseBook)).Methods("POST")

	bookRoutes.HandleFunc("/", s.RequestHandler(handler.GetBooks)).Methods("GET")
	bookRoutes.HandleFunc("/{id}", s.RequestHandler(handler.GetBookById)).Methods("GET")

	// Admin book routes
	bookAdminRoutes := bookRoutes.PathPrefix("/").Subrouter()
	bookAdminRoutes.Use(s.MiddlewareHandler(authorizeAdmin))
	bookAdminRoutes.HandleFunc("/{id}", s.RequestHandler(handler.UpdateBook)).Methods("POST")
	bookAdminRoutes.HandleFunc("/{id}", s.RequestHandler(handler.DeleteBook)).Methods("DELETE")
	bookAdminRoutes.HandleFunc("/", s.RequestHandler(handler.CreateBook)).Methods("POST")

	// Run Server
	l.Info().
		Str("port", s.addr).
		Msgf("Starting Go Book Store App on port '%s'", s.addr)

	l.Fatal().
		Err(http.ListenAndServe(":"+s.addr, router)).
		Msg("Go Book Store App Closed")
}

type RequestHandler func(db *gorm.DB, w http.ResponseWriter, r *http.Request)

func (s *Server) RequestHandler(handler RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(s.DB, w, r)
	}
}

type MiddlewareHandler func(db *gorm.DB, next http.Handler) http.Handler

func (s *Server) MiddlewareHandler(mw MiddlewareHandler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return mw(s.DB, next)
	}
}
