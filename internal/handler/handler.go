package handler

import (
	"encoding/json"
	"net/http"

	"github.com/burakkuru5534/calculate-project/internal/service"
)

type CalculateRequest struct {
	Numbers []int `json:"numbers"`
}

type CalculateResponse struct {
	Sum     int     `json:"sum"`
	Average float64 `json:"average"`
}

type Handler struct {
	svc *service.CalculatorService
}

func NewHandler(svc *service.CalculatorService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Numbers) <= 0 {
		http.Error(w, "len numbers cannot be zero.", http.StatusBadRequest)
		return
	}

	sum, avg := h.svc.Calculate(req.Numbers)

	res := CalculateResponse{
		Sum:     sum,
		Average: avg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
