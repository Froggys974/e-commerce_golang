package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Froggys974/e-commerce_golang/internal/auth"
	"github.com/Froggys974/e-commerce_golang/internal/db"
	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/user"
)

const (
	addr            = ":8000"
	minSecretLength = 32
)

func Run() error {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < minSecretLength {
		return errors.New("variable d'environnement JWT_SECRET manquante ou trop courte (32 caractères minimum)")
	}

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()
	log.Println("DB up & connected")

	if err := db.Migrate(database); err != nil {
		return err
	}
	log.Println("Schema ok")

	userService := user.NewService(user.NewRepository(database), user.HashPassword, user.CheckPassword)
	authService := auth.NewService(userService, auth.NewRepository(database), []byte(secret))

	if err := seedAdmin(userService); err != nil {
		return err
	}

	authHandler := auth.NewHandler(userService, authService)
	userHandler := user.NewHandler(userService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/confirm", authHandler.Confirm)
	mux.HandleFunc("POST /auth/confirm/resend", authHandler.ResendConfirmCode)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/forgot", authHandler.RequestPasswordReset)
	mux.HandleFunc("POST /auth/reset", authHandler.ResetPassword)
	mux.HandleFunc("GET /me", authHandler.RequireAuth(authHandler.Me))

	mux.HandleFunc("GET /admin/users", authHandler.RequireAdmin(userHandler.List))
	mux.HandleFunc("POST /admin/users", authHandler.RequireAdmin(userHandler.Create))
	mux.HandleFunc("PUT /admin/users/{id}", authHandler.RequireAdmin(userHandler.Update))
	mux.HandleFunc("DELETE /admin/users/{id}", authHandler.RequireAdmin(userHandler.Delete))
	mux.HandleFunc("POST /admin/users/{id}/confirm", authHandler.RequireAdmin(userHandler.Confirm))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		httpjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server on %s", addr)
	return server.ListenAndServe()
}

func seedAdmin(userService *user.Service) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Println("ADMIN_EMAIL ou ADMIN_PASSWORD vide : pas d'admin créé")
		return nil
	}

	if err := userService.EnsureAdmin(context.Background(), email, password); err != nil {
		return fmt.Errorf("création de l'admin : %w", err)
	}
	log.Printf("Admin ok : %s", email)
	return nil
}
