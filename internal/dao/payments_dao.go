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

type PaymentSummary struct {
	TotalRequests int64   `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryOutput struct {
	Default  *PaymentSummary `json:"default"`
	Fallback *PaymentSummary `json:"fallback"`
}

func (r *PaymentsDAO) InsertPaymentDefault(id string, amount float64) error {
	conn, err := r.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `
		INSERT INTO payments_default (id, amount, requested_at)
		VALUES ($1, $2, NOW())
	`, id, amount)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	return nil
}

func (r *PaymentsDAO) InsertPaymentFallback(id string, amount float64) error {
	conn, err := r.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `
		INSERT INTO payments_fallback (id, amount, requested_at)
		VALUES ($1, $2, NOW())
	`, id, amount)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	return nil
}

func (r *PaymentsDAO) GetPaymentsSummary(from string, to string) (*PaymentsSummaryOutput, error) {
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
	`, from, to).Scan(&defaultSum.TotalRequests, &defaultSum.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment summary: %w", err)
	}

	fallbackSum := &PaymentSummary{}
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_fallback
		WHERE requested_at >= $1 AND requested_at <= $2
	`, from, to).Scan(&fallbackSum.TotalRequests, &fallbackSum.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment summary: %w", err)
	}

	out := &PaymentsSummaryOutput{
		Default:  defaultSum,
		Fallback: fallbackSum,
	}

	return out, nil
}
