package main

import (
	"log"

	"github.com/Froggys974/e-commerce_golang/internal/tui/admin"
)

func main() {
	if err := admin.Run(); err != nil {
		log.Fatalf("Admin error : %v", err)
	}
}
