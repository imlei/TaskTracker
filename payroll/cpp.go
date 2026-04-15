package payroll

import "math"

// CPPInput is salary/wages path (Chapter 6 — not commission).
type CPPInput struct {
	// PI pensionable earnings for the pay period (includes taxable benefits as required).
	PI float64
	// D year-to-date CPP (employee) before this pay period.
	D float64
	// D2 year-to-date CPP2 before this pay period.
	D2 float64
	// PIYTD pensionable YTD before this pay period.
	PIYTD float64
	// P pay periods per year; used for basic exemption PI - 3500/P.
	P float64
	// PM months for proration (usually 12).
	PM float64
}

// CPPResult holds C and C2 after rounding.
type CPPResult struct {
	C, C2 float64
}

// CalculateCPP implements Chapter 6 formulas for Canada except Quebec (CPP + CPP2).
// rb selects CPP caps and rates; nil uses DefaultRateBook.
func CalculateCPP(rb *RateBook, in CPPInput) CPPResult {
	if in.P <= 0 {
		return CPPResult{}
	}
	r := rateBookOrDefault(rb)
	cpp := r.CPP

	pm := in.PM
	if pm <= 0 {
		pm = 12
	}

	exempt := cpp.BasicExemptionAnnual / in.P
	exempt = DropThirdDecimalCPPExemption(exempt)

	raw := cpp.TotalRate * (in.PI - exempt)
	if raw < 0 {
		raw = 0
	}

	capRemaining := cpp.MaxTotalEmployee*(pm/12) - in.D
	if capRemaining < 0 {
		capRemaining = 0
	}
	c := math.Min(capRemaining, raw)
	c = RoundCPP(c)

	ympeProrated := cpp.YMPE * (pm / 12)
	w := math.Max(in.PIYTD, ympeProrated)

	spread := in.PIYTD + in.PI - w
	if spread < 0 {
		spread = 0
	}
	raw2 := cpp.SecondAddRate * spread

	cap2Remaining := cpp.MaxSecondAdditional*(pm/12) - in.D2
	if cap2Remaining < 0 {
		cap2Remaining = 0
	}
	c2 := math.Min(cap2Remaining, raw2)
	c2 = RoundCPP(c2)

	return CPPResult{C: c, C2: c2}
}

// F5AdditionalCPP returns F5 using the default rate book’s CPP split ratios.
// For a custom RateBook, use (*RateBook).F5AdditionalCPP.
func F5AdditionalCPP(c, c2 float64) float64 {
	return defaultRateBookData.F5AdditionalCPP(c, c2)
}
