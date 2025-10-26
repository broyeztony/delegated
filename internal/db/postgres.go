package db

import (
	"context"
	"fmt"

	"github.com/broyeztony/delegated/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetMaxID returns the count and max id from the delegations table
func GetMaxID(ctx context.Context, pool *pgxpool.Pool) (count int64, maxID int64, err error) {
	err = pool.QueryRow(ctx, "SELECT COUNT(*), COALESCE(MAX(id), 0) FROM delegations").Scan(&count, &maxID)
	return
}

// BulkInsertDelegations inserts delegations with ON CONFLICT handling using bulk insert
func BulkInsertDelegations(ctx context.Context, pool *pgxpool.Pool, delegations []models.Delegation) error {
	if len(delegations) == 0 {
		return nil
	}

	query := `
		INSERT INTO delegations (id, delegator, timestamp, amount, level)
		VALUES (@id, @delegator, @timestamp, @amount, @level)
		ON CONFLICT (id) DO NOTHING`

	// Use a transaction for atomicity
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, d := range delegations {
		args := pgx.NamedArgs{
			"id":        d.ID,
			"delegator": d.Delegator,
			"timestamp": d.Timestamp,
			"amount":    d.Amount,
			"level":     d.Level,
		}
		batch.Queue(query, args)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// Consume all results to complete the batch
	for i := 0; i < len(delegations); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert delegation in batch: %w", err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	// Commit the transaction
	return tx.Commit(ctx)
}
