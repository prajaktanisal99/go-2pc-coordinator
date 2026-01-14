package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	"github.com/prajaktanisal99/go-2pc/internal/coordinator"
	"github.com/prajaktanisal99/go-2pc/internal/participants"
	"github.com/prajaktanisal99/go-2pc/internal/repository"
)

func main() {
	// 1. Initialize Connections
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	db, err := sql.Open("pgx", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	// 2. Initialize Components
	ls := repository.NewRedisLogStore(rdb)

	// We use two participants to demonstrate the 'Distributed' part of 2PC
	pgPart := participants.NewPostgresParticipant(db)
	rdPart := participants.NewRedisParticipant(rdb)

	tm := coordinator.NewTransactionManager(ls, []coordinator.Participant{pgPart, rdPart})

	// 3. Setup Graceful Shutdown Context
	// This context will be canceled if you press Ctrl+C or the system kills the process
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Start the Recovery Worker in a separate Goroutine
	go tm.StartRecoveryWorker(ctx)

	fmt.Println("--------------------------------------------------")
	fmt.Println("🚀 Distributed Transaction Coordinator is running")
	fmt.Println("--------------------------------------------------")

	// 5. Run a Simulation
	// txID := fmt.Sprintf("tx-%d", time.Now().Unix())

	txID := "tx-999" // Fixed ID for easier testing with Recovery Worker
	// CHAOS SELECTION:
	// Use "" for a perfect success run.
	// Use "CRASH_AFTER_PREPARE" to test the Recovery Worker.
	chaosMode := ""

	fmt.Printf("📝 Executing Transaction: %s (Mode: %s)\n", txID, chaosMode)

	err = tm.Execute(ctx, txID, chaosMode)
	if err != nil {
		fmt.Printf("⚠️  Transaction Flow Interrupted: %v\n", err)
	} else {
		fmt.Println("✅ Transaction Finished Successfully!")
	}

	// 6. Keep main alive to observe the Recovery Worker
	if chaosMode != "CRASH_AFTER_PREPARE" {
		fmt.Println("\nWaiting for signals... (Press Ctrl+C to exit)")
		<-ctx.Done()
	}

	fmt.Println("👋 Shutting down gracefully...")
}
