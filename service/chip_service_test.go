package service

import (
	"sync"
	"testing"
)

func TestGetBalance_DefaultValue(t *testing.T) {
	svc := NewChipService()

	balance, err := svc.GetBalance("player-1")
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}

	if balance != InitialChips {
		t.Fatalf("expected initial balance %d, got %d", InitialChips, balance)
	}
}

func TestTransfer_Success(t *testing.T) {
	svc := NewChipService()

	err := svc.Transfer("player-1", "player-2", 2_000)
	if err != nil {
		t.Fatalf("Transfer returned error: %v", err)
	}

	fromBalance, _ := svc.GetBalance("player-1")
	toBalance, _ := svc.GetBalance("player-2")

	if fromBalance != 8_000 {
		t.Fatalf("expected sender balance 8000, got %d", fromBalance)
	}
	if toBalance != 12_000 {
		t.Fatalf("expected receiver balance 12000, got %d", toBalance)
	}
}

func TestTransfer_FailsWithInsufficientChips(t *testing.T) {
	svc := NewChipService()

	err := svc.Transfer("player-1", "player-2", 11_000)
	if err == nil {
		t.Fatal("expected error for insufficient chips, got nil")
	}
	if err != ErrTransferTooLarge {
		t.Fatalf("expected ErrTransferTooLarge due to amount limit, got %v", err)
	}

	// This value is under max transfer and should fail specifically for insufficient balance.
	err = svc.Transfer("player-1", "player-2", 4_000)
	if err != nil {
		t.Fatalf("expected first transfer to succeed, got %v", err)
	}
	err = svc.Transfer("player-1", "player-2", 4_000)
	if err != nil {
		t.Fatalf("expected second transfer to succeed, got %v", err)
	}
	err = svc.Transfer("player-1", "player-2", 4_000)
	if err != ErrInsufficientChips {
		t.Fatalf("expected ErrInsufficientChips, got %v", err)
	}
}

func TestTransfer_ValidationErrors(t *testing.T) {
	svc := NewChipService()

	testCases := []struct {
		name    string
		fromID  string
		toID    string
		amount  int64
		wantErr error
	}{
		{
			name:    "self transfer",
			fromID:  "player-1",
			toID:    "player-1",
			amount:  100,
			wantErr: ErrSelfTransfer,
		},
		{
			name:    "too large",
			fromID:  "player-1",
			toID:    "player-2",
			amount:  6_000,
			wantErr: ErrTransferTooLarge,
		},
		{
			name:    "non positive amount",
			fromID:  "player-1",
			toID:    "player-2",
			amount:  0,
			wantErr: ErrInvalidAmount,
		},
		{
			name:    "missing player id",
			fromID:  "",
			toID:    "player-2",
			amount:  100,
			wantErr: ErrPlayerIDRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Transfer(tc.fromID, tc.toID, tc.amount)
			if err != tc.wantErr {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestTransfer_ConcurrentSafety(t *testing.T) {
	svc := NewChipService()
	const goroutines = 100
	const transferAmount int64 = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Run transfers both ways to stress concurrent updates and preserve totals.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = svc.Transfer("player-a", "player-b", transferAmount)
		}()
		go func() {
			defer wg.Done()
			_ = svc.Transfer("player-b", "player-a", transferAmount)
		}()
	}

	wg.Wait()

	balanceA, _ := svc.GetBalance("player-a")
	balanceB, _ := svc.GetBalance("player-b")

	total := balanceA + balanceB
	expectedTotal := InitialChips * 2
	if total != expectedTotal {
		t.Fatalf("expected total chips %d, got %d", expectedTotal, total)
	}
}
