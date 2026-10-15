package main

import (
	"log"

	"github.com/Froggys974/e-commerce_golang/internal/api"
)

func main() {
	if err := api.Run(); err != nil {
		log.Fatalf("API error : %v", err)
	}
}
