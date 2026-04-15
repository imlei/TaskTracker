package payroll

import "math"

// EIInput for insurable earnings (Chapter 7).
type EIInput struct {
	IE float64 // insurable earnings this period
	D1 float64 // YTD EI premium before this period
	// InQuebec uses QC rates and maximums.
	InQuebec bool
}

// CalculateEI returns employee EI premium for the pay period. nil rb uses DefaultRateBook.
func CalculateEI(rb *RateBook, in EIInput) float64 {
	r := rateBookOrDefault(rb)
	eip := r.EI

	var maxPrem, rate float64
	if in.InQuebec {
		maxPrem = eip.MaxEmployeeQC
		rate = eip.RateQC
	} else {
		maxPrem = eip.MaxEmployeeCanada
		rate = eip.RateCanadaExceptQC
	}
	remaining := maxPrem - in.D1
	if remaining < 0 {
		remaining = 0
	}
	raw := rate * in.IE
	if raw < 0 {
		raw = 0
	}
	return RoundEI(math.Min(remaining, raw))
}
