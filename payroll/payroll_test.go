package payroll

import (
	"math"
	"testing"
)

func TestFederalRK(t *testing.T) {
	rb := DefaultRateBook()
	r, k := rb.FederalRK(50000)
	if r != 0.14 || k != 0 {
		t.Fatalf("mid bracket: got R=%v K=%v", r, k)
	}
	r, k = rb.FederalRK(60000)
	if r != 0.205 || k != 3804 {
		t.Fatalf("second bracket: got R=%v K=%v", r, k)
	}
}

func TestProvincialRK_ON(t *testing.T) {
	set, _ := ProvincialRates2026(ON)
	v, kp := ProvincialRK(set, 60000)
	if math.Abs(v-0.0915) > 1e-9 || math.Abs(kp-2210) > 0.01 {
		t.Fatalf("ON mid: V=%v KP=%v", v, kp)
	}
}

func TestCPPWeekly(t *testing.T) {
	// Weekly P=52: exemption drop to 2 decimals ~67.30
	r := CalculateCPP(nil, CPPInput{
		PI:    1000,
		D:     0,
		D2:    0,
		PIYTD: 0,
		P:     52,
		PM:    12,
	})
	if r.C <= 0 || r.C2 != 0 {
		t.Fatalf("expected base CPP, no CPP2: %+v", r)
	}
}

func TestCalculatePayPeriodON(t *testing.T) {
	res, err := CalculatePayPeriod(PayPeriodInput{
		Province:          ON,
		PayPeriodsPerYear: 26,
		Gross:             2000,
		YTDCPP:            0,
		YTDCPP2:           0,
		YTDCPPensionable:  0,
		YTDEI:             0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CPP <= 0 || res.EI <= 0 {
		t.Fatalf("expected CPP and EI: %+v", res)
	}
	if res.FederalTax <= 0 || res.ProvincialTax < 0 {
		t.Fatalf("unexpected tax: fed=%v prov=%v", res.FederalTax, res.ProvincialTax)
	}
}

func TestQuebecRejected(t *testing.T) {
	_, err := CalculatePayPeriod(PayPeriodInput{Province: QC, PayPeriodsPerYear: 26, Gross: 1000})
	if err == nil {
		t.Fatal("expected error for QC")
	}
}

func TestFactorS1(t *testing.T) {
	if FactorS1(26, 21) != 26.0/21.0 {
		t.Fatalf("S1 mismatch: %v", FactorS1(26, 21))
	}
}

func TestOption2IncomeTaxCRAExampleShape(t *testing.T) {
	// Illustrative numbers from Chapter 5 text (tax figures are fictitious).
	T1 := 2000.0
	T2 := 1560.17
	S1 := 26.0 / 21.0
	M := 2736.40
	M1 := 0.0
	L := 0.0
	T := Option2IncomeTax(T1, T2, S1, M, M1, L)
	// (3560.17/26*21 - 2736.40) ≈ 139.12
	want := (T1+T2-M1)/S1 - M + L
	if want < 0 {
		want = L
	}
	if diff := math.Abs(T - RoundTax(want)); diff > 0.02 {
		t.Fatalf("T=%v want ~%v", T, want)
	}
}

func TestSplitOption2IncomeTaxModes(t *testing.T) {
	T := 100.0
	t1, t2 := 600.0, 400.0
	f, p := SplitOption2IncomeTax(T, t1, t2, Option2TaxSplitConfig{Mode: TaxSplitProportionalToAnnual})
	if math.Abs(f-60) > 0.01 || math.Abs(p-40) > 0.01 {
		t.Fatalf("proportional: f=%v p=%v", f, p)
	}
	f, p = SplitOption2IncomeTax(T, t1, t2, Option2TaxSplitConfig{Mode: TaxSplitAllFederal})
	if f != 100 || p != 0 {
		t.Fatalf("all fed: f=%v p=%v", f, p)
	}
	f, p = SplitOption2IncomeTax(T, t1, t2, Option2TaxSplitConfig{Mode: TaxSplitFixedFraction, FederalFraction: 0.25})
	if math.Abs(f-25) > 0.01 || math.Abs(p-75) > 0.01 {
		t.Fatalf("fixed: f=%v p=%v", f, p)
	}
}

func TestCalculatePayPeriodOption2(t *testing.T) {
	// 21st biweekly pay: YTD gross 20000 + 1000 current => 21000 cumulative I.
	res, err := CalculatePayPeriodOption2(Option2PayPeriodInput{
		Province:                ON,
		PayPeriodsPerYear:       26,
		CurrentPayPeriod:        21,
		CumulativePeriodicGross: 21000,
		CumulativeF5A:           0,
		PE:                      21000,
		IE:                      21000,
		Pensionable:             1000,
		Insurable:               1000,
		MTax:                    2500,
		M1Tax:                   0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.S1 != 26.0/21.0 {
		t.Fatalf("S1: %v", res.S1)
	}
	if res.IncomeTaxTotal < 0 {
		t.Fatalf("tax: %v", res.IncomeTaxTotal)
	}
}

func TestTaxYearUnknownWithoutRates(t *testing.T) {
	_, err := CalculatePayPeriod(PayPeriodInput{
		Province:          ON,
		PayPeriodsPerYear: 26,
		Gross:             1000,
		TaxYear:           1999,
	})
	if err == nil {
		t.Fatal("expected error for unsupported year without Rates")
	}
}

func TestTaxYear2026Embedded(t *testing.T) {
	rb, err := RateBookForTaxYear(2026)
	if err != nil {
		t.Fatal(err)
	}
	if rb.TaxYear != 2026 {
		t.Fatalf("TaxYear: %d", rb.TaxYear)
	}
}

func TestCustomRateBookYMPE(t *testing.T) {
	rb := DefaultRateBook()
	rb.CPP.YMPE = 80000 // arbitrary override
	r := CalculateCPP(&rb, CPPInput{PI: 10000, D: 0, D2: 0, PIYTD: 0, P: 12, PM: 12})
	if r.C2 != 0 {
		t.Fatalf("expected no CPP2 below custom YMPE: %+v", r)
	}
}
