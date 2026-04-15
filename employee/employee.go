// Package employee defines employee records and onboarding-related types for payroll.
package employee

import (
	"strings"
	"time"

	"github.com/imlei/prworks/payroll"
)

// MemberType classifies how the person is paid and which tax slips apply.
type MemberType int

const (
	MemberEmployee MemberType = iota
	MemberContractor
	MemberConstructionContractor
)

func (m MemberType) String() string {
	switch m {
	case MemberEmployee:
		return "employee"
	case MemberContractor:
		return "contractor"
	case MemberConstructionContractor:
		return "construction_contractor"
	default:
		return "unknown"
	}
}

// EmploymentCategory is used for display (e.g. PERMANENT).
type EmploymentCategory int

const (
	Permanent EmploymentCategory = iota
	Contract
)

func (c EmploymentCategory) String() string {
	switch c {
	case Permanent:
		return "PERMANENT"
	case Contract:
		return "CONTRACT"
	default:
		return "UNKNOWN"
	}
}

// SalaryType distinguishes salaried vs hourly-style pay.
type SalaryType int

const (
	SalarySalaried SalaryType = iota
	SalaryTimeBased
)

func (s SalaryType) String() string {
	switch s {
	case SalarySalaried:
		return "Salaried"
	case SalaryTimeBased:
		return "Time-Based"
	default:
		return ""
	}
}

// Status is the lifecycle state shown in the team list.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending_onboarding"
)

// Employee is a persisted team member used for payroll and listings.
type Employee struct {
	ID        string `json:"id"`
	LegalName string `json:"legalName"`
	Nickname  string `json:"nickname,omitempty"`
	Email     string `json:"email,omitempty"`
	Mobile    string `json:"mobile,omitempty"`

	MemberType MemberType `json:"memberType"`

	Position     string           `json:"position,omitempty"`
	PayFrequency string           `json:"payFrequency,omitempty"` // e.g. "Semi-Monthly"
	Status       Status           `json:"status"`
	Category     EmploymentCategory `json:"category"`
	SalaryType   SalaryType       `json:"salaryType"`

	HireDate         time.Time       `json:"hireDate"`
	Province         payroll.Province `json:"province"`
	SIN              string          `json:"sin,omitempty"`
	DateOfBirth      *time.Time      `json:"dateOfBirth,omitempty"`

	PayRate      float64 `json:"payRate,omitempty"`
	PayRateUnit  string  `json:"payRateUnit,omitempty"` // Hourly, Annually
	PaysPerYear  int     `json:"paysPerYear,omitempty"`
	HoursPerWeek float64 `json:"hoursPerWeek,omitempty"`

	PaidYTDThroughOtherPayroll *bool `json:"paidYTDThroughOtherPayroll,omitempty"`
	AutoVacation               *bool `json:"autoVacation,omitempty"`
}

// ValidateContact requires at least one of email or mobile (PayChequer-style invite).
func ValidateContact(email, mobile string) bool {
	e := strings.TrimSpace(email)
	m := strings.TrimSpace(mobile)
	return e != "" || m != ""
}

// TypeDisplay returns a two-line style label like "Time-Based PERMANENT".
func (e Employee) TypeDisplay() string {
	return e.SalaryType.String() + " " + e.Category.String()
}
