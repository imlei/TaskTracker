package payroll

// FactorS1 returns S1 = P / n where P is total pay periods in the year and n is the current pay period number (1-based).
// See Table 5.1 (Chapter 5).
func FactorS1(totalPayPeriods float64, currentPayPeriod int) float64 {
	if totalPayPeriods <= 0 || currentPayPeriod <= 0 {
		return 0
	}
	return totalPayPeriods / float64(currentPayPeriod)
}

// AnnualTaxableIncomeOption2 computes factor A for Option 2 (Chapter 5).
// I, F, F2, F5A, U1 are cumulative periodic amounts through the current pay period (including IYTD + current).
// B1 is YTD non-periodic before this pay; B is current non-periodic payable now (0 if none).
func AnnualTaxableIncomeOption2(S1, I, F, F2, F5A, U1, B1, F4, F5BYTD, HD, F1, B, F3, F5B float64) float64 {
	A := S1*(I-F-F2-F5A-U1) + (B1 - F4 - F5BYTD) + (B - F3 - F5B) - HD - F1
	if A < 0 {
		return 0
	}
	return A
}

// GrossAnnualEmploymentOption2 projects annual gross from employment for K4 / K4P (S1 × cumulative periodic gross + B1).
func GrossAnnualEmploymentOption2(S1, cumulativePeriodicGross, B1 float64) float64 {
	if S1 <= 0 {
		return 0
	}
	return S1*cumulativePeriodicGross + B1
}

// K2Option2 is federal factor K2 for Option 2 (non-Quebec). PE and IE include YTD + current (Chapter 5).
// nil rb uses DefaultRateBook.
func K2Option2(rb *RateBook, S1, PE, IE, B1 float64) float64 {
	r := rateBookOrDefault(rb)
	cpp := r.CPP
	eip := r.EI
	fl := r.FederalLowestRate

	cppBase := cpp.BaseRate * maxf(0, S1*PE+B1-cpp.BasicExemptionAnnual)
	if cppBase > cpp.MaxBaseEmployee {
		cppBase = cpp.MaxBaseEmployee
	}
	eiPrem := eip.RateCanadaExceptQC * (S1*IE + B1)
	if eiPrem > eip.MaxEmployeeCanada {
		eiPrem = eip.MaxEmployeeCanada
	}
	return fl*cppBase + fl*eiPrem
}

// K2POption2 is provincial factor K2P for Option 2 (Chapter 5).
// nil rb uses DefaultRateBook.
func K2POption2(rb *RateBook, lowestProvincialRate, S1, PE, IE, B1 float64) float64 {
	r := rateBookOrDefault(rb)
	cpp := r.CPP
	eip := r.EI

	cppBase := cpp.BaseRate * maxf(0, S1*PE+B1-cpp.BasicExemptionAnnual)
	if cppBase > cpp.MaxBaseEmployee {
		cppBase = cpp.MaxBaseEmployee
	}
	eiPrem := eip.RateCanadaExceptQC * (S1*IE + B1)
	if eiPrem > eip.MaxEmployeeCanada {
		eiPrem = eip.MaxEmployeeCanada
	}
	return lowestProvincialRate*cppBase + lowestProvincialRate*eiPrem
}

// Option2IncomeTax implements T = [((T1 + T2 – M1) / S1) – M] + L (Chapter 5).
// M is combined federal + provincial tax on periodic pay through the last pay period (not including current).
// M1 is combined YTD tax on non-periodic payments through the last pay period.
func Option2IncomeTax(T1, T2, S1, M, M1, L float64) float64 {
	if S1 <= 0 {
		return RoundTax(L)
	}
	raw := (T1+T2-M1)/S1 - M + L
	if raw < 0 {
		return RoundTax(L)
	}
	return RoundTax(raw)
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
