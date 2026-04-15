package payroll

// ProvincialTaxParts holds annual provincial components before per-period division.
type ProvincialTaxParts struct {
	V, KP  float64
	K1P    float64
	K2P    float64
	K4P    float64
	K5P    float64 // Alberta
	T4     float64
	V1, V2 float64 // Ontario
	S      float64 // Ontario, BC
	T2     float64
}

// OntarioTaxReductionY is the Y factor for Ontario factor S (dependants); default 0 if unknown.
type OntarioTaxReductionY float64

// ComputeProvincialAnnual calculates T4 and T2 for non-Quebec jurisdictions in Option 1.
// nil rb uses DefaultRateBook.
func ComputeProvincialAnnual(rb *RateBook, p Province, A, GrossAnnualEmployment, TCP float64, P, PM float64, C, EI float64, LCPPerPay float64, ontY OntarioTaxReductionY) ProvincialTaxParts {
	k2p := computeK2POption1(rb, p, P, PM, C, EI)
	return computeProvincialT2(rb, p, A, GrossAnnualEmployment, TCP, k2p, P, LCPPerPay, ontY)
}

func computeK2POption1(rb *RateBook, p Province, P, PM, C, EI float64) float64 {
	if p == QC || p == OutsideCanada {
		return 0
	}
	r := rateBookOrDefault(rb)
	set, ok := r.ProvincialSet(p)
	if !ok {
		return 0
	}
	low := lowestRate(set)
	cpp := r.CPP
	eip := r.EI

	maxBase := cpp.MaxBaseEmployee * (PM / 12)
	cppPart := P * C * (cpp.BaseRate / cpp.TotalRate)
	if cppPart > maxBase {
		cppPart = maxBase
	}
	eiPart := P * EI
	if eiPart > eip.MaxEmployeeCanada {
		eiPart = eip.MaxEmployeeCanada
	}
	return low*cppPart + low*eiPart
}

// ComputeProvincialAnnualWithK2P uses a precomputed K2P (Option 1 or Option 2 Chapter 5).
func ComputeProvincialAnnualWithK2P(rb *RateBook, p Province, A, GrossAnnualEmployment, TCP, K2P float64, P float64, LCPPerPay float64, ontY OntarioTaxReductionY) ProvincialTaxParts {
	return computeProvincialT2(rb, p, A, GrossAnnualEmployment, TCP, K2P, P, LCPPerPay, ontY)
}

func computeProvincialT2(rb *RateBook, p Province, A, GrossAnnualEmployment, TCP, K2P float64, P float64, LCPPerPay float64, ontY OntarioTaxReductionY) ProvincialTaxParts {
	var out ProvincialTaxParts
	if p == QC || p == OutsideCanada {
		return out
	}

	r := rateBookOrDefault(rb)
	set, ok := r.ProvincialSet(p)
	if !ok {
		return out
	}
	V, KP := ProvincialRK(set, A)
	out.V, out.KP = V, KP

	low := lowestRate(set)
	K1P := low * TCP
	out.K1P, out.K2P = K1P, K2P

	// Yukon K4P (Chapter 4 Step 5).
	if p == YT {
		ga := A
		if GrossAnnualEmployment > 0 {
			ga = GrossAnnualEmployment
		}
		out.K4P = minf(low*ga, low*r.CEA)
	}

	var k5p float64
	if p == AB {
		sum := K1P + K2P
		k5p = (sum - r.AlbertaK5PThreshold) * 0.25
		if k5p < 0 {
			k5p = 0
		}
		out.K5P = k5p
	}
	T4 := V*A - KP - K1P - K2P - out.K4P - k5p
	if T4 < 0 {
		T4 = 0
	}
	out.T4 = T4

	switch p {
	case ON:
		out.V1 = ontarioV1(T4)
		out.V2 = ontarioV2(A)
		out.S = ontarioS(T4, out.V1, float64(ontY))
	case BC:
		out.S = bcTaxReductionS(A, T4)
	}

	lcp := P * LCPPerPay
	if lcp < 0 {
		lcp = 0
	}
	T2 := T4 + out.V1 + out.V2 - out.S - lcp
	if T2 < 0 {
		T2 = 0
	}
	out.T2 = T2
	return out
}

func lowestRate(set ProvincialBracketSet) float64 {
	if len(set.V) == 0 {
		return 0
	}
	return set.V[0]
}

func ontarioV1(T4 float64) float64 {
	switch {
	case T4 <= 5818:
		return 0
	case T4 <= 7446:
		return 0.20 * (T4 - 5818)
	default:
		return (0.20 * (T4 - 5818)) + (0.36 * (T4 - 7446))
	}
}

func ontarioV2(A float64) float64 {
	switch {
	case A <= 20000:
		return 0
	case A <= 36000:
		return minf(300, 0.06*(A-20000))
	case A <= 48000:
		return minf(450, 300+0.06*(A-36000))
	case A <= 72000:
		return minf(600, 450+0.25*(A-48000))
	case A <= 200000:
		return minf(750, 600+0.25*(A-72000))
	default:
		return minf(900, 750+0.25*(A-200000))
	}
}

func ontarioS(T4, V1, Y float64) float64 {
	x := T4 + V1
	s := 2*(300+Y) - x
	if s < 0 {
		return 0
	}
	return minf(x, s)
}

func bcTaxReductionS(A, T4 float64) float64 {
	switch {
	case A <= 25570:
		return minf(T4, 575)
	case A <= 41722:
		return minf(T4, 575-(A-25570)*0.0356)
	default:
		return 0
	}
}

// AnnualTaxableIncomeOption1 computes factor A for periodic pay (Chapter 4 Step 1).
func AnnualTaxableIncomeOption1(P, I, F, F2, F5A, U1, HD, F1 float64) float64 {
	inner := P * (I - F - F2 - F5A - U1)
	A := inner - HD - F1
	return A
}

// GrossAnnualEmploymentIncome estimates box-14 gross for K4 (P × I when no separate benefits).
func GrossAnnualEmploymentIncome(P, I float64) float64 {
	if P <= 0 {
		return 0
	}
	return P * I
}
