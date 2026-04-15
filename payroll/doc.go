// Package payroll implements Canadian payroll statutory deductions per CRA T4127-JAN (see Edition).
//
// Tax year: brackets, CPP/EI caps, and personal amounts change each calendar year. Set
// [PayPeriodInput.TaxYear] / [Option2PayPeriodInput.TaxYear] to the withholding year, or pass a full
// [RateBook] from your database ([ResolveRateBook]). Years not embedded in this module return
// [ErrNoEmbeddedRateBook] unless you supply Rates.
//
// Formula vs data: [RateBook] carries table values. If CRA changes formula structure in a future
// T4127 edition, you may need a new module version or additional code paths—not only new JSON.
//
// Stability: keep public APIs stable; add embedded years by extending [RateBookForTaxYear] and
// shipping new default data.
//
// Option 2 combined tax T is defined by CRA; allocation to federal vs provincial withholding lines
// is not prescribed—use [Option2TaxSplitConfig] (and persist the policy in your datastore).
//
// This is not legal or accounting advice; validate against PDOC and current CRA publications.
package payroll
