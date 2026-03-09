package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const queueName = "image_blur_queue"

func InitRedisQueue(addr string) error {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return rdb.Ping(context.Background()).Err()
}

func Enqueue(ctx context.Context, imageID string) error {
	return rdb.RPush(ctx, queueName, imageID).Err()
}

// Dequeue blocks until an item is available or ctx is cancelled.
func Dequeue(ctx context.Context) (string, error) {
	result, err := rdb.BLPop(ctx, 0, queueName).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", fmt.Errorf("unexpected BLPop result length: %d", len(result))
	}
	return result[1], nil
}
