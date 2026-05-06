package service

type Calculator interface {
	Calculate([]int) (float64, error)
}
