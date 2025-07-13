package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
)

type PaymentGetSummaryHandler struct {
	db *database.Db
}

type ProcessorSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummaryOutput struct {
	Default  ProcessorSummary `json:"default"`
	Fallback ProcessorSummary `json:"fallback"`
}

func (p PaymentSummaryOutput) String() string {
	return fmt.Sprintf("Default.TotalRequests: %v | Default.TotalAmount: %v | Fallback.TotalRequests: %v | Fallback.TotalAmount: %v", p.Default.TotalRequests, p.Default.TotalAmount, p.Fallback.TotalRequests, p.Fallback.TotalAmount)
}

func NewPaymentGetSummaryHandler(db *database.Db) *PaymentGetSummaryHandler {
	return &PaymentGetSummaryHandler{db}
}

func (p *PaymentGetSummaryHandler) Handle() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sqlQuery string
		var sqlQueryParams []any

		qsFrom := r.URL.Query().Get("from")
		qsTo := r.URL.Query().Get("to")

		if qsFrom != "" || qsTo != "" {
			sqlQueryParams = []any{qsFrom, qsTo}
			sqlQuery = `
				SELECT
				p.processor_used AS processor,
				COUNT(*) AS total_requests,
				SUM(p.amount)/100 AS total_amount
				FROM payments p
				WHERE p.processor_used IS NOT NULL AND p.requested_at BETWEEN $1 AND $2
				GROUP BY p.processor_used;
			`
		} else {
			sqlQuery = `
				SELECT
				p.processor_used AS processor,
				COUNT(*) AS total_requests,
				SUM(p.amount)/100 AS total_amount
				FROM payments p
				WHERE p.processor_used IS NOT NULL
				GROUP BY p.processor_used;
			`
		}

		res, err := p.db.Conn.Query(r.Context(), sqlQuery, sqlQueryParams...)
		if err != nil {
			slog.Error("Error reading summary rom database:", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		var summary PaymentSummaryOutput
		var processor string
		var totalRequests int
		var totalAmount float64

		for res.Next() {
			err = res.Scan(&processor, &totalRequests, &totalAmount)
			if err != nil {
				slog.Error(fmt.Sprintf("Error scanning columns from database: %v", err))
				return
			}

			if processor == service.ProcessorDefault {
				summary.Default = ProcessorSummary{totalRequests, totalAmount}
			} else {
				summary.Fallback = ProcessorSummary{totalRequests, totalAmount}
			}
		}

		slog.Info("Summary read", "summary", summary.String(), "qsFrom", qsFrom, "qsTo", qsTo)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&summary)
	})
}
