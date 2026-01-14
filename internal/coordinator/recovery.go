package coordinator

import (
	"context"
	"log"
	"time"
)

// StartRecoveryWorker handles the "Janitor" logic in a background loop.
func (tm *TransactionManager) StartRecoveryWorker(ctx context.Context) {
	// Run every 10 seconds to check for orphaned transactions
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("🛠️  Recovery Worker started...")

	for {
		select {
		case <-ticker.C:
			tm.processPending(ctx)
		case <-ctx.Done():
			log.Println("🛑 Stopping Recovery Worker...")
			return
		}
	}
}

func (tm *TransactionManager) processPending(ctx context.Context) {
	pending, err := tm.logStore.GetPending(ctx)
	if err != nil {
		log.Printf("[RECOVERY ERROR] Could not fetch pending: %v", err)
		return
	}

	for _, tx := range pending {
		log.Printf("[RECOVERY] Resolving TX %s (State: %s)", tx.ID, tx.State)

		switch tx.State {
		case "PREPARED":
			// Decision was made to commit before the crash
			err := tm.coordinatePhase2(ctx, tx.ID, true)
			if err == nil {
				tm.logStore.UpdateState(ctx, tx.ID, "COMMITTED")
			}
		case "START":
			// We crashed before the decision was reached; safe to rollback
			err := tm.coordinatePhase2(ctx, tx.ID, false)
			if err == nil {
				tm.logStore.UpdateState(ctx, tx.ID, "ABORTED")
			}
		}
	}
}
