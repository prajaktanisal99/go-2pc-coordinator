package participants

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresParticipant struct {
	DB *sql.DB
}

func NewPostgresParticipant(db *sql.DB) *PostgresParticipant {
	return &PostgresParticipant{DB: db}
}

// Prepare executes the business logic and moves the transaction to a "Prepared" state.
func (p *PostgresParticipant) Prepare(ctx context.Context, txID string) error {
	// 1. Start a standard SQL transaction
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pg begin tx failed: %w", err)
	}

	// 2. Execute Business Logic (e.g., deducting money)
	// Note: These changes are NOT visible to other users yet.
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - 100 WHERE id = 1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("pg business logic failed: %w", err)
	}

	// 3. The Magic Command: PREPARE TRANSACTION
	// This detaches the transaction from the current session and saves it to disk.
	// After this, the 'tx' object is no longer usable.
	prepareSQL := fmt.Sprintf("PREPARE TRANSACTION '%s'", txID)
	_, err = tx.ExecContext(ctx, prepareSQL)
	if err != nil {
		// If PREPARE fails, we don't need to Rollback; the session will clean it up.
		return fmt.Errorf("pg prepare failed: %w", err)
	}

	return nil
}

// Commit finalizes the prepared transaction using the global txID.
func (p *PostgresParticipant) Commit(ctx context.Context, txID string) error {
	// We use p.DB directly, not the tx object, because the transaction
	// is now managed globally by Postgres.
	query := fmt.Sprintf("COMMIT PREPARED '%s'", txID)
	_, err := p.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("pg commit prepared failed: %w", err)
	}
	return nil
}

// Rollback cancels the prepared transaction using the global txID.
func (p *PostgresParticipant) Rollback(ctx context.Context, txID string) error {
	query := fmt.Sprintf("ROLLBACK PREPARED '%s'", txID)
	_, err := p.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("pg rollback prepared failed: %w", err)
	}
	return nil
}
