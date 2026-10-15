package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Froggys974/e-commerce_golang/internal/auth"
	"github.com/Froggys974/e-commerce_golang/internal/db"
	"github.com/Froggys974/e-commerce_golang/internal/product"
)

func Run() error {
	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()
	fmt.Printf("DB up & connected\n")

	if err := db.Migrate(database); err != nil {
		return err
	}
	fmt.Printf("Schema ok\n")

	mux := http.NewServeMux()

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return errors.New("JWT_SECRET manquant ou invalide")
	}

	authHandler := auth.NewHandler(database, []byte(jwtSecret))
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/confirm", authHandler.Confirm)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/forgot", authHandler.Forgot)
	mux.HandleFunc("POST /auth/reset", authHandler.Reset)
	mux.HandleFunc("GET /me", authHandler.RequireAuth(authHandler.Me))

	productHandler := product.NewHandler(database)
	mux.HandleFunc("POST /admin/products", authHandler.RequireAdmin(productHandler.Create))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok !\n"))
	})

	fmt.Println("Server on :8000")
	return http.ListenAndServe(":8000", mux)
}
