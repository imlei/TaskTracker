package payroll

import "math"

// FederalBPAF returns the dynamic federal basic personal amount (Chapter 2) using NI = A + HD.
// Uses the default rate book; for configurable BPAF use (*RateBook).FederalBPAF.
func FederalBPAF(NI float64) float64 {
	return defaultRateBookData.FederalBPAF(NI)
}

// ManitobaBPAMB returns BPAMB for Manitoba when TD1MB not filed.
func ManitobaBPAMB(NI float64) float64 {
	return defaultRateBookData.ManitobaBPAMB(NI)
}

// DefaultFederalTC returns TC when no federal TD1: BPAF(NI), or 0 for non-resident (caller sets).
func DefaultFederalTC(NI float64) float64 {
	return defaultRateBookData.DefaultFederalTC(NI)
}

// DefaultProvincialTCP returns TCP when no provincial TD1 (Table 8.2 + MB/YT rules).
func DefaultProvincialTCP(p Province, NI float64) float64 {
	return defaultRateBookData.DefaultProvincialTCP(p, NI)
}

// minf is a local helper.
func minf(a, b float64) float64 {
	return math.Min(a, b)
}
