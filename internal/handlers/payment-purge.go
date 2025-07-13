package handlers

import (
	"log/slog"
	"net/http"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
)

type PaymentPurgeHandler struct {
	db          *database.Db
	procService *service.ProcessorService
}

func NewPaymentsPurgeHandler(db *database.Db, procService *service.ProcessorService) *PaymentPurgeHandler {
	return &PaymentPurgeHandler{db, procService}
}

func (p *PaymentPurgeHandler) Handle() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, resStatus, err := p.procService.MakeRequestDefault(http.MethodPost, "/admin/purge-payments", nil)
		if err != nil || resStatus > 399 {
			slog.Error("Error calling default processor payment purge", "err", err, "resStatus", resStatus)

			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		_, resStatus, err = p.procService.MakeRequestFallback(http.MethodPost, "/admin/purge-payments", nil)
		if err != nil || resStatus > 399 {
			slog.Error("Error calling default processor payment purge", "err", err, "resStatus", resStatus)

			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		_, err = p.db.Conn.Exec(r.Context(), "TRUNCATE TABLE payments;")
		if err != nil {
			slog.Error("Error purging database", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("payments purged")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})
}
