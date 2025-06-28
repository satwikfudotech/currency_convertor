package currency

func ConvertWithoutPointers(amount, fromRate, toRate float64) float64 {
	return amount * (toRate / fromRate)
}

func ConvertWithPointers(amount float64, fromRate, toRate *float64) float64 {
	return amount * (*toRate / *fromRate)
}
