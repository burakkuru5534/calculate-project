package service

import (
	"errors"
	"sync"
)

const (
	InitialChips      int64 = 10_000
	MaxTransferAmount int64 = 5_000
)

var (
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
	ErrTransferTooLarge   = errors.New("transfer amount exceeds maximum limit of 5000")
	ErrSelfTransfer       = errors.New("self transfers are not allowed")
	ErrInsufficientChips  = errors.New("insufficient chips")
	ErrPlayerIDRequired   = errors.New("player id is required")
	ErrTransferIDRequired = errors.New("transfer id is required")
	ErrTransferIDConflict = errors.New("transfer id already used with different request payload")
	ErrTransferInProgress = errors.New("transfer is currently in progress")
)

type TransferStatus string

const (
	TransferStatusPending TransferStatus = "pending"
	TransferStatusSuccess TransferStatus = "success"
	TransferStatusFail    TransferStatus = "fail"
)

type TransferResult struct {
	TransferID string
	Status     TransferStatus
	Replayed   bool
}

type transferRecord struct {
	TransferID    string
	FromPlayerID  string
	ToPlayerID    string
	Amount        int64
	Status        TransferStatus
	FailureCode   string
	FailureReason string
}

type ChipService struct {
	mu        sync.RWMutex
	balances  map[string]int64
	transfers map[string]*transferRecord
}

func NewChipService() *ChipService {
	return &ChipService{
		balances:  make(map[string]int64),
		transfers: make(map[string]*transferRecord),
	}
}

func (s *ChipService) GetBalance(playerID string) (int64, error) {
	if playerID == "" {
		return 0, ErrPlayerIDRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.ensurePlayer(playerID), nil
}

func (s *ChipService) Transfer(transferID, fromPlayerID, toPlayerID string, amount int64) (TransferResult, error) {
	if transferID == "" {
		return TransferResult{}, ErrTransferIDRequired
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.transfers[transferID]; ok {
		if existing.FromPlayerID != fromPlayerID || existing.ToPlayerID != toPlayerID || existing.Amount != amount {
			return TransferResult{}, ErrTransferIDConflict
		}

		switch existing.Status {
		case TransferStatusPending:
			return TransferResult{
				TransferID: transferID,
				Status:     TransferStatusPending,
				Replayed:   true,
			}, ErrTransferInProgress
		case TransferStatusSuccess:
			return TransferResult{
				TransferID: transferID,
				Status:     TransferStatusSuccess,
				Replayed:   true,
			}, nil
		case TransferStatusFail:
			return TransferResult{
				TransferID: transferID,
				Status:     TransferStatusFail,
				Replayed:   true,
			}, codeToError(existing.FailureCode)
		}
	}

	record := &transferRecord{
		TransferID:   transferID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Amount:       amount,
		Status:       TransferStatusPending,
	}
	s.transfers[transferID] = record

	if fromPlayerID == "" || toPlayerID == "" {
		return s.markTransferFailed(record, ErrPlayerIDRequired)
	}
	if fromPlayerID == toPlayerID {
		return s.markTransferFailed(record, ErrSelfTransfer)
	}
	if amount <= 0 {
		return s.markTransferFailed(record, ErrInvalidAmount)
	}
	if amount > MaxTransferAmount {
		return s.markTransferFailed(record, ErrTransferTooLarge)
	}

	fromBalance := s.ensurePlayer(fromPlayerID)
	toBalance := s.ensurePlayer(toPlayerID)

	if fromBalance < amount {
		return s.markTransferFailed(record, ErrInsufficientChips)
	}

	// Keep debit+credit in a single critical section to guarantee atomicity.
	s.balances[fromPlayerID] = fromBalance - amount
	s.balances[toPlayerID] = toBalance + amount
	record.Status = TransferStatusSuccess

	return TransferResult{
		TransferID: transferID,
		Status:     TransferStatusSuccess,
	}, nil
}

func (s *ChipService) ensurePlayer(playerID string) int64 {
	if _, ok := s.balances[playerID]; !ok {
		s.balances[playerID] = InitialChips
	}

	return s.balances[playerID]
}

func (s *ChipService) markTransferFailed(record *transferRecord, err error) (TransferResult, error) {
	record.Status = TransferStatusFail
	record.FailureCode = errorToCode(err)
	record.FailureReason = err.Error()

	return TransferResult{
		TransferID: record.TransferID,
		Status:     TransferStatusFail,
	}, err
}

func errorToCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidAmount):
		return "invalid_amount"
	case errors.Is(err, ErrTransferTooLarge):
		return "transfer_too_large"
	case errors.Is(err, ErrSelfTransfer):
		return "self_transfer"
	case errors.Is(err, ErrInsufficientChips):
		return "insufficient_chips"
	case errors.Is(err, ErrPlayerIDRequired):
		return "player_id_required"
	case errors.Is(err, ErrTransferIDRequired):
		return "transfer_id_required"
	case errors.Is(err, ErrTransferIDConflict):
		return "transfer_id_conflict"
	case errors.Is(err, ErrTransferInProgress):
		return "transfer_in_progress"
	default:
		return "unknown"
	}
}

func codeToError(code string) error {
	switch code {
	case "invalid_amount":
		return ErrInvalidAmount
	case "transfer_too_large":
		return ErrTransferTooLarge
	case "self_transfer":
		return ErrSelfTransfer
	case "insufficient_chips":
		return ErrInsufficientChips
	case "player_id_required":
		return ErrPlayerIDRequired
	case "transfer_id_required":
		return ErrTransferIDRequired
	case "transfer_id_conflict":
		return ErrTransferIDConflict
	case "transfer_in_progress":
		return ErrTransferInProgress
	default:
		return errors.New("transfer failed")
	}
}
