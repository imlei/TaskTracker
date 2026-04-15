package payroll

import "fmt"

// PayPeriodInput is the minimum data set for Option 1 periodic payroll (non-commission, non-Quebec provincial tax).
type PayPeriodInput struct {
	Province Province

	PayPeriodsPerYear float64 // P
	MonthsCPP         float64 // PM, default 12 if 0

	// Gross remuneration for the period (I) — box 14 style before non-periodic add-ons.
	Gross float64
	// Pensionable and insurable earnings for the period (often equal to Gross + taxable benefits).
	Pensionable float64
	Insurable  float64

	// YTD amounts before this pay (D, D1, D2, PIYTD).
	YTDCPP, YTDCPP2, YTDCPPensionable, YTDEI float64

	// Payroll deductions (Chapter 4).
	F, F2, U1, HD, F1 float64

	// If nil, TC and TCP are derived from BPA / provincial defaults using projected NI = A + HD.
	TC  *float64
	TCP *float64

	LCFPerPay float64 // labour-sponsored funds per pay
	LCPPerPay float64 // provincial labour-sponsored per pay

	// Ontario dependant amounts for factor Y (optional).
	OntarioDependantY OntarioTaxReductionY

	// Rates selects tables and brackets. If nil, TaxYear selects embedded data (see [ResolveRateBook]).
	Rates *RateBook
	// TaxYear is the calendar withholding year (e.g. 2026). If 0 with Rates nil, embedded default is used.
	TaxYear int
}

// PayPeriodResult contains statutory deductions for one pay period.
type PayPeriodResult struct {
	CPP, CPP2 float64
	EI        float64

	FederalTaxAnnual    float64
	ProvincialTaxAnnual float64
	FederalTax          float64
	ProvincialTax       float64

	FederalParts    FederalTaxParts
	ProvincialParts ProvincialTaxParts

	AnnualTaxableIncomeA float64
}

// CalculatePayPeriod runs CPP, EI, and income tax per T4127 Option 1 for a standard periodic pay.
//
// Quebec: QPP and provincial tax are not modeled here; use Revenu Québec rules. This function still computes federal tax using non-Quebec CPP/EI credits unless you swap inputs.
//
// Outside Canada: provincial T2 is 0; federal tax follows the “outside Canada” branch only if you apply the surtax in a separate layer (not included here).
func CalculatePayPeriod(in PayPeriodInput) (PayPeriodResult, error) {
	if in.PayPeriodsPerYear <= 0 {
		return PayPeriodResult{}, fmt.Errorf("PayPeriodsPerYear must be positive")
	}
	if in.Province == QC {
		return PayPeriodResult{}, fmt.Errorf("Quebec provincial tax and QPP require Revenu Québec formulas; not implemented in this core package")
	}

	rb, err := ResolveRateBook(in.TaxYear, in.Rates)
	if err != nil {
		return PayPeriodResult{}, err
	}

	P := in.PayPeriodsPerYear
	PM := in.MonthsCPP
	if PM <= 0 {
		PM = 12
	}

	pi := in.Pensionable
	if pi <= 0 {
		pi = in.Gross
	}
	ie := in.Insurable
	if ie <= 0 {
		ie = in.Gross
	}

	cppR := CalculateCPP(rb, CPPInput{
		PI:    pi,
		D:     in.YTDCPP,
		D2:    in.YTDCPP2,
		PIYTD: in.YTDCPPensionable,
		P:     P,
		PM:    PM,
	})
	f5 := rb.F5AdditionalCPP(cppR.C, cppR.C2)
	f5a := f5

	A := AnnualTaxableIncomeOption1(P, in.Gross, in.F, in.F2, f5a, in.U1, in.HD, in.F1)
	if A < 0 {
		A = 0
	}

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

	ei := CalculateEI(rb, EIInput{IE: ie, D1: in.YTDEI, InQuebec: false})

	grossAnnual := GrossAnnualEmploymentIncome(P, in.Gross)

	fed := ComputeFederalAnnual(rb, A, grossAnnual, tc, P, PM, cppR.C, ei, in.LCFPerPay)

	var prov ProvincialTaxParts
	if in.Province != OutsideCanada {
		prov = ComputeProvincialAnnual(rb, in.Province, A, grossAnnual, tcp, P, PM, cppR.C, ei, in.LCPPerPay, in.OntarioDependantY)
	}

	fedPer := RoundTax(fed.T1 / P)
	provPer := RoundTax(prov.T2 / P)

	return PayPeriodResult{
		CPP:                  cppR.C,
		CPP2:                 cppR.C2,
		EI:                   ei,
		FederalTaxAnnual:     fed.T1,
		ProvincialTaxAnnual:  prov.T2,
		FederalTax:           fedPer,
		ProvincialTax:        provPer,
		FederalParts:       fed,
		ProvincialParts:    prov,
		AnnualTaxableIncomeA: A,
	}, nil
}
