package redisparticipant

// redis microservice

import (
	"log"
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/prajaktanisal99/go-2pc/api/proto"
	internalGrpc "github.com/prajaktanisal99/go-2pc/internal/grpc"
	"github.com/prajaktanisal99/go-2pc/internal/participants"
)

func main() {
	// 1. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2. Initialize the Redis Logic
	rdLogic := participants.NewRedisParticipant(rdb)

	// 3. Setup gRPC Server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// We reuse the same server wrapper!
	proto.RegisterParticipantServiceServer(s, internalGrpc.NewParticipantServer(rdLogic))

	log.Println("🔴 Redis Participant Service starting on :50052...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
