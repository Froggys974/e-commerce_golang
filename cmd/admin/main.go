package main

import (
    "fmt"
    "os"
    "github.com/Froggys974/e-commerce_golang/internal/admin"
)

func main() {
    if err:= admin.Run(); err != nil {
        fmt.Printf("Admin error : %v", err)
        os.Exit(1)
    }
}