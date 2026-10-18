package main

import (
	"log"

	"github.com/Froggys974/e-commerce_golang/internal/sshserver"
)

func main() {
	if err := sshserver.Run(); err != nil {
		log.Fatalf("SSH error : %v", err)
	}
}
