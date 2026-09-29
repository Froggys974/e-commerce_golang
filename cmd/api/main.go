package main

import (
    "fmt"
    "os"
    "github.com/Froggys974/e-commerce_golang/internal/api"
)

func main() {
    if err:= api.Run(); err != nil {
        fmt.Printf("API error : %v", err)
        os.Exit(1)
    }
}