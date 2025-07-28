package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/entities"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/repositories"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/workers"
)

type PaymentCreateHandler struct {
	paymentRep  *repositories.PaymentRepository
	procService *service.ProcessorService
	dlq         *workers.DLQ
}

func NewPaymentCreateHandler(paymentRep *repositories.PaymentRepository, procService *service.ProcessorService, dlq *workers.DLQ) *PaymentCreateHandler {
	return &PaymentCreateHandler{paymentRep, procService, dlq}
}

func (p *PaymentCreateHandler) Handle() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var processorUsed string = service.ProcessorDefault
		var payment entities.Payment
		payment.RequestedAt = time.Now().UTC()

		err := json.NewDecoder(r.Body).Decode(&payment)
		if err != nil {
			slog.Error("Failed to decode body", "err", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		logger := slog.With("correlationId", payment.CorrelationId)

		paymentJSONBytes, _ := json.Marshal(&payment)

		// rep, err := p.paymentRep.Start()
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// _, err = rep.Create(&payment, processorUsed)
		// if err != nil {
		// 	logger.Error("Error inserting to database:", "err", err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		_, resStatus, err := p.procService.MakeRequestDefault(http.MethodPost, "/payments", bytes.NewReader(paymentJSONBytes), 0)
		if err != nil || resStatus > 399 {
			logger.Error("Error calling default processor, sending to DQL", "err", err, "resStatus", resStatus)
			failureReason := workers.FailureProcError
			if resStatus >= 399 {
				failureReason = workers.FailureProcTimeout
			}

			p.dlq.PushToQueue(&entities.PaymentRetry{
				P:                 &payment,
				FailureCount:      1,
				LastProcessorUsed: service.ProcessorDefault,
				LastFailureReason: failureReason,
			})

			// rep.Cancel()
		}

		p.paymentRep.SavePayment(&payment, processorUsed)

		// rep.Finish()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})
}
