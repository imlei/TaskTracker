// Example: biweekly Ontario employee, 2026 T4127-JAN rates.
package main

import (
	"fmt"

	"github.com/imlei/prworks/payroll"
)

func main() {
	res, err := payroll.CalculatePayPeriod(payroll.PayPeriodInput{
		Province:          payroll.ON,
		PayPeriodsPerYear: 26,
		Gross:             2500.00,
		YTDCPP:            0,
		YTDCPP2:           0,
		YTDCPPensionable:  0,
		YTDEI:             0,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Edition: %s\n", payroll.Edition)
	fmt.Printf("Annual taxable income (A): %.2f\n", res.AnnualTaxableIncomeA)
	fmt.Printf("CPP: %.2f  CPP2: %.2f  EI: %.2f\n", res.CPP, res.CPP2, res.EI)
	fmt.Printf("Federal tax (period): %.2f\n", res.FederalTax)
	fmt.Printf("Ontario tax (period): %.2f\n", res.ProvincialTax)
	fmt.Printf("Total statutory (period): %.2f\n", res.CPP+res.CPP2+res.EI+res.FederalTax+res.ProvincialTax)
}
