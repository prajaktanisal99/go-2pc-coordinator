package grpc

import (
	"context"

	"github.com/prajaktanisal99/go-2pc/api/proto"
	"github.com/prajaktanisal99/go-2pc/internal/coordinator"
)

type ParticipantServer struct {
	proto.UnimplementedParticipantServiceServer
	logic coordinator.Participant
}

func NewParticipantServer(logic coordinator.Participant) *ParticipantServer {
	return &ParticipantServer{logic: logic}
}

func (s *ParticipantServer) Prepare(ctx context.Context, req *proto.PrepareRequest) (*proto.PrepareResponse, error) {
	err := s.logic.Prepare(ctx, req.TxId)
	if err != nil {
		return &proto.PrepareResponse{Success: false, ErrorMessage: err.Error()}, nil
	}
	return &proto.PrepareResponse{Success: true}, nil
}

func (s *ParticipantServer) Commit(ctx context.Context, req *proto.CommitRequest) (*proto.CommitResponse, error) {
	err := s.logic.Commit(ctx, req.TxId)
	if err != nil {
		return &proto.CommitResponse{Success: false}, nil
	}
	return &proto.CommitResponse{Success: true}, nil
}

func (s *ParticipantServer) Rollback(ctx context.Context, req *proto.RollbackRequest) (*proto.RollbackResponse, error) {
	err := s.logic.Rollback(ctx, req.TxId)
	if err != nil {
		return &proto.RollbackResponse{Success: false}, nil
	}
	return &proto.RollbackResponse{Success: true}, nil
}
