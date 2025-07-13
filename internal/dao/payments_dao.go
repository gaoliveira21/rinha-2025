package dao

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentsDAO struct {
	pool *pgxpool.Pool
}

func NewPaymentsDAO(pool *pgxpool.Pool) *PaymentsDAO {
	return &PaymentsDAO{pool: pool}
}

type PaymentSummaryInput struct {
	From string
	To   string
}

type PaymentSummary struct {
	TotalRequests int64   `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummaryOutput struct {
	Default  *PaymentSummary `json:"default"`
	Fallback *PaymentSummary `json:"fallback"`
}

func (r *PaymentsDAO) GetPaymentSummary(in *PaymentSummaryInput) (*PaymentSummaryOutput, error) {
	conn, err := r.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	defaultSum := &PaymentSummary{}
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_default
		WHERE requested_at >= $1 AND requested_at <= $2
	`, in.From, in.To).Scan(&defaultSum.TotalRequests, &defaultSum.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment summary: %w", err)
	}

	fallbackSum := &PaymentSummary{}
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_fallback
		WHERE requested_at >= $1 AND requested_at <= $2
	`, in.From, in.To).Scan(&fallbackSum.TotalRequests, &fallbackSum.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment summary: %w", err)
	}

	out := &PaymentSummaryOutput{
		Default:  defaultSum,
		Fallback: fallbackSum,
	}

	return out, nil
}
