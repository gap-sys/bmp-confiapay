package helpers

import (
	"math"
)

func Round(val float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(val*factor) / factor

}

func EqualFloat(a, b float64) bool {
	return Round(a, 2) == Round(b, 2)
}

func DiffFloat(a, b float64) float64 {
	a = Round(a, 2)
	b = Round(b, 2)

	return math.Abs(a - b)

}
