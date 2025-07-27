package repositories

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/entities"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
)

type PaymentRepository struct {
	db *database.Db
	tx pgx.Tx
}

func NewPaymentRepository(db *database.Db) *PaymentRepository {
	return &PaymentRepository{db, nil}
}

func (r *PaymentRepository) Start() (*PaymentRepository, error) {
	tx, err := r.db.Conn.Begin(context.TODO())
	if err != nil {
		return nil, err
	}

	return &PaymentRepository{r.db, tx}, nil
}

func (r *PaymentRepository) Finish() {
	if r.tx == nil {
		return
	}

	r.tx.Commit(context.TODO())
	r.tx = nil
}

func (r *PaymentRepository) Cancel() {
	if r.tx == nil {
		return
	}

	r.tx.Rollback(context.TODO())
	r.tx = nil
}

func (r *PaymentRepository) Create(p *entities.Payment, processorUsed string) (pgconn.CommandTag, error) {
	query := "INSERT INTO payments (correlation_id, amount, processor_used, requested_at) VALUES($1, $2, $3, $4)"
	args := []any{p.CorrelationId, int(math.Round(p.Amount * 100)), processorUsed, p.RequestedAt.Format("2006-01-02T15:04:05.000Z")}

	if r.tx != nil {
		return r.tx.Exec(context.TODO(), query, args...)

	}

	return r.db.Conn.Exec(context.TODO(), query, args)
}
