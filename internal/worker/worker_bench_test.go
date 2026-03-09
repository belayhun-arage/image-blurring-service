package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
)

func newBenchWorker(q *mockQueue, p *mockProcessor, s *mockStorage) *Worker {
	return New(q, p, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// BenchmarkWorker_ProcessJobs measures end-to-end pipeline throughput (dequeue
// → blur → save) using mock dependencies. No real I/O is performed, so this
// isolates the worker coordination overhead.
func BenchmarkWorker_ProcessJobs(b *testing.B) {
	items := make([]string, b.N)
	for i := range items {
		items[i] = fmt.Sprintf("img%d.jpg", i)
	}

	q := &mockQueue{items: items}
	p := &mockProcessor{blurredPath: "/assets/blurred.jpg"}
	s := &mockStorage{}
	w := newBenchWorker(q, p, s)

	b.ResetTimer()
	w.Run(context.Background(), 1)
}

// BenchmarkWorker_ProcessJobs_Parallel benchmarks the worker pool with
// multiple goroutines sharing the same mock queue.
func BenchmarkWorker_ProcessJobs_Parallel(b *testing.B) {
	const concurrency = 4

	items := make([]string, b.N)
	for i := range items {
		items[i] = fmt.Sprintf("img%d.jpg", i)
	}

	q := &mockQueue{items: items}
	p := &mockProcessor{blurredPath: "/assets/blurred.jpg"}
	s := &mockStorage{}
	w := newBenchWorker(q, p, s)

	b.ResetTimer()
	w.Run(context.Background(), concurrency)
}
