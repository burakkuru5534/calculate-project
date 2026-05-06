package main

import (
	"log"
	"net/http"

	"github.com/burakkuru5534/calculate-project/api"
	"github.com/burakkuru5534/calculate-project/service"
)

func main() {
	chipService := service.NewChipService()
	handler := api.NewHandler(chipService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := ":8080"
	log.Printf("chip transfer service is running on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
