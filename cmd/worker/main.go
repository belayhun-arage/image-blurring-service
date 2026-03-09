// Worker is the image processing entry point.
// Its only responsibility is wiring dependencies and starting the worker pool.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/belayhun-arage/image-blur-service/internal/config"
	"github.com/belayhun-arage/image-blur-service/internal/processor"
	"github.com/belayhun-arage/image-blur-service/internal/queue"
	"github.com/belayhun-arage/image-blur-service/internal/storage"
	"github.com/belayhun-arage/image-blur-service/internal/worker"
	"github.com/subosito/gotenv"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := gotenv.Load(".env"); err != nil {
		log.Warn("could not load .env file", "err", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	q, err := queue.NewRedisQueue(cfg.RedisAddr)
	if err != nil {
		log.Error("failed to connect to Redis", "err", err)
		os.Exit(1)
	}

	store, err := storage.NewPostgresStorage(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to Postgres", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	proc := processor.NewImagingProcessor(cfg.AssetsDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("shutdown signal received")
		cancel()
	}()

	log.Info("starting worker pool", "count", cfg.WorkerCount)
	worker.New(q, proc, store, log).Run(ctx, cfg.WorkerCount)
	log.Info("all workers stopped, exiting")
}
