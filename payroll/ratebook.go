package payroll

import "errors"

// RateBook holds all CRA table-driven amounts and brackets for one edition (e.g. T4127-JAN-2026).
// Load from your backend (JSON) or use DefaultRateBook(). Nil *RateBook in API inputs means "use default".
//
// TaxYear is the calendar year these tables apply to (withholding year). CRA publishes a new
// T4127 edition most years; rates and brackets change — store a full RateBook per year in your DB
// when you need years not embedded in this module.
type RateBook struct {
	Edition string `json:"edition"`
	TaxYear int    `json:"tax_year"` // e.g. 2026 for T4127-JAN-2026 effective Jan 1, 2026

	Federal FederalSchedule `json:"federal"`

	// Provincial maps province of employment to Table 8.1 V/KP schedules (QC / OutsideCanada omitted).
	Provincial map[Province]ProvincialBracketSet `json:"provincial"`

	CPP CPPParams `json:"cpp"`
	EI  EIParams  `json:"ei"`

	// FederalLowestRate is the lowest federal bracket rate (K1, K2, K4 multipliers).
	FederalLowestRate float64 `json:"federal_lowest_rate"`
	CEA               float64 `json:"cea"` // Canada employment amount (Table 8.2)

	AlbertaK5PThreshold float64 `json:"alberta_k5p_threshold"`

	BPAFederal  BPAFederalParams    `json:"bpa_federal"`
	MBParams    ManitobaBPAMBParams `json:"mb_params"` // BPAMB when TD1MB not filed
	TCPDefaults TCPDefaults         `json:"tcp_defaults"` // Table 8.2 when TD1 not filed (fixed provinces)
}

// FederalSchedule is Table 8.1 federal R/K bands.
type FederalSchedule struct {
	Bands []FederalBand `json:"bands"`
}

// FederalBand is one row: income A < UpTo uses R and K (same semantics as existing federalBrackets).
type FederalBand struct {
	UpTo float64 `json:"up_to"`
	R    float64 `json:"r"`
	K    float64 `json:"k"`
}

// ProvincialBracketSet holds V and KP for one jurisdiction (Table 8.1).
type ProvincialBracketSet struct {
	Thresholds []float64 `json:"thresholds"` // lower bounds of each band (column A)
	V          []float64 `json:"v"`
	KP         []float64 `json:"kp"`
}

// FederalRK returns R and K for annual taxable income A.
func (fs *FederalSchedule) FederalRK(A float64) (R, K float64) {
	if fs == nil || len(fs.Bands) == 0 {
		return 0, 0
	}
	for i, b := range fs.Bands {
		if A < b.UpTo || i == len(fs.Bands)-1 {
			return b.R, b.K
		}
	}
	last := fs.Bands[len(fs.Bands)-1]
	return last.R, last.K
}

// CPPParams is Table 8.3–8.6 (Canada except Quebec) for salary CPP/CPP2.
type CPPParams struct {
	YMPE  float64 `json:"ympe"`
	YAMPE float64 `json:"yampe"`
	YMCE  float64 `json:"ymce"`

	TotalRate     float64 `json:"total_rate"`
	BaseRate      float64 `json:"base_rate"`
	FirstAddRate  float64 `json:"first_add_rate"`
	SecondAddRate float64 `json:"second_add_rate"`

	MaxTotalEmployee     float64 `json:"max_total_employee"`
	MaxBaseEmployee      float64 `json:"max_base_employee"`
	MaxFirstAdditional   float64 `json:"max_first_additional"`
	MaxSecondAdditional  float64 `json:"max_second_additional"`
	BasicExemptionAnnual float64 `json:"basic_exemption_annual"`
}

// EIParams is Table 8.7.
type EIParams struct {
	MaxInsurableEarnings float64 `json:"max_insurable_earnings"`
	RateCanadaExceptQC   float64 `json:"rate_canada_except_qc"`
	RateQC               float64 `json:"rate_qc"`
	MaxEmployeeCanada    float64 `json:"max_employee_canada"`
	MaxEmployeeQC        float64 `json:"max_employee_qc"`
}

// BPAFederalParams is the dynamic federal BPAF (Chapter 2).
type BPAFederalParams struct {
	Tier1MaxNI float64 `json:"tier1_max_ni"`
	Tier2MaxNI float64 `json:"tier2_max_ni"`
	BPAFull    float64 `json:"bpa_full"`
	BPAReduced float64 `json:"bpa_reduced"`
	// Reduction = (NI - Tier1MaxNI) * (ReductionNumerator / ReductionDenominator)
	ReductionNumerator   float64 `json:"reduction_numerator"`
	ReductionDenominator float64 `json:"reduction_denominator"`
}

// ManitobaBPAMBParams is BPAMB when TD1MB not filed.
type ManitobaBPAMBParams struct {
	Tier1MaxNI float64 `json:"tier1_max_ni"`
	Tier2MaxNI float64 `json:"tier2_max_ni"`
	FullAmount float64 `json:"full_amount"`
	// Reduction slope: (NI - Tier1MaxNI) * (FullAmount / Tier2Span); Tier2Span = Tier2MaxNI - Tier1MaxNI
	Tier2Span float64 `json:"tier2_span"`
}

// TCPDefaults holds provincial basic personal amounts (Table 8.2) for DefaultProvincialTCP.
// Manitoba and Yukon use formulas; set MB/YT to 0 here.
type TCPDefaults struct {
	AB, BC, NB, NL, NS, NT, NU, ON, PE, SK float64
}

// DefaultRateBook returns a copy of the embedded T4127-JAN-2026 tables.
func DefaultRateBook() RateBook {
	return defaultRateBookData
}

// FederalRK delegates to Federal.FederalRK.
func (rb *RateBook) FederalRK(A float64) (R, K float64) {
	if rb == nil {
		return defaultRateBookData.Federal.FederalRK(A)
	}
	return rb.Federal.FederalRK(A)
}

// ProvincialSet returns Table 8.1 data for a province, if present.
func (rb *RateBook) ProvincialSet(p Province) (ProvincialBracketSet, bool) {
	if rb == nil || rb.Provincial == nil {
		s, ok := defaultRateBookData.Provincial[p]
		return s, ok
	}
	s, ok := rb.Provincial[p]
	return s, ok
}

// F5AdditionalCPP returns F5 = C×(first_add/total)+C2 using book CPP rates.
func (rb *RateBook) F5AdditionalCPP(c, c2 float64) float64 {
	r := rateBookOrDefault(rb)
	cpp := r.CPP
	return c*(cpp.FirstAddRate/cpp.TotalRate) + c2
}

// FederalBPAF computes BPAF from net income NI using book parameters.
func (rb *RateBook) FederalBPAF(NI float64) float64 {
	r := rateBookOrDefault(rb)
	p := r.BPAFederal
	switch {
	case NI <= p.Tier1MaxNI:
		return p.BPAFull
	case NI < p.Tier2MaxNI:
		v := p.BPAFull - (NI-p.Tier1MaxNI)*(p.ReductionNumerator/p.ReductionDenominator)
		return RoundTax(v)
	default:
		return p.BPAReduced
	}
}

// ManitobaBPAMB computes BPAMB using book parameters.
func (rb *RateBook) ManitobaBPAMB(NI float64) float64 {
	r := rateBookOrDefault(rb)
	p := r.MBParams
	switch {
	case NI <= p.Tier1MaxNI:
		return p.FullAmount
	case NI < p.Tier2MaxNI:
		v := p.FullAmount - (NI-p.Tier1MaxNI)*(p.FullAmount/p.Tier2Span)
		return RoundTax(v)
	default:
		return 0
	}
}

// DefaultFederalTC returns TC when no federal TD1 (BPAF).
func (rb *RateBook) DefaultFederalTC(NI float64) float64 {
	return rb.FederalBPAF(NI)
}

// DefaultProvincialTCP returns TCP when no provincial TD1 (Table 8.2 + MB/YT rules).
func (rb *RateBook) DefaultProvincialTCP(p Province, NI float64) float64 {
	r := rateBookOrDefault(rb)
	d := r.TCPDefaults
	switch p {
	case AB:
		return d.AB
	case BC:
		return d.BC
	case MB:
		return r.ManitobaBPAMB(NI)
	case NB:
		return d.NB
	case NL:
		return d.NL
	case NS:
		return d.NS
	case NT:
		return d.NT
	case NU:
		return d.NU
	case ON:
		return d.ON
	case PE:
		return d.PE
	case SK:
		return d.SK
	case YT:
		return r.FederalBPAF(NI)
	case QC, OutsideCanada:
		return 0
	default:
		return 0
	}
}

// Validate performs minimal sanity checks (optional call after loading from JSON).
func (rb *RateBook) Validate() error {
	if rb == nil {
		return nil
	}
	if len(rb.Federal.Bands) == 0 {
		return errors.New("payroll.RateBook: federal.bands is empty")
	}
	if rb.CPP.TotalRate <= 0 || rb.CPP.BaseRate <= 0 {
		return errors.New("payroll.RateBook: cpp rates must be positive")
	}
	if rb.Provincial != nil {
		for p, set := range rb.Provincial {
			if len(set.V) == 0 || len(set.Thresholds) != len(set.V) {
				return errors.New("payroll.RateBook: provincial schedule invalid for " + p.String())
			}
		}
	}
	return nil
}

// MustValidate panics if Validate returns non-nil (for tests / startup).
func (rb *RateBook) MustValidate() {
	if rb == nil {
		return
	}
	if err := rb.Validate(); err != nil {
		panic(err)
	}
}

// Clone returns a deep copy sufficient for safe mutation (maps are replaced).
func (rb *RateBook) Clone() RateBook {
	if rb == nil {
		return DefaultRateBook()
	}
	out := *rb
	if rb.Provincial != nil {
		out.Provincial = make(map[Province]ProvincialBracketSet, len(rb.Provincial))
		for k, v := range rb.Provincial {
			out.Provincial[k] = cloneProvincialSet(v)
		}
	}
	out.Federal.Bands = append([]FederalBand(nil), rb.Federal.Bands...)
	return out
}

func cloneProvincialSet(s ProvincialBracketSet) ProvincialBracketSet {
	return ProvincialBracketSet{
		Thresholds: append([]float64(nil), s.Thresholds...),
		V:          append([]float64(nil), s.V...),
		KP:         append([]float64(nil), s.KP...),
	}
}
