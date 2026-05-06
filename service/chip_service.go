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
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrTransferTooLarge  = errors.New("transfer amount exceeds maximum limit of 5000")
	ErrSelfTransfer      = errors.New("self transfers are not allowed")
	ErrInsufficientChips = errors.New("insufficient chips")
	ErrPlayerIDRequired  = errors.New("player id is required")
)

type ChipService struct {
	mu       sync.RWMutex
	balances map[string]int64
}

func NewChipService() *ChipService {
	return &ChipService{
		balances: make(map[string]int64),
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

func (s *ChipService) Transfer(fromPlayerID, toPlayerID string, amount int64) error {
	if fromPlayerID == "" || toPlayerID == "" {
		return ErrPlayerIDRequired
	}
	if fromPlayerID == toPlayerID {
		return ErrSelfTransfer
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > MaxTransferAmount {
		return ErrTransferTooLarge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fromBalance := s.ensurePlayer(fromPlayerID)
	toBalance := s.ensurePlayer(toPlayerID)

	if fromBalance < amount {
		return ErrInsufficientChips
	}

	// Keep debit+credit in a single critical section to guarantee atomicity.
	s.balances[fromPlayerID] = fromBalance - amount
	s.balances[toPlayerID] = toBalance + amount

	return nil
}

func (s *ChipService) ensurePlayer(playerID string) int64 {
	if _, ok := s.balances[playerID]; !ok {
		s.balances[playerID] = InitialChips
	}

	return s.balances[playerID]
}
