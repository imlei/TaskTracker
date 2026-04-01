package store

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"tasktracker/internal/models"
)

var expenseCodeFormatRx = regexp.MustCompile(`^5[0-9]{3}$`)

// ValidExpenseCodeFormat 费用科目 5XXX
func ValidExpenseCodeFormat(code string) bool {
	return expenseCodeFormatRx.MatchString(strings.TrimSpace(code))
}

// ListExpenseCodeRows 合并 expense_codes 与 expenses 中出现过的 account_code，并计算指定自然年的支出合计（YTD）。
func (s *Store) ListExpenseCodeRows(year int) []models.ExpenseCodeRow {
	if year <= 0 {
		year = time.Now().Year()
	}
	yStr := strconv.Itoa(year)
	rows, err := s.db.Query(`
		SELECT code FROM (
			SELECT code FROM expense_codes
			UNION
			SELECT account_code AS code FROM expenses WHERE IFNULL(account_code,'') != ''
		) AS u ORDER BY code`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	codes := make([]string, 0)
	for rows.Next() {
		var c string
		if rows.Scan(&c) != nil {
			continue
		}
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		codes = append(codes, c)
	}
	sort.Strings(codes)
	out := make([]models.ExpenseCodeRow, 0, len(codes))
	for _, code := range codes {
		name := ""
		_ = s.db.QueryRow(`SELECT name FROM expense_codes WHERE code=?`, code).Scan(&name)
		var bal sql.NullFloat64
		_ = s.db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0) FROM expenses
			WHERE account_code = ? AND strftime('%Y', created_at) = ?`, code, yStr).Scan(&bal)
		v := 0.0
		if bal.Valid {
			v = bal.Float64
		}
		out = append(out, models.ExpenseCodeRow{
			Code:        code,
			Name:        strings.TrimSpace(name),
			BalanceYtd:  v,
			BalanceYear: year,
		})
	}
	return out
}

// UpsertExpenseCode 写入或更新科目名称（catalog）
func (s *Store) UpsertExpenseCode(code, name string) error {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if !ValidExpenseCodeFormat(code) {
		return errors.New("invalid expense code (expect 5XXX)")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
		INSERT INTO expense_codes (code, name) VALUES (?,?)
		ON CONFLICT(code) DO UPDATE SET name=excluded.name`, code, name)
	return err
}
