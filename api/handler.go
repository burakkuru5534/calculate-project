package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/burakkuru5534/calculate-project/service"
)

type Handler struct {
	chipService *service.ChipService
}

func NewHandler(chipService *service.ChipService) *Handler {
	return &Handler{
		chipService: chipService,
	}
}

type transferRequest struct {
	FromPlayerID string `json:"fromPlayerId"`
	ToPlayerID   string `json:"toPlayerId"`
	Amount       int64  `json:"amount"`
}

type transferResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type balanceResponse struct {
	PlayerID    string `json:"playerId"`
	ChipBalance int64  `json:"chipBalance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/transfer-chips", h.handleTransferChips)
	mux.HandleFunc("/chip-balance/", h.handleChipBalance)
}

func (h *Handler) handleTransferChips(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	err := h.chipService.Transfer(req.FromPlayerID, req.ToPlayerID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSelfTransfer),
			errors.Is(err, service.ErrTransferTooLarge),
			errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrPlayerIDRequired):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInsufficientChips):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, transferResponse{
		Success: true,
		Message: "Transfer completed successfully",
	})
}

func (h *Handler) handleChipBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	playerID := strings.TrimPrefix(r.URL.Path, "/chip-balance/")

	balance, err := h.chipService.GetBalance(playerID)
	if err != nil {
		if errors.Is(err, service.ErrPlayerIDRequired) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		PlayerID:    playerID,
		ChipBalance: balance,
	})
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
