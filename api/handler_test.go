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
		`{"transferId":"tx-1","fromPlayerId":"player-123","toPlayerId":"player-456","amount":2000}`,
	))
	transferRes := httptest.NewRecorder()

	mux.ServeHTTP(transferRes, transferReq)

	if transferRes.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", transferRes.Code)
	}
	if !strings.Contains(transferRes.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success status in transfer response, got: %s", transferRes.Body.String())
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

func TestTransfer_IdempotentReplayViaAPI(t *testing.T) {
	chipService := service.NewChipService()
	handler := NewHandler(chipService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"transferId":"tx-dup","fromPlayerId":"player-1","toPlayerId":"player-2","amount":2000}`

	firstReq := httptest.NewRequest(http.MethodPost, "/transfer-chips", strings.NewReader(body))
	firstRes := httptest.NewRecorder()
	mux.ServeHTTP(firstRes, firstReq)
	if firstRes.Code != http.StatusOK {
		t.Fatalf("expected first status 200, got %d", firstRes.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/transfer-chips", strings.NewReader(body))
	secondRes := httptest.NewRecorder()
	mux.ServeHTTP(secondRes, secondReq)
	if secondRes.Code != http.StatusOK {
		t.Fatalf("expected second status 200, got %d", secondRes.Code)
	}
	if !strings.Contains(secondRes.Body.String(), "already processed") {
		t.Fatalf("expected replay message, got: %s", secondRes.Body.String())
	}

	balanceReq := httptest.NewRequest(http.MethodGet, "/chip-balance/player-1", nil)
	balanceRes := httptest.NewRecorder()
	mux.ServeHTTP(balanceRes, balanceReq)
	if !strings.Contains(balanceRes.Body.String(), `"chipBalance":8000`) {
		t.Fatalf("expected single debit (8000), got: %s", balanceRes.Body.String())
	}
}
