package reportexport

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderRemittanceSample(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderRemittanceSample(&buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "CRA SOURCE DEDUCTIONS") || !strings.Contains(s, "1180595 BC LTD") {
		t.Fatalf("unexpected remittance HTML output")
	}
}

func TestRenderPayslipSample(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderPayslipSample(&buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "XU, JING") || !strings.Contains(s, "4218 REPERT ST") {
		t.Fatal("payslip template missing expected fields")
	}
}
