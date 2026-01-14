package coordinator

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

type TransactionManager struct {
	logStore     LogStore
	participants []Participant
}

func NewTransactionManager(ls LogStore, ps []Participant) *TransactionManager {
	return &TransactionManager{
		logStore:     ls,
		participants: ps,
	}
}

func (tm *TransactionManager) Execute(ctx context.Context, txID string, chaosMode string) error {
	// 1. Idempotency Check
	isDone, err := tm.logStore.CheckStarted(ctx, txID)
	if err != nil {
		return err
	}
	if isDone {
		return nil
	}

	// 2. Log Start
	if err := tm.logStore.UpdateState(ctx, txID, "START"); err != nil {
		return err
	}

	// 3. PHASE 1: PREPARE
	prepareCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tm.coordinatePhase1(prepareCtx, txID); err != nil {
		tm.coordinatePhase2(ctx, txID, false)
		tm.logStore.UpdateState(ctx, txID, "ABORTED")
		return fmt.Errorf("prepare failed: %w", err)
	}

	// 4. LOG DECISION (COMMIT POINT)
	if err := tm.logStore.UpdateState(ctx, txID, "PREPARED"); err != nil {
		return err
	}

	// Chaos Simulation
	if chaosMode == "CRASH_AFTER_PREPARE" {
		log.Println("!!! CRASHING FOR TEST !!!")
		os.Exit(1)
	}

	// 5. PHASE 2: COMMIT
	if err := tm.coordinatePhase2(ctx, txID, true); err != nil {
		log.Printf("[PHASE 2] Partial failure: %v", err)
	}

	return tm.logStore.UpdateState(ctx, txID, "COMMITTED")
}

func (tm *TransactionManager) coordinatePhase1(ctx context.Context, txID string) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, p := range tm.participants {
		p := p
		g.Go(func() error { return p.Prepare(ctx, txID) })
	}
	return g.Wait()
}

func (tm *TransactionManager) coordinatePhase2(ctx context.Context, txID string, commit bool) error {
	for _, p := range tm.participants {
		if commit {
			p.Commit(ctx, txID)
		} else {
			p.Rollback(ctx, txID)
		}
	}
	return nil
}
