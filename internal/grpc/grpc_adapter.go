package grpc

import (
	"context"
	"fmt"

	"github.com/prajaktanisal99/go-2pc/api/proto"
)

// GrpcParticipantClient implements coordinator.Participant
type GrpcParticipantClient struct {
	client proto.ParticipantServiceClient
	name   string
}

func NewGrpcParticipantClient(client proto.ParticipantServiceClient, name string) *GrpcParticipantClient {
	return &GrpcParticipantClient{client: client, name: name}
}

func (g *GrpcParticipantClient) Prepare(ctx context.Context, txID string) error {
	resp, err := g.client.Prepare(ctx, &proto.PrepareRequest{TxId: txID})
	if err != nil {
		return fmt.Errorf("rpc error [%s]: %v", g.name, err)
	}
	if !resp.Success {
		return fmt.Errorf("prepare failed [%s]: %s", g.name, resp.ErrorMessage)
	}
	return nil
}

func (g *GrpcParticipantClient) Commit(ctx context.Context, txID string) error {
	resp, err := g.client.Commit(ctx, &proto.CommitRequest{TxId: txID})
	if err != nil || !resp.Success {
		return fmt.Errorf("commit failed [%s]: %v", g.name, err)
	}
	return nil
}

func (g *GrpcParticipantClient) Rollback(ctx context.Context, txID string) error {
	resp, err := g.client.Rollback(ctx, &proto.RollbackRequest{TxId: txID})
	if err != nil || !resp.Success {
		return fmt.Errorf("rollback failed [%s]: %v", g.name, err)
	}
	return nil
}
