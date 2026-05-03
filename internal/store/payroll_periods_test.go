package store

import (
	"testing"

	"simpletask/internal/models"
)

func TestCreatePayrollPeriodIDsAreUniqueAcrossCompanies(t *testing.T) {
	s := newTestStore(t)

	first, err := s.CreatePayrollPeriod(models.PayrollPeriod{
		CompanyID:    "PC0001",
		PayDate:      "2026-01-15",
		PeriodStart:  "2026-01-01",
		PeriodEnd:    "2026-01-15",
		PaysPerYear:  24,
		PayFrequency: "Semi-monthly",
		PayrollType:  "regular",
	})
	if err != nil {
		t.Fatalf("create first period: %v", err)
	}
	if first.ID != "PP00001" {
		t.Fatalf("first ID = %q, want PP00001", first.ID)
	}

	second, err := s.CreatePayrollPeriod(models.PayrollPeriod{
		CompanyID:    "PC0002",
		PayDate:      "2026-01-31",
		PeriodStart:  "2026-01-16",
		PeriodEnd:    "2026-01-31",
		PaysPerYear:  12,
		PayFrequency: "Monthly",
		PayrollType:  "regular",
	})
	if err != nil {
		t.Fatalf("create second period: %v", err)
	}
	if second.ID != "PP00002" {
		t.Fatalf("second ID = %q, want PP00002", second.ID)
	}
}
