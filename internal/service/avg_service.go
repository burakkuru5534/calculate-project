package service

import "fmt"

type AvgService struct{}

func (a *AvgService) Calculate(nums []int) (float64, error) {
	if len(nums) == 0 {
		return 0, fmt.Errorf("empty input")
	}

	sum := 0
	for _, n := range nums {
		sum += n
	}

	return float64(sum) / float64(len(nums)), nil
}
