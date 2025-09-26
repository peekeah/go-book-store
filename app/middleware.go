package app

import (
	"context"
	"net/http"
	"strings"

	"github.com/peekeah/book-store/handler"
	"github.com/peekeah/book-store/model"
	"github.com/peekeah/book-store/utils"
	"gorm.io/gorm"
)

func authenticate(db *gorm.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, "unauthorized"}
			res.Dispatch()
			return
		}

		tokenStr = strings.Replace(tokenStr, "Bearer ", "", 1)

		id, err := utils.VerifyJWTToken(tokenStr)
		if err != nil {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, "unauthorized"}
			res.Dispatch()
			return
		}

		// validate user in db
		user := model.User{}

		if err := db.First(&user, id).Error; err != nil {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, err.Error()}
			res.Dispatch()
			return
		}

		if user.ID == 0 {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, err.Error()}
			res.Dispatch()
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authorizeAdmin(db *gorm.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read from ctx
		userId := r.Context().Value("user_id")

		// validate user in db
		user := model.User{}

		if err := db.First(&user, userId).Error; err != nil {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, err.Error()}
			res.Dispatch()
			return
		}

		if user.ID == 0 {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, "unauthorized"}
			res.Dispatch()
			return
		}

		if user.Role != "admin" {
			res := handler.ErrorResponse{w, http.StatusUnauthorized, "only admin are allowed"}
			res.Dispatch()
			return
		}

		next.ServeHTTP(w, r)

		ctx := context.WithValue(r.Context(), "role", "admin")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
