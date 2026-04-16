package writecheque

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

//go:embed templates/cheque.tmpl.html
var templateFS embed.FS

var chequeTmpl = template.Must(
	template.New("cheque").ParseFS(templateFS, "templates/cheque.tmpl.html"),
)

// Handler holds the dependencies for writecheque HTTP handlers.
type Handler struct {
	Store Store
}

// HandlePreview renders the cheque template as HTML (used in iframe preview).
// GET /api/writecheque/preview?bank_id=...&payee=...&amount=...&memo=...&date=...&currency=...&check_no=...
func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	p := paramsFromRequest(r)
	data, err := Build(h.Store, p)
	if err != nil {
		http.Error(w, "build cheque: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if err := chequeTmpl.ExecuteTemplate(w, "cheque.tmpl.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandlePrint renders the cheque template with print chrome stripped out —
// the browser will open the print dialog immediately via JS.
// GET /writecheque/print?...   (same params as preview)
// This endpoint is also the Payroll integration entry point.
func (h *Handler) HandlePrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	p := paramsFromRequest(r)
	data, err := Build(h.Store, p)
	if err != nil {
		http.Error(w, "build cheque: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if err := chequeTmpl.ExecuteTemplate(w, "cheque.tmpl.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// paramsFromRequest extracts ChequeParams from URL query string.
func paramsFromRequest(r *http.Request) Params {
	q := r.URL.Query()
	amount, _ := strconv.ParseFloat(strings.TrimSpace(q.Get("amount")), 64)
	return Params{
		BankID:   strings.TrimSpace(q.Get("bank_id")),
		Payee:    strings.TrimSpace(q.Get("payee")),
		Amount:   amount,
		Currency: strings.TrimSpace(q.Get("currency")),
		Memo:     strings.TrimSpace(q.Get("memo")),
		Date:     strings.TrimSpace(q.Get("date")),
		CheckNo:  strings.TrimSpace(q.Get("check_no")),
	}
}
