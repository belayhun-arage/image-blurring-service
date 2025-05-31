package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

const queueName = "image_blur_queue"

func InitRedisQueue(addr string) error {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return rdb.Ping(ctx).Err()
}

func Enqueue(imageID string) error {
	return rdb.RPush(ctx, queueName, imageID).Err()
}

func Dequeue() (string, error) {
	result, err := rdb.BLPop(ctx, 0, queueName).Result()
	if err != nil || len(result) < 2 {
		return "", err
	}
	return result[1], nil
}
