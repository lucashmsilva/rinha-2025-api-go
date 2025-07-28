package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/repositories"
)

type PaymentGetSummaryHandler struct {
	paymentRep *repositories.PaymentRepository
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

func NewPaymentGetSummaryHandler(paymentRep *repositories.PaymentRepository) *PaymentGetSummaryHandler {
	return &PaymentGetSummaryHandler{paymentRep}
}

func (p *PaymentGetSummaryHandler) Handle() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		summary, err := p.paymentRep.GetSummary()
		if err != nil {
			return nil, err
		}

		slog.Info("Summary read", "summary", summary.String(), "qsFrom", qsFrom, "qsTo", qsTo)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&summary)
	})
}
