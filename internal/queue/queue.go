// Package queue defines the contract for a job queue used by this service.
package queue

import "context"

// Queue is the interface for enqueuing and dequeuing image processing jobs.
// Any implementation (Redis, in-memory, SQS, etc.) must satisfy this contract.
type Queue interface {
	// Enqueue adds an imageID to the back of the queue.
	Enqueue(ctx context.Context, imageID string) error

	// Dequeue blocks until an imageID is available or ctx is cancelled.
	// Returns context.Canceled or context.DeadlineExceeded on cancellation.
	Dequeue(ctx context.Context) (string, error)
}
