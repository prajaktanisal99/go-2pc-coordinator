package postgresparticipant

// postgres microservice

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/prajaktanisal99/go-2pc/api/proto"
	internalGrpc "github.com/prajaktanisal99/go-2pc/internal/grpc"
	"github.com/prajaktanisal99/go-2pc/internal/participants"
)

func main() {
	// 1. Connect to Postgres
	db, err := sql.Open("pgx", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize the Postgres Logic (Your existing adapter)
	pgLogic := &participants.PostgresParticipant{DB: db}

	// 3. Setup gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterParticipantServiceServer(s, internalGrpc.NewParticipantServer(pgLogic))

	log.Println("🐘 Postgres Participant Service starting on :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
