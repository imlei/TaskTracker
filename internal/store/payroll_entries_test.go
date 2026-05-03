package store

import (
	"testing"
	"time"

	"simpletask/internal/models"
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

func TestUpsertPayrollEntryPaymentUpdatePreservesCalculatedYTDAndStatus(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO payroll_entries
		  (id, period_id, employee_id, company_id, hours, pay_rate, gross_pay,
		   cpp_ee, cpp2_ee, ei_ee, federal_tax, provincial_tax,
		   total_deductions, net_pay, ytd_gross, ytd_cpp_ee, ytd_cpp2_ee, ytd_ei_ee,
		   payment_type, status, created_at, updated_at)
		VALUES
		  ('PEPAY01', 'PPPAY01', 'EMPPAY01', 'PCPAY', 0, 1000, 1000,
		   50, 5, 20, 100, 40,
		   215, 785, 2000, 100, 10, 40,
		   'cheque', 'calculated', ?, ?)`,
		now, now)
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	_, err = s.UpsertPayrollEntry(models.PayrollEntry{
		PeriodID:    "PPPAY01",
		EmployeeID:  "EMPPAY01",
		Hours:       0,
		PayRate:     1000,
		GrossPay:    1000,
		PaymentType: "deposit",
	})
	if err != nil {
		t.Fatalf("upsert payment type: %v", err)
	}

	var status, paymentType string
	var ytdGross, ytdCPP, ytdCPP2, ytdEI float64
	err = s.db.QueryRow(`
		SELECT status, payment_type, ytd_gross, ytd_cpp_ee, ytd_cpp2_ee, ytd_ei_ee
		FROM payroll_entries WHERE id='PEPAY01'`).
		Scan(&status, &paymentType, &ytdGross, &ytdCPP, &ytdCPP2, &ytdEI)
	if err != nil {
		t.Fatalf("read updated entry: %v", err)
	}
	if status != "calculated" || paymentType != "deposit" {
		t.Fatalf("status/paymentType = %q/%q, want calculated/deposit", status, paymentType)
	}
	if ytdGross != 2000 || ytdCPP != 100 || ytdCPP2 != 10 || ytdEI != 40 {
		t.Fatalf("YTD = gross %v CPP %v CPP2 %v EI %v, want 2000 100 10 40", ytdGross, ytdCPP, ytdCPP2, ytdEI)
	}
}
