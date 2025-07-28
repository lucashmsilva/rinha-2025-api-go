package repositories

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/entities"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
)

type PaymentRepository struct {
	db *database.Db
}

func NewPaymentRepository(db *database.Db) *PaymentRepository {
	return &PaymentRepository{db}
}

func (r *PaymentRepository) SavePayment(p *entities.Payment, processorUsed string) error {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(p); err != nil {
		return err
	}

	return r.db.Conn.HSet(r.db.Ctx, fmt.Sprintf("payments:%s", processorUsed), p.CorrelationId, b.Bytes(), 0).Err()
}

func (r *PaymentRepository) GetAllPayments() ([]*entities.Payment, []*entities.Payment, error) {
	var defaultPayments, fallbackPayments []*entities.Payment
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		defaultPaymentsMap := r.db.Conn.HGetAll(r.db.Ctx, "payments:default").Val()

		for _, redisPayment := range defaultPaymentsMap {
			var payment entities.Payment

			b := bytes.NewReader([]byte(redisPayment))
			gob.NewDecoder(b).Decode(&payment)

			defaultPayments = append(defaultPayments, &payment)
		}
	}()

	go func() {
		fallbackPaymentsMap := r.db.Conn.HGetAll(r.db.Ctx, "payments:fallback").Val()

		for _, redisPayment := range fallbackPaymentsMap {
			var payment entities.Payment

			b := bytes.NewReader([]byte(redisPayment))
			gob.NewDecoder(b).Decode(&payment)

			fallbackPayments = append(fallbackPayments, &payment)
		}
	}()

	wg.Wait()

	return defaultPayments, fallbackPayments, nil
}

func (r *PaymentRepository) GetSummary() (*entities.PaymentSummary, error) {
	defaultTotalRequests, err := r.db.Conn.Get(r.db.Ctx, "summary:default:totalRequests").Int()
	if err != nil {
		return nil, err
	}

	defaultTotalAmount, err := r.db.Conn.Get(r.db.Ctx, "summary:default:totalAmount").Float64()
	if err != nil {
		return nil, err
	}

	fallbackTotalRequests, err := r.db.Conn.Get(r.db.Ctx, "summary:fallback:totalRequests").Int()
	if err != nil {
		return nil, err
	}

	fallbackTotalAmount, err := r.db.Conn.Get(r.db.Ctx, "summary:fallback:totalAmount").Float64()
	if err != nil {
		return nil, err
	}

	summary := &entities.PaymentSummary{
		Default: entities.Metrics{
			TotalRequests: defaultTotalRequests,
			TotalAmount:   defaultTotalAmount,
		},
		Fallback: entities.Metrics{
			TotalRequests: fallbackTotalRequests,
			TotalAmount:   fallbackTotalAmount,
		},
	}

	return summary, nil
}
