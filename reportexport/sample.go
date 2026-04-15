package reportexport

// SampleRemittanceReport63 returns data matching "Remittance Report 63" (1180595 BC Ltd, Nov 2025).
func SampleRemittanceReport63() *RemittanceReport {
	r := &RemittanceReport{
		CompanyLegalName:    "1180595 BC LTD",
		PayrollYear:         2025,
		ReportNumber:        63,
		PaymentDateDisplay:  "NOV 30, 2025",
		CRAAccountNumber:    "732000914RP0001",
		EmployeeCount:       4,
		RemittanceFrequency: "Monthly",
		TotalPaymentsCRA:    "$842.42",
		TotalRemittance:     "$842.42",
		SourcePaymentNote:   "SOURCE DEDUCTIONS payment",
		PrintedAt:           "10-Dec-2025 4:10:13 PM",
	}
	r.TotalGrossPayroll = Pair{Previous: "4,905.79", Current: "4,629.48"}
	r.CPP.Employee = Pair{Previous: "205.82", Current: "218.08"}
	r.CPP.Employer = Pair{Previous: "205.82", Current: "218.08"}
	r.CPP.Total = Pair{Previous: "411.64", Current: "436.16"}
	r.EI.Employee = Pair{Previous: "80.45", Current: "75.92"}
	r.EI.Employer = Pair{Previous: "112.64", Current: "106.29"}
	r.EI.Total = Pair{Previous: "193.09", Current: "182.21"}
	r.FederalTax = Pair{Previous: "127.73", Current: "224.05"}
	r.TotalToRemit = Pair{Previous: "732.46", Current: "842.42"}
	return r
}

// SamplePayslipXuApril2025 returns the first employee block from "Payroll 202504" sample PDF.
func SamplePayslipXuApril2025() *Payslip {
	return &Payslip{
		EmployeeID:   "000038",
		EmployeeName: "XU, JING",
		PeriodFrom:   "01/04/2025",
		PeriodTo:     "30/04/2025",
		RegularHours: Pair{Current: "56.50", YTD: "218.33"},
		NonTaxableTotal: Pair{Current: "0.00", YTD: "0.00"},
		RegularEarnings: Pair{Current: "983.10", YTD: "3,798.94"},
		Tips:            Pair{Current: "254.00", YTD: "983.00"},
		BasicRate:       "17.40",
		BasicRateUnit:   "Hourly",
		EI:              Pair{Current: "20.29", YTD: "78.42"},
		CPP:             Pair{Current: "56.25", YTD: "215.10"},
		TotalTaxableGross: Pair{Current: "1,237.10", YTD: "4,781.94"},
		TotalDeductions:   Pair{Current: "76.54", YTD: "293.52"},
		NetPay:            Pair{Current: "1,160.56", YTD: "4,488.42"},
		VacationBalance:   Pair{Current: "0.00", YTD: "0.00"},
		DocumentNumber:    "000295",
		PaymentDate:       "30/04/2025",
		NetPayWords:       "****One Thousand One Hundred Sixty and 56/100****",
		NetPayAmount:      "*** 1,160.56",
		PayeeName:         "JING XU",
		AddressLines: []string{
			"4218 REPERT ST",
			"VANCOUVER, BRITISH COLUMBIA",
			"V5R 2H7",
		},
	}
}
