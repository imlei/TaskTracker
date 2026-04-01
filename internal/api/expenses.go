package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"tasktracker/internal/models"
	"tasktracker/internal/store"
)

var expenseAccountCodeRx = regexp.MustCompile(`^5[0-9]{3}$`)

func validateExpenseAccountCode(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("accountCode is required (5XXX)")
	}
	if !expenseAccountCodeRx.MatchString(s) {
		return errors.New("accountCode must be 4 digits starting with 5 (5XXX)")
	}
	return nil
}

func (s *Server) handleExpenses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := s.Store.ListExpenses()
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var e models.Expense
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateExpenseAccountCode(e.AccountCode); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Store.CreateExpense(e)
		if errors.Is(err, store.ErrExpenseTaskNotFound) {
			http.Error(w, "task not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleExpenseByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/expenses/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		e, err := s.Store.GetExpense(id)
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, e)
	case http.MethodPut:
		if _, err := s.Store.GetExpense(id); errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		var e models.Expense
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateExpenseAccountCode(e.AccountCode); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updated, err := s.Store.UpdateExpense(id, e)
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, store.ErrExpenseTaskNotFound) {
			http.Error(w, "task not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
