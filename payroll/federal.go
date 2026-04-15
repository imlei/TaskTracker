package payroll

// FederalTaxParts exposes intermediate factors for auditing (T4127 Option 1).
type FederalTaxParts struct {
	R, K   float64
	K1     float64
	K2     float64
	K4     float64
	T3, T1 float64
}

// ComputeFederalAnnual computes T3 and T1 for Canada except Quebec, outside Canada,
// and beyond provincial limits — salaried Option 1 (no Quebec abatement branch here).
// nil rb uses DefaultRateBook.
func ComputeFederalAnnual(rb *RateBook, A, GrossAnnualEmployment, TC float64, P, PM float64, C, EI float64, LCFPerPay float64) FederalTaxParts {
	r := rateBookOrDefault(rb)
	if A < 0 {
		A = 0
	}
	R, K := r.FederalRK(A)
	K1 := r.FederalLowestRate * TC

	maxBase := r.CPP.MaxBaseEmployee * (PM / 12)
	if maxBase < 0 {
		maxBase = 0
	}
	cppCreditAnnual := P * C * (r.CPP.BaseRate / r.CPP.TotalRate)
	if cppCreditAnnual > maxBase {
		cppCreditAnnual = maxBase
	}
	maxEI := r.EI.MaxEmployeeCanada
	eiCreditAnnual := P * EI
	if eiCreditAnnual > maxEI {
		eiCreditAnnual = maxEI
	}
	K2 := r.FederalLowestRate*cppCreditAnnual + r.FederalLowestRate*eiCreditAnnual

	ga := GrossAnnualEmployment
	if ga <= 0 {
		ga = A
	}
	K4 := minf(r.FederalLowestRate*ga, r.FederalLowestRate*r.CEA)

	T3 := R*A - K - K1 - K2 - K4
	if T3 < 0 {
		T3 = 0
	}

	lcfAnnual := P * LCFPerPay
	if lcfAnnual < 0 {
		lcfAnnual = 0
	}
	T1 := T3 - lcfAnnual
	if T1 < 0 {
		T1 = 0
	}

	return FederalTaxParts{R: R, K: K, K1: K1, K2: K2, K4: K4, T3: T3, T1: T1}
}

// ComputeFederalAnnualWithK2 is used by Option 2 (Chapter 5): K2 follows the S1 / PE / IE / B1 formulas.
// nil rb uses DefaultRateBook.
func ComputeFederalAnnualWithK2(rb *RateBook, A, grossAnnualEmploymentK4, TC, K2, P float64, LCFPerPay float64) FederalTaxParts {
	r := rateBookOrDefault(rb)
	if A < 0 {
		A = 0
	}
	R, K := r.FederalRK(A)
	K1 := r.FederalLowestRate * TC

	ga := grossAnnualEmploymentK4
	if ga <= 0 {
		ga = A
	}
	K4 := minf(r.FederalLowestRate*ga, r.FederalLowestRate*r.CEA)

	T3 := R*A - K - K1 - K2 - K4
	if T3 < 0 {
		T3 = 0
	}

	lcfAnnual := P * LCFPerPay
	if lcfAnnual < 0 {
		lcfAnnual = 0
	}
	T1 := T3 - lcfAnnual
	if T1 < 0 {
		T1 = 0
	}

	return FederalTaxParts{R: R, K: K, K1: K1, K2: K2, K4: K4, T3: T3, T1: T1}
}
