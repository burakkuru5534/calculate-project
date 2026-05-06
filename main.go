package main

import (
	"log"
	"net/http"

	"github.com/burakkuru5534/calculate-project/internal/handler"
	"github.com/burakkuru5534/calculate-project/internal/service"
)

func main() {
	svc := &service.CalculatorService{}
	h := handler.NewHandler(svc)

	http.HandleFunc("/calculate", h.Calculate)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
