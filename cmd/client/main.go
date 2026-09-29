package main

import (
    "fmt"
    "os"
    "github.com/Froggys974/e-commerce_golang/internal/client"
)

func main() {
    if err:= client.Run(); err != nil {
        fmt.Printf("Client error : %v", err)
        os.Exit(1)
    }
}