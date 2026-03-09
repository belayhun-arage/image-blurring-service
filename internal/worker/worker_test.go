package worker

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"
)

// --- mocks ---

type mockQueue struct {
	items []string
	err   error
}

func (m *mockQueue) Enqueue(_ context.Context, id string) error {
	m.items = append(m.items, id)
	return nil
}

func (m *mockQueue) Dequeue(_ context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if len(m.items) == 0 {
		return "", context.Canceled // signal the loop to stop
	}
	id := m.items[0]
	m.items = m.items[1:]
	return id, nil
}

type mockProcessor struct {
	blurredPath string
	err         error
	calls       []string
}

func (m *mockProcessor) Blur(imageID string) (string, error) {
	m.calls = append(m.calls, imageID)
	return m.blurredPath, m.err
}

type mockStorage struct {
	saved []string
	err   error
}

func (m *mockStorage) SaveBlurredImage(_ context.Context, sourceID, _ string) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, sourceID)
	return nil
}

func (m *mockStorage) DeleteBlurredImage(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func newTestWorker(q *mockQueue, p *mockProcessor, s *mockStorage) *Worker {
	return New(q, p, s, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// --- tests ---

func TestWorker_ProcessesJobSuccessfully(t *testing.T) {
	q := &mockQueue{items: []string{"img001.jpg"}}
	p := &mockProcessor{blurredPath: "/assets/blurred.jpg"}
	s := &mockStorage{}

	w := newTestWorker(q, p, s)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	w.Run(ctx, 1)

	if len(p.calls) != 1 || p.calls[0] != "img001.jpg" {
		t.Errorf("expected Blur called with img001.jpg, got %v", p.calls)
	}
	if len(s.saved) != 1 || s.saved[0] != "img001.jpg" {
		t.Errorf("expected SaveBlurredImage called with img001.jpg, got %v", s.saved)
	}
}

func TestWorker_SkipsEmptyImageID(t *testing.T) {
	q := &mockQueue{items: []string{""}}
	p := &mockProcessor{}
	s := &mockStorage{}

	w := newTestWorker(q, p, s)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	w.Run(ctx, 1)

	if len(p.calls) != 0 {
		t.Errorf("expected no Blur calls for empty imageID, got %d", len(p.calls))
	}
}

func TestWorker_ContinuesAfterBlurError(t *testing.T) {
	q := &mockQueue{items: []string{"bad.jpg", "good.jpg"}}
	p := &mockProcessor{err: errors.New("blur failed")}
	s := &mockStorage{}

	w := newTestWorker(q, p, s)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w.Run(ctx, 1)

	if len(p.calls) != 2 {
		t.Errorf("expected Blur called twice, got %d", len(p.calls))
	}
	if len(s.saved) != 0 {
		t.Errorf("expected no saves on blur failure, got %d", len(s.saved))
	}
}

func TestWorker_ContinuesAfterStorageError(t *testing.T) {
	q := &mockQueue{items: []string{"img.jpg"}}
	p := &mockProcessor{blurredPath: "/assets/blurred.jpg"}
	s := &mockStorage{err: errors.New("db down")}

	w := newTestWorker(q, p, s)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	w.Run(ctx, 1)

	if len(p.calls) != 1 {
		t.Errorf("expected Blur to be called, got %d calls", len(p.calls))
	}
	// storage returned error — nothing should be in saved
	if len(s.saved) != 0 {
		t.Errorf("expected no successful saves, got %d", len(s.saved))
	}
}

func TestWorker_StopsOnContextCancellation(t *testing.T) {
	// Queue that never returns — worker must stop when ctx is cancelled.
	q := &mockQueue{err: context.Canceled}
	p := &mockProcessor{}
	s := &mockStorage{}

	w := newTestWorker(q, p, s)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		w.Run(ctx, 1)
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}
