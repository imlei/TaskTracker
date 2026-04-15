package payroll

import "fmt"

// Option2PayPeriodInput is the data set for T4127 Chapter 5 (cumulative averaging).
//
// Cumulative periodic amounts (I, F, F2, F5A, U1) must include IYTD + current pay, per the guide’s definition of factor I.
// PE = pensionable earnings for the period plus PIYTD; IE = insurable earnings for the period plus IEYTD.
// MTax is combined federal + provincial tax withheld on periodic pay through the last pay period (not including current).
// M1Tax is combined YTD tax on non-periodic payments through the last pay period.
type Option2PayPeriodInput struct {
	Province Province

	PayPeriodsPerYear float64
	CurrentPayPeriod  int // 1 .. P
	MonthsCPP         float64

	// Cumulative periodic amounts through this pay (including current).
	CumulativePeriodicGross float64
	CumulativeF             float64
	CumulativeF2            float64
	CumulativeF5A           float64
	CumulativeU1            float64

	HD, F1 float64

	B1 float64 // YTD non-periodic before this pay
	B  float64 // current non-periodic payable now (0 if none)
	F3, F4, F5B, F5BYTD float64

	PE float64 // PI + PIYTD
	IE float64 // insurable this period + IEYTD

	// Current period for CPP/EI (same as Option 1).
	Pensionable float64
	Insurable   float64
	YTDCPP, YTDCPP2, YTDCPPensionable, YTDEI float64

	MTax  float64 // M
	M1Tax float64 // M1

	TC  *float64
	TCP *float64

	LCFPerPay float64
	LCPPerPay float64
	L         float64 // factor L — additional tax per pay

	OntarioDependantY OntarioTaxReductionY

	// TaxSplit configures federal vs provincial allocation of Option 2 combined tax T.
	// Zero value means TaxSplitProportionalToAnnual (same as pre-config behaviour).
	TaxSplit Option2TaxSplitConfig

	// Rates selects tables and brackets; if nil, TaxYear selects embedded data (see [ResolveRateBook]).
	Rates *RateBook
	// TaxYear is the calendar withholding year (e.g. 2026). If 0 with Rates nil, embedded default is used.
	TaxYear int
}

// Option2PayPeriodResult holds Option 2 deductions and annual factors.
type Option2PayPeriodResult struct {
	CPP, CPP2 float64
	EI        float64

	S1 float64

	T1Annual, T2Annual float64
	IncomeTaxTotal     float64 // T for this pay (federal + provincial combined formula)

	FederalTax    float64
	ProvincialTax float64

	AnnualTaxableIncomeA float64

	FederalParts    FederalTaxParts
	ProvincialParts ProvincialTaxParts
}

// CalculatePayPeriodOption2 implements Chapter 5 Option 2 for salaried employees outside Quebec.
func CalculatePayPeriodOption2(in Option2PayPeriodInput) (Option2PayPeriodResult, error) {
	if in.PayPeriodsPerYear <= 0 {
		return Option2PayPeriodResult{}, fmt.Errorf("PayPeriodsPerYear must be positive")
	}
	if in.CurrentPayPeriod < 1 || float64(in.CurrentPayPeriod) > in.PayPeriodsPerYear {
		return Option2PayPeriodResult{}, fmt.Errorf("CurrentPayPeriod must be between 1 and PayPeriodsPerYear")
	}
	if in.Province == QC {
		return Option2PayPeriodResult{}, fmt.Errorf("Quebec Option 2 requires Revenu Québec; not implemented")
	}

	rb, err := ResolveRateBook(in.TaxYear, in.Rates)
	if err != nil {
		return Option2PayPeriodResult{}, err
	}

	P := in.PayPeriodsPerYear
	PM := in.MonthsCPP
	if PM <= 0 {
		PM = 12
	}

	pi := in.Pensionable
	if pi <= 0 {
		pi = in.CumulativePeriodicGross // weak fallback
	}
	ie := in.Insurable
	if ie <= 0 {
		ie = in.CumulativePeriodicGross
	}

	cppR := CalculateCPP(rb, CPPInput{
		PI:    pi,
		D:     in.YTDCPP,
		D2:    in.YTDCPP2,
		PIYTD: in.YTDCPPensionable,
		P:     P,
		PM:    PM,
	})
	ei := CalculateEI(rb, EIInput{IE: ie, D1: in.YTDEI, InQuebec: false})

	S1 := FactorS1(P, in.CurrentPayPeriod)
	A := AnnualTaxableIncomeOption2(S1, in.CumulativePeriodicGross, in.CumulativeF, in.CumulativeF2, in.CumulativeF5A, in.CumulativeU1,
		in.B1, in.F4, in.F5BYTD, in.HD, in.F1, in.B, in.F3, in.F5B)

	grossK4 := GrossAnnualEmploymentOption2(S1, in.CumulativePeriodicGross, in.B1)

	NI := A + in.HD
	var tc float64
	if in.TC != nil {
		tc = *in.TC
	} else {
		tc = rb.DefaultFederalTC(NI)
	}
	var tcp float64
	if in.TCP != nil {
		tcp = *in.TCP
	} else {
		tcp = rb.DefaultProvincialTCP(in.Province, NI)
	}

	K2 := K2Option2(rb, S1, in.PE, in.IE, in.B1)
	fed := ComputeFederalAnnualWithK2(rb, A, grossK4, tc, K2, P, in.LCFPerPay)

	var prov ProvincialTaxParts
	if in.Province != OutsideCanada {
		set, ok := rb.ProvincialSet(in.Province)
		if ok {
			low := lowestRate(set)
			k2p := K2POption2(rb, low, S1, in.PE, in.IE, in.B1)
			prov = ComputeProvincialAnnualWithK2P(rb, in.Province, A, grossK4, tcp, k2p, P, in.LCPPerPay, in.OntarioDependantY)
		}
	}

	T := Option2IncomeTax(fed.T1, prov.T2, S1, in.MTax, in.M1Tax, in.L)

	fedW, provW := SplitOption2IncomeTax(T, fed.T1, prov.T2, in.TaxSplit)

	return Option2PayPeriodResult{
		CPP:                  cppR.C,
		CPP2:                 cppR.C2,
		EI:                   ei,
		S1:                   S1,
		T1Annual:             fed.T1,
		T2Annual:             prov.T2,
		IncomeTaxTotal:       T,
		FederalTax:           fedW,
		ProvincialTax:        provW,
		AnnualTaxableIncomeA: A,
		FederalParts:         fed,
		ProvincialParts:      prov,
	}, nil
}
