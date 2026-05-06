package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/burakkuru5534/calculate-project/service"
)

func TestTransferAndGetBalanceEndpoints(t *testing.T) {
	chipService := service.NewChipService()
	handler := NewHandler(chipService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	transferReq := httptest.NewRequest(http.MethodPost, "/transfer-chips", strings.NewReader(
		`{"fromPlayerId":"player-123","toPlayerId":"player-456","amount":2000}`,
	))
	transferRes := httptest.NewRecorder()

	mux.ServeHTTP(transferRes, transferReq)

	if transferRes.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", transferRes.Code)
	}

	balanceReq := httptest.NewRequest(http.MethodGet, "/chip-balance/player-123", nil)
	balanceRes := httptest.NewRecorder()
	mux.ServeHTTP(balanceRes, balanceReq)

	if balanceRes.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", balanceRes.Code)
	}
	if !strings.Contains(balanceRes.Body.String(), `"chipBalance":8000`) {
		t.Fatalf("expected sender balance to be 8000, response: %s", balanceRes.Body.String())
	}
}
