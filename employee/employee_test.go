package employee

import (
	"testing"

	"github.com/imlei/prworks/payroll"
)

func TestValidateContact(t *testing.T) {
	if !ValidateContact("a@b.co", "") {
		t.Fatal("email only should pass")
	}
	if !ValidateContact("", "6045550100") {
		t.Fatal("mobile only should pass")
	}
	if ValidateContact("", "") {
		t.Fatal("empty should fail")
	}
	if ValidateContact("   ", "  ") {
		t.Fatal("whitespace should fail")
	}
}

func TestValidateSIN(t *testing.T) {
	// Valid test SIN (passes Luhn) — do not use real numbers in production logs.
	valid := "046454286"
	if !ValidateSIN(valid) {
		t.Errorf("expected valid SIN %s", valid)
	}
	if ValidateSIN("000000000") {
		t.Fatal("all zeros invalid")
	}
	if ValidateSIN("123") {
		t.Fatal("short invalid")
	}
}

func TestEmployeeTypeDisplay(t *testing.T) {
	e := Employee{
		SalaryType: SalaryTimeBased,
		Category:   Permanent,
		Province:   payroll.BC,
	}
	if got := e.TypeDisplay(); got != "Time-Based PERMANENT" {
		t.Fatalf("got %q", got)
	}
}
