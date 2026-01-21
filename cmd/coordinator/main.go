package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/prajaktanisal99/go-2pc/api/proto"
	"github.com/prajaktanisal99/go-2pc/internal/coordinator"
	internalGrpc "github.com/prajaktanisal99/go-2pc/internal/grpc"
	"github.com/prajaktanisal99/go-2pc/internal/repository"
)

func main() {
	// 1. Establish gRPC Connections to Participants
	// Use insecure credentials for local development
	pgConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Postgres service: %v", err)
	}
	defer pgConn.Close()

	rdConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Redis service: %v", err)
	}
	defer rdConn.Close()

	// 2. Create gRPC Clients from the generated proto code
	pgServiceClient := proto.NewParticipantServiceClient(pgConn)
	rdServiceClient := proto.NewParticipantServiceClient(rdConn)

	// 3. Wrap gRPC clients in our Adapter (so they fit the Participant interface)
	participantsList := []coordinator.Participant{
		internalGrpc.NewGrpcParticipantClient(pgServiceClient, "Postgres-Service"),
		internalGrpc.NewGrpcParticipantClient(rdServiceClient, "Redis-Service"),
	}

	// 4. Initialize Local LogStore (Coordinator needs its own Redis for the WAL)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ls := repository.NewRedisLogStore(rdb)

	// 5. Initialize Transaction Manager
	tm := coordinator.NewTransactionManager(ls, participantsList)

	// 6. Setup Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Recovery Worker to heal orphaned transactions
	go tm.StartRecoveryWorker(ctx)

	fmt.Println("--------------------------------------------------")
	fmt.Println("🚀 G-RPC Distributed Coordinator is running")
	fmt.Println("--------------------------------------------------")

	// 7. Simulation Run
	txID := fmt.Sprintf("tx-grpc-%d", time.Now().Unix())
	chaosMode := "" // Set to "CRASH_AFTER_PREPARE" to test recovery

	fmt.Printf("📝 Requesting Transaction: %s\n", txID)

	err = tm.Execute(ctx, txID, chaosMode)
	if err != nil {
		fmt.Printf("⚠️  Transaction Flow Interrupted: %v\n", err)
	} else {
		fmt.Println("✅ Transaction Finished Successfully via gRPC!")
	}

	// Wait for shutdown signal
	fmt.Println("\nWaiting for signals... (Press Ctrl+C to exit)")
	<-ctx.Done()
	fmt.Println("👋 Shutting down gracefully...")
}
