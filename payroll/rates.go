package payroll

// Edition identifies the embedded default rate set (see DefaultRateBook).
const Edition = "T4127-JAN-2026"

// Legacy numeric constants mirror defaultRateBookData for backward compatibility.
// Prefer reading from RateBook (e.g. DefaultRateBook().CPP.YMPE) when using configurable tables.
const (
	YMPE2026  = 74600.00
	YAMPE2026 = 85000.00
	YMCE2026  = 71100.00

	CPPTotalRate     = 0.0595
	CPPBaseRate      = 0.0495
	CPPFirstAddRate  = 0.0100
	CPPSecondAddRate = 0.0400

	MaxCPPTotalEmployee2026    = 4230.45
	MaxCPPBaseEmployee2026     = 3519.45
	MaxCPPFirstAdditional2026  = 711.00
	MaxCPPSecondAdditional2026 = 416.00
	PensionableSpanSecondTier  = 10400.00

	BasicExemptionAnnual = 3500.00

	MaxInsurableEarnings2026 = 68900.00
	EIRateCanadaExceptQC     = 0.0163
	EIRateQC                 = 0.0130
	MaxEIEmployeeCanada2026  = 1123.07
	MaxEIEmployeeQC2026      = 895.70

	CEA2026 = 1501.00

	FederalLowestRate   = 0.14
	AlbertaK5PThreshold = 4896.00
)

// ProvincialRates2026 returns provincial brackets from the default RateBook.
// Deprecated: use (*RateBook).ProvincialSet with your configured book.
func ProvincialRates2026(p Province) (ProvincialBracketSet, bool) {
	return defaultRateBookData.ProvincialSet(p)
}
