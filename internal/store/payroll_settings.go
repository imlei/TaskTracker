package store

import (
	"encoding/json"
	"sort"
	"time"

	"simpletask/internal/payroll/calculator"
)

// ── Provincial rate types ──────────────────────────────────────────────────────

type ProvinceTaxBand struct {
	Upper float64 `json:"upper"` // 0 = no limit (top bracket)
	Rate  float64 `json:"rate"`
}

type ProvinceSurtax struct {
	Over float64 `json:"over"`
	Rate float64 `json:"rate"`
}

type ProvinceRateSetting struct {
	Code       string            `json:"code"`
	BPA        float64           `json:"bpa"`
	BottomRate float64           `json:"bottomRate"`
	Bands      []ProvinceTaxBand `json:"bands"`
	Surtax     []ProvinceSurtax  `json:"surtax,omitempty"`
}

// ── Rate Settings ──────────────────────────────────────────────────────────────

// PayrollRateSetting holds the editable CRA rate fields for one tax year.
type PayrollRateSetting struct {
	Year         int                   `json:"year"`
	CPPRate      float64               `json:"cppRate"`
	YMPE         float64               `json:"ympe"`
	YBE          float64               `json:"ybe"`
	CPPMaxEE     float64               `json:"cppMaxEe"`
	CPP2Rate     float64               `json:"cpp2Rate"`
	YAMPE        float64               `json:"yampe"`
	CPP2MaxEE    float64               `json:"cpp2MaxEe"`
	EIRate       float64               `json:"eiRate"`
	EIRateQC     float64               `json:"eiRateQc"`
	MaxInsurable float64               `json:"maxInsurable"`
	EIMaxEE      float64               `json:"eiMaxEe"`
	EIMaxEEQC    float64               `json:"eiMaxEeQc"`
	EIErFactor   float64               `json:"eiErFactor"`
	FederalBPA   float64               `json:"federalBpa"`
	Provincial   []ProvinceRateSetting `json:"provincial"`
	UpdatedAt    string                `json:"updatedAt"`
}

// GetPayrollRateSetting fetches stored overrides for a year.
// Returns nil, nil if no row exists (caller should use hardcoded defaults).
func (s *Store) GetPayrollRateSetting(year int) (*PayrollRateSetting, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.db.QueryRow(`
		SELECT year, cpp_rate, ympe, ybe, cpp_max_ee, cpp2_rate, yampe, cpp2_max_ee,
		       ei_rate, ei_rate_qc, max_insurable, ei_max_ee, ei_max_ee_qc, ei_er_factor,
		       federal_bpa, COALESCE(provincial_json,'[]'), updated_at
		FROM payroll_rate_settings WHERE year = ?`, year)

	var r PayrollRateSetting
	var provJSON string
	err := row.Scan(&r.Year, &r.CPPRate, &r.YMPE, &r.YBE, &r.CPPMaxEE,
		&r.CPP2Rate, &r.YAMPE, &r.CPP2MaxEE,
		&r.EIRate, &r.EIRateQC, &r.MaxInsurable, &r.EIMaxEE, &r.EIMaxEEQC, &r.EIErFactor,
		&r.FederalBPA, &provJSON, &r.UpdatedAt)
	if err != nil {
		return nil, nil // no row
	}
	_ = json.Unmarshal([]byte(provJSON), &r.Provincial)
	if r.Provincial == nil {
		r.Provincial = []ProvinceRateSetting{}
	}
	return &r, nil
}

// UpsertPayrollRateSetting saves (or overwrites) rate settings for a year.
func (s *Store) UpsertPayrollRateSetting(r PayrollRateSetting) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	provJSON, _ := json.Marshal(r.Provincial)
	_, err := s.db.Exec(`
		INSERT INTO payroll_rate_settings
		  (year, cpp_rate, ympe, ybe, cpp_max_ee, cpp2_rate, yampe, cpp2_max_ee,
		   ei_rate, ei_rate_qc, max_insurable, ei_max_ee, ei_max_ee_qc, ei_er_factor,
		   federal_bpa, provincial_json, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(year) DO UPDATE SET
		  cpp_rate=excluded.cpp_rate, ympe=excluded.ympe, ybe=excluded.ybe,
		  cpp_max_ee=excluded.cpp_max_ee, cpp2_rate=excluded.cpp2_rate,
		  yampe=excluded.yampe, cpp2_max_ee=excluded.cpp2_max_ee,
		  ei_rate=excluded.ei_rate, ei_rate_qc=excluded.ei_rate_qc,
		  max_insurable=excluded.max_insurable, ei_max_ee=excluded.ei_max_ee,
		  ei_max_ee_qc=excluded.ei_max_ee_qc, ei_er_factor=excluded.ei_er_factor,
		  federal_bpa=excluded.federal_bpa, provincial_json=excluded.provincial_json,
		  updated_at=excluded.updated_at`,
		r.Year, r.CPPRate, r.YMPE, r.YBE, r.CPPMaxEE,
		r.CPP2Rate, r.YAMPE, r.CPP2MaxEE,
		r.EIRate, r.EIRateQC, r.MaxInsurable, r.EIMaxEE, r.EIMaxEEQC, r.EIErFactor,
		r.FederalBPA, string(provJSON), r.UpdatedAt,
	)
	return err
}

// ListPayrollRateYears returns sorted list of years stored in DB.
func (s *Store) ListPayrollRateYears() []int {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT year FROM payroll_rate_settings ORDER BY year ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var years []int
	for rows.Next() {
		var y int
		if rows.Scan(&y) == nil {
			years = append(years, y)
		}
	}
	return years
}

// GetPayrollRatesForYear returns a TaxYear with DB overrides applied on top of
// the hardcoded defaults (2025 and 2026 are fully hardcoded; other years use 2025 as base).
func (s *Store) GetPayrollRatesForYear(year int) calculator.TaxYear {
	var base calculator.TaxYear
	switch year {
	case 2026:
		base = calculator.Rates2026()
	default:
		base = calculator.Rates2025()
	}
	base.Year = year

	override, _ := s.GetPayrollRateSetting(year)
	if override == nil {
		return base
	}
	base.CPPRate = override.CPPRate
	base.YMPE = override.YMPE
	base.YBE = override.YBE
	base.CPPMaxEmployeeAnnual = override.CPPMaxEE
	base.CPP2Rate = override.CPP2Rate
	base.YAMPE = override.YAMPE
	base.CPP2MaxEmployeeAnnual = override.CPP2MaxEE
	base.EIRate = override.EIRate
	base.EIRateQC = override.EIRateQC
	base.MaxInsurableEarnings = override.MaxInsurable
	base.EIMaxEmployeeAnnual = override.EIMaxEE
	base.EIMaxEmployeeAnnualQC = override.EIMaxEEQC
	base.EIEmployerFactor = override.EIErFactor
	base.FederalBPA = override.FederalBPA

	// Apply provincial overrides
	for _, prov := range override.Provincial {
		if prov.Code == "" {
			continue
		}
		pr := calculator.ProvincialRates{
			BPA:        prov.BPA,
			BottomRate: prov.BottomRate,
		}
		for _, b := range prov.Bands {
			pr.Bands = append(pr.Bands, calculator.TaxBand{Upper: b.Upper, Rate: b.Rate})
		}
		for _, st := range prov.Surtax {
			pr.SurtaxThresholds = append(pr.SurtaxThresholds, calculator.SurtaxThreshold{Over: st.Over, Rate: st.Rate})
		}
		if base.Provincial == nil {
			base.Provincial = map[string]calculator.ProvincialRates{}
		}
		base.Provincial[prov.Code] = pr
	}

	return base
}

// DefaultRateSetting returns a PayrollRateSetting populated from hardcoded defaults for the year.
func DefaultRateSetting(year int) PayrollRateSetting {
	var d calculator.TaxYear
	switch year {
	case 2026:
		d = calculator.Rates2026()
	default:
		d = calculator.Rates2025()
	}
	r := PayrollRateSetting{
		Year:         year,
		CPPRate:      d.CPPRate,
		YMPE:         d.YMPE,
		YBE:          d.YBE,
		CPPMaxEE:     d.CPPMaxEmployeeAnnual,
		CPP2Rate:     d.CPP2Rate,
		YAMPE:        d.YAMPE,
		CPP2MaxEE:    d.CPP2MaxEmployeeAnnual,
		EIRate:       d.EIRate,
		EIRateQC:     d.EIRateQC,
		MaxInsurable: d.MaxInsurableEarnings,
		EIMaxEE:      d.EIMaxEmployeeAnnual,
		EIMaxEEQC:    d.EIMaxEmployeeAnnualQC,
		EIErFactor:   d.EIEmployerFactor,
		FederalBPA:   d.FederalBPA,
	}

	// Sort provinces alphabetically for consistent display
	codes := make([]string, 0, len(d.Provincial))
	for code := range d.Provincial {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		prov := d.Provincial[code]
		ps := ProvinceRateSetting{
			Code:       code,
			BPA:        prov.BPA,
			BottomRate: prov.BottomRate,
		}
		for _, b := range prov.Bands {
			ps.Bands = append(ps.Bands, ProvinceTaxBand{Upper: b.Upper, Rate: b.Rate})
		}
		for _, st := range prov.SurtaxThresholds {
			ps.Surtax = append(ps.Surtax, ProvinceSurtax{Over: st.Over, Rate: st.Rate})
		}
		r.Provincial = append(r.Provincial, ps)
	}
	return r
}

// ── Plan Settings ──────────────────────────────────────────────────────────────

// PayrollPlan describes a subscription plan.
type PayrollPlan struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MaxCompanies int     `json:"maxCompanies"`
	MaxEmployees int     `json:"maxEmployees"`
	PriceMonthly float64 `json:"priceMonthly"`
	Description  string  `json:"description"`
	IsActive     bool    `json:"isActive"`
	SortOrder    int     `json:"sortOrder"`
}

// ListPayrollPlans returns all plans ordered by sort_order.
func (s *Store) ListPayrollPlans() []PayrollPlan {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT id, name, max_companies, max_employees, price_monthly,
		       description, is_active, sort_order
		FROM payroll_plans ORDER BY sort_order ASC`)
	if err != nil {
		return []PayrollPlan{}
	}
	defer rows.Close()

	var list []PayrollPlan
	for rows.Next() {
		var p PayrollPlan
		var active int
		if err := rows.Scan(&p.ID, &p.Name, &p.MaxCompanies, &p.MaxEmployees,
			&p.PriceMonthly, &p.Description, &active, &p.SortOrder); err == nil {
			p.IsActive = active != 0
			list = append(list, p)
		}
	}
	if list == nil {
		return []PayrollPlan{}
	}
	return list
}

// UpsertPayrollPlan saves (or overwrites) a plan.
func (s *Store) UpsertPayrollPlan(p PayrollPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	active := 0
	if p.IsActive {
		active = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO payroll_plans (id, name, max_companies, max_employees, price_monthly,
		                           description, is_active, sort_order)
		VALUES (?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
		  name=excluded.name, max_companies=excluded.max_companies,
		  max_employees=excluded.max_employees, price_monthly=excluded.price_monthly,
		  description=excluded.description, is_active=excluded.is_active,
		  sort_order=excluded.sort_order`,
		p.ID, p.Name, p.MaxCompanies, p.MaxEmployees, p.PriceMonthly,
		p.Description, active, p.SortOrder,
	)
	return err
}
