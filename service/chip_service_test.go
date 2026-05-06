package service

import (
	"fmt"
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

	result, err := svc.Transfer("tx-1", "player-1", "player-2", 2_000)
	if err != nil {
		t.Fatalf("Transfer returned error: %v", err)
	}
	if result.Status != TransferStatusSuccess {
		t.Fatalf("expected transfer status success, got %s", result.Status)
	}
	if result.Replayed {
		t.Fatalf("expected replayed=false, got true")
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

	_, err := svc.Transfer("tx-large", "player-1", "player-2", 11_000)
	if err == nil {
		t.Fatal("expected error for insufficient chips, got nil")
	}
	if err != ErrTransferTooLarge {
		t.Fatalf("expected ErrTransferTooLarge due to amount limit, got %v", err)
	}

	// This value is under max transfer and should fail specifically for insufficient balance.
	_, err = svc.Transfer("tx-2", "player-1", "player-2", 4_000)
	if err != nil {
		t.Fatalf("expected first transfer to succeed, got %v", err)
	}
	_, err = svc.Transfer("tx-3", "player-1", "player-2", 4_000)
	if err != nil {
		t.Fatalf("expected second transfer to succeed, got %v", err)
	}
	_, err = svc.Transfer("tx-4", "player-1", "player-2", 4_000)
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
			_, err := svc.Transfer("tx-"+tc.name, tc.fromID, tc.toID, tc.amount)
			if err != tc.wantErr {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestTransfer_IdempotentReplayDoesNotDoubleDebit(t *testing.T) {
	svc := NewChipService()
	transferID := "tx-idempotent-1"

	first, err := svc.Transfer(transferID, "player-1", "player-2", 2_000)
	if err != nil {
		t.Fatalf("expected first transfer success, got %v", err)
	}
	if first.Replayed {
		t.Fatal("expected first transfer replayed=false")
	}

	second, err := svc.Transfer(transferID, "player-1", "player-2", 2_000)
	if err != nil {
		t.Fatalf("expected idempotent replay success, got %v", err)
	}
	if !second.Replayed {
		t.Fatal("expected replayed=true for duplicate transfer id")
	}
	if second.Status != TransferStatusSuccess {
		t.Fatalf("expected replay status success, got %s", second.Status)
	}

	fromBalance, _ := svc.GetBalance("player-1")
	toBalance, _ := svc.GetBalance("player-2")
	if fromBalance != 8_000 {
		t.Fatalf("expected sender to be debited once (8000), got %d", fromBalance)
	}
	if toBalance != 12_000 {
		t.Fatalf("expected receiver to be credited once (12000), got %d", toBalance)
	}
}

func TestTransfer_SameIDWithDifferentPayloadFails(t *testing.T) {
	svc := NewChipService()

	if _, err := svc.Transfer("tx-same-id", "player-1", "player-2", 1_000); err != nil {
		t.Fatalf("expected first transfer success, got %v", err)
	}
	if _, err := svc.Transfer("tx-same-id", "player-1", "player-2", 2_000); err != ErrTransferIDConflict {
		t.Fatalf("expected ErrTransferIDConflict, got %v", err)
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
		i := i
		go func() {
			defer wg.Done()
			_, _ = svc.Transfer(fmt.Sprintf("tx-ab-%d", i), "player-a", "player-b", transferAmount)
		}()
		go func() {
			defer wg.Done()
			_, _ = svc.Transfer(fmt.Sprintf("tx-ba-%d", i), "player-b", "player-a", transferAmount)
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
