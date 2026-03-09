// Worker that processes image IDs from a queue
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/belayhun-arage/image-blur-service/internal/queue"
	"github.com/belayhun-arage/image-blur-service/platforms/helper"
	"github.com/disintegration/imaging"
	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"
)

type worker struct {
	db *sql.DB
}

func init() {
	if err := gotenv.Load(".env"); err != nil {
		log.Println("Warning: could not load .env file:", err)
	}
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := queue.InitRedisQueue(redisAddr); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	workerCount := 2
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workerCount = n
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down worker...")
		cancel()
	}()

	w := &worker{db: db}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.processImages(ctx)
		}()
	}
	wg.Wait()
	log.Println("All workers stopped.")
}

func (w *worker) processImages(ctx context.Context) {
	for {
		log.Println("Waiting for image jobs...")
		imageID, err := queue.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled, clean shutdown
			}
			log.Printf("Error while dequeuing: %v\n", err)
			continue
		}
		if imageID == "" {
			log.Println("Received empty image ID, skipping")
			continue
		}

		blurredPath, err := blurImageWithImaging(imageID)
		if err != nil {
			log.Printf("Failed to blur image %q: %v\n", imageID, err)
			continue
		}

		if err := w.saveBlurredImage(imageID, blurredPath); err != nil {
			log.Printf("Failed to save metadata for %q: %v\n", imageID, err)
			continue
		}

		log.Printf("Successfully processed and saved blur for: %s\n", imageID)
	}
}

func blurImageWithImaging(imageID string) (string, error) {
	assetsDir := os.Getenv("ASSETS_DIRECTORY")
	if assetsDir == "" {
		return "", fmt.Errorf("ASSETS_DIRECTORY not set")
	}

	// Prevent path traversal: ensure the resolved path stays inside assetsDir
	absAssets, err := filepath.Abs(assetsDir)
	if err != nil {
		return "", fmt.Errorf("could not resolve assets directory: %w", err)
	}
	originalPath := filepath.Join(assetsDir, imageID)
	absOriginal, err := filepath.Abs(originalPath)
	if err != nil {
		return "", fmt.Errorf("could not resolve image path: %w", err)
	}
	if !strings.HasPrefix(absOriginal, absAssets+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid image ID: path traversal detected")
	}

	img, err := imaging.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	blurred := imaging.Blur(img, 8.0)

	ext := filepath.Ext(imageID)
	blurredName := helper.GenerateRandomString(12, helper.CHARACTERS) + "_blurred" + ext
	blurredPath := filepath.Join(assetsDir, blurredName)

	if err := imaging.Save(blurred, blurredPath); err != nil {
		return "", fmt.Errorf("failed to save blurred image: %w", err)
	}

	return blurredPath, nil
}

func (w *worker) saveBlurredImage(sourceID, blurredPath string) error {
	_, err := w.db.Exec(`INSERT INTO blurred_img (source_img_id, blurred_img_url) VALUES ($1, $2)`, sourceID, blurredPath)
	return err
}
