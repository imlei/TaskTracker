// Package reportexport renders remittance summaries and employee payslips for print/PDF export.
// Layouts follow typical CRA remittance and Canadian payroll slip conventions (see client PDF samples).
package reportexport

// Pair is a Previous / Current column (remittance) or Current / YTD (payslip).
type Pair struct {
	Previous string // remittance: prior period; payslip: leave empty when using YTD-only
	Current  string
	YTD      string // payslip YTD column; empty for remittance rows that use Previous/Current
}

// RemittanceReport is CRA source-deduction remittance summary (one payment / report).
type RemittanceReport struct {
	CompanyLegalName    string
	PayrollYear         int
	ReportNumber        int
	PaymentDateDisplay  string // e.g. "NOV 30, 2025"
	CRAAccountNumber    string // e.g. "732000914RP0001"
	EmployeeCount       int
	RemittanceFrequency string // e.g. "Monthly"

	TotalGrossPayroll Pair
	CPP               struct{ Employee, Employer, Total Pair }
	EI                struct{ Employee, Employer, Total Pair }
	FederalTax        Pair
	TotalToRemit      Pair

	TotalPaymentsCRA  string // e.g. "$842.42"
	TotalRemittance   string
	SourcePaymentNote string // footer line e.g. payment batch id
	PrintedAt         string // optional
}

// Payslip is one employee pay statement / cheque stub.
type Payslip struct {
	EmployeeID   string
	EmployeeName string // "LAST, FIRST"
	PeriodFrom   string
	PeriodTo     string

	RegularHours    Pair // Current / YTD
	NonTaxableTotal Pair
	RegularEarnings Pair
	Tips            Pair
	BasicRate       string
	BasicRateUnit   string // "Hourly"

	EI              Pair // label Canada Pension Plan / EI in template
	CPP             Pair
	ProvincialTax   Pair // optional; empty if not used
	OtherDeductions Pair // optional

	TotalTaxableGross Pair
	TotalDeductions   Pair
	NetPay            Pair
	VacationBalance   Pair

	DocumentNumber string // cheque / deposit ref
	PaymentDate    string
	NetPayWords    string // legal line
	NetPayAmount   string // "*** 1,160.56"
	PayeeName      string
	AddressLines   []string
}
