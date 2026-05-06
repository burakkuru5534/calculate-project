package service

import "testing"

func TestSumService(t *testing.T) {
	s := &SumService{}

	result, err := s.Calculate([]int{1, 2, 3})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 6 {
		t.Fatalf("expected 6, got %v", result)
	}
}
