package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
)

type PaymentCreateHandler struct {
	db          *database.Db
	procService *service.ProcessorService
}

type PaymentCreateInput struct {
	CorrelationId string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

func NewPaymentCreateHandler(db *database.Db, procService *service.ProcessorService) *PaymentCreateHandler {
	return &PaymentCreateHandler{db, procService}
}

func (p *PaymentCreateHandler) Handle() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var processorUsed string = service.ProcessorDefault
		var payment PaymentCreateInput
		payment.RequestedAt = time.Now().UTC()

		err := json.NewDecoder(r.Body).Decode(&payment)
		if err != nil {
			slog.Error("Failed to decode body", "err", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		logger := slog.With("correlationId", payment.CorrelationId)

		paymentJSONBytes, _ := json.Marshal(&payment)

		// logger.Info("payment received:",
		// 	slog.Group("payment",
		// 		"correlationId", payment.CorrelationId,
		// 		"amount", payment.Amount,
		// 		"requestedAt", payment.RequestedAt,
		// 	),
		// )

		_, resStatus, err := p.procService.MakeRequestDefault(http.MethodPost, "/payments", bytes.NewReader(paymentJSONBytes))
		if err != nil || resStatus > 399 {
			// logger.Error("Error calling default processor, retrying with fallback", "err", err, "resStatus", resStatus)

			processorUsed = service.ProcessorFallback
			_, resStatus, err = p.procService.MakeRequestFallback(http.MethodPost, "/payments", bytes.NewReader(paymentJSONBytes))
			if err != nil || resStatus > 399 {
				logger.Error("Error calling fallback processor", "err", err, "resStatus", resStatus)
				// w.WriteHeader(http.StatusUnprocessableEntity)
				// return
				processorUsed = ""
			}
		}

		_, err = p.db.Conn.Exec(context.TODO(),
			"INSERT INTO payments (correlation_id, amount, processor_used, requested_at) VALUES($1, $2, $3, $4)",
			payment.CorrelationId, int(math.Round(payment.Amount*100)), processorUsed, payment.RequestedAt.Format("2006-01-02T15:04:05.000Z"),
		)
		if err != nil {
			logger.Error("Error inserting to database:", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})
}
