package writecheque

import (
	"fmt"
	"strings"
	"time"

	"simpletask/internal/models"
	"simpletask/internal/store"
)

// Params are the query parameters for building a ChequeData.
// They can come from URL query string (manual entry) or be assembled
// from a payroll period (see FromPeriod).
type Params struct {
	BankID   string
	Payee    string
	Amount   float64
	Currency string
	Memo     string
	Date     string // ISO: 2006-01-02; empty → today
	CheckNo  string // overrides the bank's stored cheque number
}

// Store is the subset of store.Store methods needed by this service.
type Store interface {
	GetBankAccount(id string) (models.BankAccount, error)
	GetDefaultBankAccount() (models.BankAccount, error)
	GetPayrollCompanyName(id string) (string, error)
	GetAppSettingsCompanyName() string
}

// Build assembles a ChequeData ready for template rendering.
func Build(st Store, p Params) (ChequeData, error) {
	var bank models.BankAccount
	var err error
	if strings.TrimSpace(p.BankID) != "" {
		bank, err = st.GetBankAccount(strings.TrimSpace(p.BankID))
	} else {
		bank, err = st.GetDefaultBankAccount()
	}
	if err != nil && err != store.ErrNotFound {
		return ChequeData{}, fmt.Errorf("bank account: %w", err)
	}

	// Company name: prefer bank-linked payroll company, then global settings
	companyName := ""
	if strings.TrimSpace(bank.CompanyID) != "" {
		companyName, _ = st.GetPayrollCompanyName(bank.CompanyID)
	}
	if companyName == "" {
		companyName = st.GetAppSettingsCompanyName()
	}

	checkNo := strings.TrimSpace(p.CheckNo)
	if checkNo == "" {
		checkNo = bank.BankChequeNumber
	}
	if checkNo == "" {
		checkNo = "000001"
	}

	currency := strings.ToUpper(strings.TrimSpace(p.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(bank.DefaultChequeCurrency))
	}
	if currency == "" {
		currency = "CAD"
	}

	dateStr := strings.TrimSpace(p.Date)
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	displayDate := formatDate(dateStr)

	micrLine := BuildMICR(
		bank.MICRCountry,
		bank.BankInstitution,
		bank.BankTransit,
		bank.BankRoutingABA,
		bank.BankAccount,
		bank.BankIBAN,
		bank.MICRLineOverride,
		checkNo,
	)

	return ChequeData{
		CompanyName:    companyName,
		CheckNo:        checkNo,
		Date:           displayDate,
		Payee:          strings.TrimSpace(p.Payee),
		AmountBox:      FormatAmountBox(p.Amount, currency),
		AmountWords:    AmountToWords(p.Amount),
		Memo:           strings.TrimSpace(p.Memo),
		MICRLine:       micrLine,
		Currency:       currency,
		BankName:       bank.BankName,
		BankAddress:    bank.BankAddress,
		BankCity:       bank.BankCity,
		BankProvince:   bank.BankProvince,
		BankPostalCode: bank.BankPostalCode,
	}, nil
}

// formatDate converts "2006-01-02" to "2006/01/02" for cheque display.
func formatDate(iso string) string {
	iso = strings.TrimSpace(iso)
	if len(iso) >= 10 && iso[4] == '-' && iso[7] == '-' {
		return iso[:4] + "/" + iso[5:7] + "/" + iso[8:10]
	}
	return iso
}
