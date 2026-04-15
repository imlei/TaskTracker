package reportexport

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var (
	remittanceTmpl *template.Template
	payslipTmpl    *template.Template
)

func init() {
	var err error
	remittanceTmpl, err = template.ParseFS(templateFS, "templates/remittance.html.tmpl")
	if err != nil {
		panic(err)
	}
	payslipTmpl, err = template.ParseFS(templateFS, "templates/payslip.html.tmpl")
	if err != nil {
		panic(err)
	}
}

// RenderRemittance writes HTML suitable for print / Save as PDF.
func RenderRemittance(w io.Writer, r *RemittanceReport) error {
	return remittanceTmpl.Execute(w, r)
}

// RenderPayslip writes one employee payslip HTML.
func RenderPayslip(w io.Writer, p *Payslip) error {
	return payslipTmpl.Execute(w, p)
}

// RenderRemittanceSample writes the 1180595 BC remittance report #63 sample.
func RenderRemittanceSample(w io.Writer) error {
	return RenderRemittance(w, SampleRemittanceReport63())
}

// RenderPayslipSample writes the April 2025 payslip sample (XU, JING).
func RenderPayslipSample(w io.Writer) error {
	return RenderPayslip(w, SamplePayslipXuApril2025())
}
