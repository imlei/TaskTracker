package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const frankfurterAPI = "https://api.frankfurter.dev"
const exchangeRatesMaxWeekdays = 31
const frankfurterHTTPTimeout = 25 * time.Second

type frankfurterRateRow struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

type frankfurterCurrencyRow struct {
	ISOCode string `json:"iso_code"`
	Name    string `json:"name"`
}

func lastNWeekdaysBeforeOrOn(anchor time.Time, n int, loc *time.Location) []string {
	if n <= 0 {
		return nil
	}
	d := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, loc)
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, -1)
	}
	var rev []string
	cur := d
	for len(rev) < n {
		if cur.Weekday() != time.Saturday && cur.Weekday() != time.Sunday {
			rev = append(rev, cur.Format("2006-01-02"))
		}
		cur = cur.AddDate(0, 0, -1)
		if cur.Before(time.Date(1999, 1, 1, 0, 0, 0, 0, loc)) {
			break
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

func weekdaysBetweenInclusive(fromStr, toStr string, loc *time.Location) ([]string, error) {
	from, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
	if err != nil {
		return nil, fmt.Errorf("invalid from date")
	}
	to, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
	if err != nil {
		return nil, fmt.Errorf("invalid to date")
	}
	if to.Before(from) {
		from, to = to, from
	}
	var out []string
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		out = append(out, d.Format("2006-01-02"))
		if len(out) > exchangeRatesMaxWeekdays {
			return nil, fmt.Errorf("range exceeds %d weekdays", exchangeRatesMaxWeekdays)
		}
	}
	return out, nil
}

func httpGetJSON(target string, v any) error {
	client := &http.Client{Timeout: frankfurterHTTPTimeout}
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TaskTracker/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<22))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("frankfurter: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, v)
}

func fetchFrankfurterRates(requestedDate, base string) (map[string]float64, error) {
	u, err := url.Parse(frankfurterAPI + "/v2/rates")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("date", requestedDate)
	q.Set("base", base)
	u.RawQuery = q.Encode()
	var rows []frankfurterRateRow
	if err := httpGetJSON(u.String(), &rows); err != nil {
		return nil, err
	}
	out := make(map[string]float64)
	for _, r := range rows {
		qc := strings.ToUpper(strings.TrimSpace(r.Quote))
		if qc == "" || strings.EqualFold(qc, base) {
			continue
		}
		out[qc] = r.Rate
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no rates returned (check base currency code)")
	}
	return out, nil
}

func fetchFrankfurterCurrencyNames() (map[string]string, error) {
	var rows []frankfurterCurrencyRow
	if err := httpGetJSON(frankfurterAPI+"/v2/currencies", &rows); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, r := range rows {
		c := strings.ToUpper(strings.TrimSpace(r.ISOCode))
		if c == "" {
			continue
		}
		out[c] = strings.TrimSpace(r.Name)
	}
	return out, nil
}

func (s *Server) handleExchangeRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	st, err := s.Store.GetSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	base := strings.ToUpper(strings.TrimSpace(st.BaseCurrency))
	if base == "" {
		base = "CAD"
	}
	loc := time.Local
	fromStr := strings.TrimSpace(r.URL.Query().Get("from"))
	toStr := strings.TrimSpace(r.URL.Query().Get("to"))
	qFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	var dates []string
	switch {
	case fromStr == "" && toStr == "":
		yesterday := time.Now().In(loc).AddDate(0, 0, -1)
		dates = lastNWeekdaysBeforeOrOn(yesterday, 5, loc)
	case fromStr != "" && toStr != "":
		dates, err = weekdaysBetweenInclusive(fromStr, toStr, loc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "specify both from and to (YYYY-MM-DD), or omit both for default window", http.StatusBadRequest)
		return
	}
	if len(dates) == 0 {
		http.Error(w, "no weekdays in range", http.StatusBadRequest)
		return
	}

	for _, d := range dates {
		n, err := s.Store.CountExchangeRatesForDate(d, base)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if n > 0 {
			continue
		}
		rates, err := fetchFrankfurterRates(d, base)
		if err != nil {
			http.Error(w, "fetch exchange rates: "+err.Error(), http.StatusBadGateway)
			return
		}
		fetchedAt := time.Now().UTC().Format(time.RFC3339)
		if err := s.Store.ReplaceExchangeRatesForDate(d, base, rates, fetchedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	grid, err := s.Store.GetExchangeRatesForDates(base, dates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	names, err := fetchFrankfurterCurrencyNames()
	if err != nil {
		http.Error(w, "fetch currency names: "+err.Error(), http.StatusBadGateway)
		return
	}

	quoteSet := map[string]struct{}{}
	for _, dm := range grid {
		for code := range dm {
			if strings.EqualFold(code, base) {
				continue
			}
			quoteSet[strings.ToUpper(code)] = struct{}{}
		}
	}
	type rowOut struct {
		Code  string             `json:"code"`
		Name  string             `json:"name"`
		Rates map[string]float64 `json:"rates"`
	}
	rows := make([]rowOut, 0, len(quoteSet))
	for code := range quoteSet {
		name := names[code]
		if name == "" {
			name = code
		}
		if qFilter != "" {
			if !strings.Contains(strings.ToLower(code), qFilter) && !strings.Contains(strings.ToLower(name), qFilter) {
				continue
			}
		}
		rates := make(map[string]float64)
		for _, dt := range dates {
			if dm, ok := grid[dt]; ok {
				if v, ok := dm[code]; ok {
					rates[dt] = v
				}
			}
		}
		rows = append(rows, rowOut{Code: code, Name: name, Rates: rates})
	}
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"base":     base,
		"dates":    dates,
		"rows":     rows,
		"source":   frankfurterAPI,
		"sourceDoc": "https://frankfurter.dev/",
	})
}
