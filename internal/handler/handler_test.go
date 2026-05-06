package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mock calculator
type MockCalculator struct{}

func (m *MockCalculator) Calculate(nums []int) (float64, error) {
	return 42, nil
}

// override handler for test
func TestCalculateHandler_DI(t *testing.T) {
	mock := &MockCalculator{}
	h := NewHandler(mock)

	reqBody := Request{
		Operation: "sum",
		Numbers:   []int{1, 2, 3},
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/calculate", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.Calculate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
