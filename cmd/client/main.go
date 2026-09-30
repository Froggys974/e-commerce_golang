package main

import (
	"fmt"
	"github.com/Froggys974/e-commerce_golang/internal/client"
	"os"
)

func main() {
	if err := client.Run(); err != nil {
		fmt.Printf("Client error : %v", err)
		os.Exit(1)
	}
}
