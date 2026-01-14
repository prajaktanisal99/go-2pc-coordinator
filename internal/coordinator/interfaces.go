package coordinator

import "context"

type Participant interface {
    Prepare(ctx context.Context, txID string) error
    Commit(ctx context.Context, txID string) error
    Rollback(ctx context.Context, txID string) error
}

type LogStore interface {
    UpdateState(ctx context.Context, txID string, state string) error
    CheckStarted(ctx context.Context, txID string) (bool, error)
    GetPending(ctx context.Context) ([]TransactionRecord, error)
}

type TransactionRecord struct {
    ID    string
    State string
}