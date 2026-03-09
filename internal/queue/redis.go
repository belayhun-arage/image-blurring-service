package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const queueKey = "image_blur_queue"

// RedisQueue is a Redis-backed FIFO queue.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a RedisQueue and verifies the connection.
func NewRedisQueue(addr string) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &RedisQueue{client: client}, nil
}

// Enqueue pushes imageID to the tail of the Redis list.
func (q *RedisQueue) Enqueue(ctx context.Context, imageID string) error {
	return q.client.RPush(ctx, queueKey, imageID).Err()
}

// Dequeue blocks until an item is available or ctx is cancelled.
// It returns context.Canceled when ctx is done.
func (q *RedisQueue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.BLPop(ctx, 0, queueKey).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", fmt.Errorf("unexpected BLPop result length: %d", len(result))
	}
	return result[1], nil
}

// Compile-time check that RedisQueue satisfies the Queue interface.
var _ Queue = (*RedisQueue)(nil)
