package service

type CalculatorService struct{}

func (s *CalculatorService) Calculate(numbers []int) (int, float64) {
	sum := 0
	for _, n := range numbers {
		sum += n
	}

	avg := float64(sum) / float64(len(numbers))
	return sum, avg
}
