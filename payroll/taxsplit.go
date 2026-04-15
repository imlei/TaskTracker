package payroll

// Option2TaxSplitConfig controls how combined Option 2 income tax T is allocated to federal vs provincial
// withholding lines on a pay statement.
//
// T4127 gives T as a single amount; CRA does not mandate how to split it for pay stubs. Persist this
// policy in your backend and pass it each run so behaviour stays stable and auditable.
type Option2TaxSplitConfig struct {
	Mode TaxSplitMode

	// FederalFraction is used when Mode == TaxSplitFixedFraction.
	// Valid range is [0, 1]. Provincial is T minus federal after rounding.
	FederalFraction float64
}

// TaxSplitMode selects the allocation rule for Option 2 combined tax T.
type TaxSplitMode int

const (
	// TaxSplitProportionalToAnnual allocates T in the ratio T1Annual : T2Annual (default, backward compatible).
	TaxSplitProportionalToAnnual TaxSplitMode = iota
	// TaxSplitAllFederal assigns all of T to federal withholding.
	TaxSplitAllFederal
	// TaxSplitAllProvincial assigns all of T to provincial withholding.
	TaxSplitAllProvincial
	// TaxSplitFixedFraction assigns Federal = round(T × FederalFraction), Provincial = T − Federal.
	TaxSplitFixedFraction
)

// DefaultOption2TaxSplit returns the recommended default: proportional to annual T1 and T2.
func DefaultOption2TaxSplit() Option2TaxSplitConfig {
	return Option2TaxSplitConfig{Mode: TaxSplitProportionalToAnnual}
}

// SplitOption2IncomeTax allocates withholding T to federal and provincial amounts.
// T is the combined income tax for the pay period (Option 2 formula); T1Annual and T2Annual are
// annual federal and provincial tax before per-period division.
func SplitOption2IncomeTax(T, T1Annual, T2Annual float64, cfg Option2TaxSplitConfig) (federal, provincial float64) {
	if T == 0 {
		return 0, 0
	}
	switch cfg.Mode {
	case TaxSplitAllFederal:
		return RoundTax(T), 0
	case TaxSplitAllProvincial:
		return 0, RoundTax(T)
	case TaxSplitFixedFraction:
		f := cfg.FederalFraction
		if f < 0 {
			f = 0
		}
		if f > 1 {
			f = 1
		}
		fed := RoundTax(T * f)
		return fed, RoundTax(T - fed)
	default: // TaxSplitProportionalToAnnual and any unset mode
		sum := T1Annual + T2Annual
		if sum > 0 {
			fed := RoundTax(T * T1Annual / sum)
			return fed, RoundTax(T - fed)
		}
		return RoundTax(T), 0
	}
}
