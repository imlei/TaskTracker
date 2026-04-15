package payroll

import (
	"errors"
	"fmt"
)

// ErrNoEmbeddedRateBook is returned when TaxYear has no embedded tables and Rates is nil.
// Load that year’s T4127 CSV / PDF into RateBook from your database or config service.
var ErrNoEmbeddedRateBook = errors.New("payroll: no embedded RateBook for this tax year")

// EmbeddedTaxYears lists calendar years for which this module ships a full embedded [RateBook].
// Add new years when you import the next CRA T4127-JAN edition into the codebase.
var EmbeddedTaxYears = []int{2026}

// RateBookForTaxYear returns a copy of the embedded tables for the given calendar tax year
// (withholding year), e.g. pay dated in 2026 uses tax year 2026.
func RateBookForTaxYear(year int) (RateBook, error) {
	switch year {
	case 2026:
		return DefaultRateBook(), nil
	default:
		return RateBook{}, fmt.Errorf("%w: %d", ErrNoEmbeddedRateBook, year)
	}
}

// ResolveRateBook selects the effective rate book for a payroll run.
//
// If rates is non-nil, it is used as-is (your backend already chose the edition). Optionally
// ensure rates.TaxYear matches the payroll tax year in your service layer.
//
// If rates is nil and taxYear is 0, the default embedded book for the latest shipped year is used
// (currently 2026) — backward compatible with callers that omit both fields.
//
// If rates is nil and taxYear is non-zero, an embedded book is returned when available; otherwise
// [ErrNoEmbeddedRateBook].
func ResolveRateBook(taxYear int, rates *RateBook) (*RateBook, error) {
	if rates != nil {
		return rates, nil
	}
	if taxYear == 0 {
		rb := defaultRateBookData
		return &rb, nil
	}
	rb, err := RateBookForTaxYear(taxYear)
	if err != nil {
		return nil, err
	}
	out := rb
	return &out, nil
}

// rateBookOrDefault is used after ResolveRateBook at API entry; kept for internal helpers that
// already hold a resolved *RateBook.
func rateBookOrDefault(rb *RateBook) *RateBook {
	if rb != nil {
		return rb
	}
	return &defaultRateBookData
}
