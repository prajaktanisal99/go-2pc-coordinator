package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/prajaktanisal99/go-2pc/internal/coordinator" // 1. Import the coordinator
	"github.com/redis/go-redis/v9"
)

const (
	activeTxSet = "active_transactions"
)

// 2. REMOVED: local TransactionRecord struct (we use the coordinator's version now)

type RedisLogStore struct {
	client *redis.Client
}

func NewRedisLogStore(client *redis.Client) *RedisLogStore {
	return &RedisLogStore{client: client}
}

func (r *RedisLogStore) UpdateState(ctx context.Context, txID string, state string) error {
	key := fmt.Sprintf("tx:%s", txID)
	pipe := r.client.Pipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"state":      state,
		"updated_at": time.Now().Format(time.RFC3339),
	})

	if state == "COMMITTED" || state == "ABORTED" {
		pipe.SRem(ctx, activeTxSet, txID)
		pipe.Expire(ctx, key, 24*time.Hour)
	} else {
		pipe.SAdd(ctx, activeTxSet, txID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update state in redis: %w", err)
	}
	return nil
}

func (r *RedisLogStore) CheckStarted(ctx context.Context, txID string) (bool, error) {
	key := fmt.Sprintf("tx:%s", txID)
	state, err := r.client.HGet(ctx, key, "state").Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state == "COMMITTED" || state == "ABORTED", nil
}

// 3. UPDATED: GetPending now returns coordinator.TransactionRecord
func (r *RedisLogStore) GetPending(ctx context.Context) ([]coordinator.TransactionRecord, error) {
	txIDs, err := r.client.SMembers(ctx, activeTxSet).Result()
	if err != nil {
		return nil, fmt.Errorf("could not fetch active set: %w", err)
	}

	var pending []coordinator.TransactionRecord
	for _, id := range txIDs {
		key := fmt.Sprintf("tx:%s", id)
		state, err := r.client.HGet(ctx, key, "state").Result()
		if err != nil {
			continue
		}

		pending = append(pending, coordinator.TransactionRecord{
			ID:    id,
			State: state,
		})
	}

	return pending, nil
}
