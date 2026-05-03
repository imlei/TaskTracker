package store

import (
	"testing"
	"time"
)

func TestGetYTDBeforePeriodUsesCalculatedAndFinalizedPayroll(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO payroll_periods
		  (id, company_id, period_start, period_end, pay_date, pays_per_year, pay_frequency, payroll_type, status, created_at, updated_at)
		VALUES
		  ('PPYTD01', 'PCYTD', '2026-01-01', '2026-01-15', '2026-01-15', 24, 'Semi-monthly', 'regular', 'finalized', ?, ?),
		  ('PPYTD02', 'PCYTD', '2026-01-16', '2026-01-31', '2026-01-31', 24, 'Semi-monthly', 'regular', 'open', ?, ?)`,
		now, now, now, now)
	if err != nil {
		t.Fatalf("seed periods: %v", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO payroll_entries
		  (id, period_id, employee_id, company_id, gross_pay, cpp_ee, cpp2_ee, ei_ee, status, created_at, updated_at)
		VALUES
		  ('PEYTD01', 'PPYTD01', 'EMPYTD01', 'PCYTD', 1000, 50, 5, 20, 'calculated', ?, ?)`,
		now, now)
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	ytd := s.GetYTDBeforePeriod("EMPYTD01", "PPYTD02", "2026")
	if ytd.Gross != 1000 {
		t.Fatalf("Gross = %v, want 1000", ytd.Gross)
	}
	if ytd.CPPEe != 50 || ytd.CPP2Ee != 5 || ytd.EIEe != 20 {
		t.Fatalf("deduction YTD = CPP %v CPP2 %v EI %v, want 50 5 20", ytd.CPPEe, ytd.CPP2Ee, ytd.EIEe)
	}
}
