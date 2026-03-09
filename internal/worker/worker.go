// Package worker contains the core image-blur processing logic.
// It depends only on interfaces, making it fully testable without
// real Redis, Postgres, or filesystem access.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/belayhun-arage/image-blur-service/internal/processor"
	"github.com/belayhun-arage/image-blur-service/internal/queue"
	"github.com/belayhun-arage/image-blur-service/internal/storage"
)

// Worker pulls image IDs from a Queue, blurs them via a Processor,
// and persists the result via Storage.
type Worker struct {
	queue     queue.Queue
	processor processor.Processor
	storage   storage.Storage
	log       *slog.Logger
}

// New creates a Worker with the given dependencies.
func New(q queue.Queue, p processor.Processor, s storage.Storage, log *slog.Logger) *Worker {
	return &Worker{
		queue:     q,
		processor: p,
		storage:   s,
		log:       log,
	}
}

// Run launches count goroutines that each process jobs until ctx is cancelled.
// It blocks until all goroutines have exited.
func (w *Worker) Run(ctx context.Context, count int) {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.log.Info("worker started", "id", id)
			w.loop(ctx, id)
			w.log.Info("worker stopped", "id", id)
		}(i)
	}
	wg.Wait()
}

// loop runs the dequeue → blur → save cycle for a single worker goroutine.
func (w *Worker) loop(ctx context.Context, id int) {
	for {
		imageID, err := w.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return // clean shutdown
			}
			w.log.Error("dequeue failed", "worker", id, "err", err)
			continue
		}
		if imageID == "" {
			w.log.Warn("received empty image ID, skipping", "worker", id)
			continue
		}

		w.process(ctx, id, imageID)
	}
}

// process handles a single image job. Errors are logged but not fatal.
func (w *Worker) process(ctx context.Context, workerID int, imageID string) {
	log := w.log.With("worker", workerID, "imageID", imageID)

	blurredPath, err := w.processor.Blur(imageID)
	if err != nil {
		log.Error("blur failed", "err", err)
		return
	}

	if err := w.storage.SaveBlurredImage(ctx, imageID, blurredPath); err != nil {
		log.Error("save failed", "blurredPath", blurredPath, "err", err)
		return
	}

	log.Info("image processed", "blurredPath", blurredPath)
}
