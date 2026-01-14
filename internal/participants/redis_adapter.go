package participants

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisParticipant implements the coordinator.Participant interface.
type RedisParticipant struct {
	client *redis.Client
}

func NewRedisParticipant(client *redis.Client) *RedisParticipant {
	return &RedisParticipant{
		client: client,
	}
}

// Prepare acts as Phase 1.
// It "locks" the resource so no other transaction can modify it.
func (r *RedisParticipant) Prepare(ctx context.Context, txID string) error {
	// In a real app, the lock key would include the specific resource ID (e.g., account:123)
	lockKey := fmt.Sprintf("lock:resource:%s", txID)

	// We use SETNX (Set if Not eXists) to acquire a lock.
	// We set a TTL (Time-To-Live) of 1 minute to prevent permanent deadlocks
	// if the coordinator disappears before Phase 2.
	success, err := r.client.SetNX(ctx, lockKey, "locked", 1*time.Minute).Result()
	if err != nil {
		return fmt.Errorf("redis prepare network error: %w", err)
	}

	if !success {
		return fmt.Errorf("redis prepare failed: resource already locked by another transaction")
	}

	return nil
}

// Commit acts as Phase 2 (Success).
// It applies the actual change and releases the lock.
func (r *RedisParticipant) Commit(ctx context.Context, txID string) error {
	lockKey := fmt.Sprintf("lock:resource:%s", txID)

	// Use a Pipeline to ensure the data update and lock release happen together.
	pipe := r.client.Pipeline()

	// Simulate a business logic update (e.g., updating a cached balance)
	pipe.HIncrBy(ctx, "account:ledger:1", "balance", -100)

	// Remove the lock
	pipe.Del(ctx, lockKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis commit failed: %w", err)
	}

	return nil
}

// Rollback acts as Phase 2 (Failure).
// It simply releases the lock so other transactions can proceed.
func (r *RedisParticipant) Rollback(ctx context.Context, txID string) error {
	lockKey := fmt.Sprintf("lock:resource:%s", txID)

	err := r.client.Del(ctx, lockKey).Err()
	if err != nil {
		return fmt.Errorf("redis rollback failed: %w", err)
	}

	return nil
}
