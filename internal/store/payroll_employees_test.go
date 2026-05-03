package store

import (
	"testing"

	"simpletask/internal/models"
)

func TestCreatePayrollEmployeeIDsAreUniqueAcrossCompanies(t *testing.T) {
	s := newTestStore(t)

	first, err := s.CreatePayrollEmployee(models.PayrollEmployee{
		CompanyID: "PC0001",
		LegalName: "First Employee",
	})
	if err != nil {
		t.Fatalf("create first employee: %v", err)
	}
	if first.ID != "EMP00001" {
		t.Fatalf("first ID = %q, want EMP00001", first.ID)
	}

	second, err := s.CreatePayrollEmployee(models.PayrollEmployee{
		CompanyID: "PC0002",
		LegalName: "Second Employee",
	})
	if err != nil {
		t.Fatalf("create second employee: %v", err)
	}
	if second.ID != "EMP00002" {
		t.Fatalf("second ID = %q, want EMP00002", second.ID)
	}
}
