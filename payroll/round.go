package payroll

import "math"

// RoundTax rounds income tax per Chapter 1 (third decimal ≥5 rounds second up).
func RoundTax(x float64) float64 {
	return roundToDecimals(x, 2, true)
}

// RoundCPP rounds CPP contributions per Chapter 1.
func RoundCPP(x float64) float64 {
	return roundToDecimals(x, 2, true)
}

// RoundEI rounds EI premiums per Chapter 1.
func RoundEI(x float64) float64 {
	return roundToDecimals(x, 2, true)
}

// DropThirdDecimalCPPExemption drops the third decimal on the basic exemption per pay (CPP).
func DropThirdDecimalCPPExemption(x float64) float64 {
	return math.Trunc(x*100) / 100
}

func roundToDecimals(x float64, places int, roundHalfUp bool) float64 {
	p := math.Pow10(places)
	v := x * p
	if roundHalfUp {
		return math.Floor(v+0.5) / p
	}
	return math.Trunc(v) / p
}
