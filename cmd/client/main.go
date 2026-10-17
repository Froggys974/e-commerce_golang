package main

import (
	"log"

	"github.com/Froggys974/e-commerce_golang/internal/tui/client"
)

func main() {
	if err := client.Run(); err != nil {
		log.Fatalf("Client error : %v", err)
	}
}
