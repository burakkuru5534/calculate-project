package main

import (
	"log"
	"net/http"

	"github.com/burakkuru5534/calculate-project/internal/handler"
)

func main() {
	// handler oluştur
	h := handler.NewHandler()

	// route bağla
	http.HandleFunc("/calculate", h.Calculate)

	// server start
	log.Println("Server starting on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
