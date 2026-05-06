package handler

import (
	"encoding/json"
	"net/http"

	"github.com/burakkuru5534/calculate-project/internal/service"
)

type Request struct {
	Operation string `json:"operation"`
	Numbers   []int  `json:"numbers"`
}

type Response struct {
	Result float64 `json:"result"`
}

// Handler struct (dependency injection için)
type Handler struct{}

// Constructor (ileride dependency eklemek için hazır)
func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// sadece POST kabul et
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request

	// JSON decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// validation
	if len(req.Numbers) == 0 {
		http.Error(w, "numbers cannot be empty", http.StatusBadRequest)
		return
	}

	if req.Operation == "" {
		http.Error(w, "operation is required", http.StatusBadRequest)
		return
	}

	// strategy seçimi
	calculator := getCalculator(req.Operation)
	if calculator == nil {
		http.Error(w, "unsupported operation", http.StatusBadRequest)
		return
	}

	// business logic
	result, err := calculator.Calculate(req.Numbers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// response
	res := Response{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")

	// context kullanımı (şu an basit ama gösterilmiş oldu)
	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	default:
		json.NewEncoder(w).Encode(res)
	}
}

// strategy selector
func getCalculator(op string) service.Calculator {
	switch op {
	case "sum":
		return &service.SumService{}
	case "avg":
		return &service.AvgService{}
	default:
		return nil
	}
}
