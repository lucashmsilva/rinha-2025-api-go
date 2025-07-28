package repositories

import (
	"fmt"

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
	requestsKey := fmt.Sprintf("summary:%s:totalRequests", processorUsed)
	amountKey := fmt.Sprintf("summary:%s:totalAmount", processorUsed)

	err := r.db.Conn.Incr(r.db.Ctx, requestsKey).Err()
	if err != nil {
		return err
	}

	err = r.db.Conn.IncrByFloat(r.db.Ctx, amountKey, p.Amount).Err()
	if err != nil {
		return err
	}

	return nil
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
