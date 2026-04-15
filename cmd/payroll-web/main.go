// Payroll-web serves static files from web/ and a small JSON API for employees.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/imlei/prworks/employee"
	"github.com/imlei/prworks/payroll"
	"github.com/imlei/prworks/reportexport"
)

type store struct {
	mu         sync.Mutex
	nextSerial int
	items      []employee.Employee
}

func (s *store) seed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSerial = 17
	t := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	hire := t(2024, time.January, 15)
	dob := t(1990, time.March, 10)
	s.items = []employee.Employee{
		{
			ID: "000002", LegalName: "OU, HONGCHENG", Position: "",
			PayFrequency: "Semi-Monthly", Status: employee.StatusActive,
			Category: employee.Permanent, SalaryType: employee.SalaryTimeBased,
			HireDate: hire, Province: payroll.BC, Email: "ou@example.com",
			DateOfBirth: &dob, PayRate: 18, PayRateUnit: "Hourly", PaysPerYear: 24,
			MemberType: employee.MemberEmployee,
		},
		{
			ID: "000016", LegalName: "SAWAENGPON, JIRAYU", Position: "",
			PayFrequency: "Semi-Monthly", Status: employee.StatusActive,
			Category: employee.Permanent, SalaryType: employee.SalaryTimeBased,
			HireDate: t(2023, time.June, 1), Province: payroll.ON,
			PayRate: 17.42, PayRateUnit: "Hourly", PaysPerYear: 24,
			MemberType: employee.MemberEmployee,
		},
	}
}

func (s *store) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]employee.Employee, len(s.items))
	copy(out, s.items)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *store) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in employee.Employee
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if in.LegalName == "" {
		http.Error(w, "legalName required", http.StatusBadRequest)
		return
	}
	if !employee.ValidateContact(in.Email, in.Mobile) {
		http.Error(w, "email or mobile required", http.StatusBadRequest)
		return
	}
	if in.SIN != "" && !employee.ValidateSIN(in.SIN) {
		http.Error(w, "invalid SIN", http.StatusBadRequest)
		return
	}
	if in.Status == "" {
		in.Status = employee.StatusActive
	}
	if in.HireDate.IsZero() {
		in.HireDate = time.Now().Truncate(24 * time.Hour)
	}
	s.mu.Lock()
	s.nextSerial++
	in.ID = fmt.Sprintf("%06d", s.nextSerial)
	s.items = append(s.items, in)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(in)
}

func handleExportRemittanceSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := reportexport.RenderRemittanceSample(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleExportPayslipSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := reportexport.RenderPayslipSample(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func webRoot() string {
	if d := os.Getenv("PRWORKS_WEB"); d != "" {
		return d
	}
	if d := os.Getenv("WEB_DIR"); d != "" {
		return d
	}
	return "web"
}

func listenAddr() string {
	if a := os.Getenv("LISTEN_ADDR"); a != "" {
		return a
	}
	return ":8080"
}

func main() {
	root := webRoot()
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		log.Fatalf("static web directory not found or not a directory: %q (set PRWORKS_WEB or WEB_DIR, or run from repo root)", root)
	}

	s := &store{}
	s.seed()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/employees", s.handleList)
	mux.HandleFunc("POST /api/employees", s.handleCreate)
	mux.HandleFunc("GET /export/remittance/sample", handleExportRemittanceSample)
	mux.HandleFunc("GET /export/payslip/sample", handleExportPayslipSample)
	mux.Handle("/", http.FileServer(http.Dir(root)))

	addr := listenAddr()
	log.Printf("payroll-web listening on %s (static %q + /api/employees + /export/*/sample)", addr, root)
	log.Fatal(http.ListenAndServe(addr, mux))
}
