package api

import (
	"fmt"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/db"
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
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok !\n"))
	})

	fmt.Println("Server on :8000")
	return http.ListenAndServe(":8000", mux)
}
