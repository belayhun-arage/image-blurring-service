// Worker that processes image IDs from a channel
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/belayhun-arage/image-blur-service/internal/queue"
	"github.com/belayhun-arage/image-blur-service/platforms/helper"
	"github.com/disintegration/imaging"
	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"
)

var db *sql.DB

func init() {
	err := gotenv.Load("../../.env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	errr := queue.InitRedisQueue("localhost:6379")
	if errr != nil {
		log.Fatalf("Failed to connect to Redis: %v", errr)
	}

	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()
	go ProcessImages()
	ProcessImages()
}

func ProcessImages() {
	for {
		fmt.Println("Waiting for image jobs...")
		imageID, err := queue.Dequeue()
		if err != nil {
			log.Printf("Error while dequeuing: %v\n", err)
			continue
		}

		// blurredPath, err := BlurImage(imageID)
		// if err != nil {
		// 	log.Printf("Failed to blur image: %v\n", err)
		// 	continue
		// }

		// // Insert metadata into blurred_img table
		// err = SaveBlurredImage(imageID, blurredPath)
		// if err != nil {
		// 	log.Printf("Failed to save metadata: %v\n", err)
		// 	continue
		// }

		blurredPathPath, err := BlurImageWithImaging(imageID)
		if err != nil {
			log.Printf("Failed to blur image: %v\n", err)
			continue
		}

		// Insert metadata into blurred_img table
		err = SaveBlurredImage(imageID, blurredPathPath)
		if err != nil {
			log.Printf("Failed to save metadata: %v\n", err)
			continue
		}

		log.Printf("Successfully processed and saved blur for: %s\n", imageID)
	}
}

func BlurImage(imageID string) (string, error) {
	const maxSizeKB = 20

	assetsDir := os.Getenv("ASSETS_DIRECTORY")
	if assetsDir == "" {
		return "", fmt.Errorf("assets directory not set")
	}
	originalPath := filepath.Join(assetsDir, imageID)

	// Output path
	blurredName := helper.GenerateRandomString(12, helper.CHARACTERS) + "FFMPEG" + "_blurred.jpg"
	blurredPath := filepath.Join(assetsDir, blurredName)

	// FFmpeg command with gblur and JPEG compression
	cmd := exec.Command("ffmpeg",
		"-y",               // Overwrite if exists
		"-i", originalPath, // Input file
		"-vf", "gblur=sigma=20", // Gaussian blur
		"-q:v", "5", // JPEG quality (1=best, 31=worst)
		blurredPath, // Output file
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %s", stderr.String())
	}

	// Check file size
	info, err := os.Stat(blurredPath)
	if err != nil {
		return "", fmt.Errorf("could not stat output file: %w", err)
	}

	if info.Size() > maxSizeKB*1024 {
		os.Remove(blurredPath) // Clean up if too big
		return "", fmt.Errorf("blurred image exceeds %dKB: %d bytes", maxSizeKB, info.Size())
	}

	return blurredPath, nil
}

func BlurImageWithImaging(imageID string) (string, error) {
	assetsDir := os.Getenv("ASSETS_DIRECTORY")
	if assetsDir == "" {
		return "", fmt.Errorf("assets directory not set")
	}
	originalPath := filepath.Join(assetsDir, imageID)

	img, err := imaging.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	blurred := imaging.Blur(img, 8.0)

	// Extract original extension and reuse it
	ext := filepath.Ext(imageID)
	blurredName := helper.GenerateRandomString(12, helper.CHARACTERS) + "Imaging" + "_blurred" + ext
	blurredPath := filepath.Join(assetsDir, blurredName)

	err = imaging.Save(blurred, blurredPath)
	if err != nil {
		return "", fmt.Errorf("failed to save blurred image: %w", err)
	}

	return blurredPath, nil
}

func FetchImage(imageID string) ([]byte, error) {
	// Mock: Read from a file path or dummy data
	path := filepath.Join(os.Getenv("ASSETS_DIRECTORY"), imageID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	return data, nil
}

func SaveBlurredImage(sourceID, blurredPath string) error {
	fmt.Println("Saving blurred image metadata to DB...")
	fmt.Println("Blurred image path:", blurredPath)
	fmt.Println("Source image ID:", sourceID)
	query := `INSERT INTO blurred_img (source_img_id, blurred_img_url) VALUES ($1, $2)`
	_, err := db.Exec(query, sourceID, blurredPath)
	return err
}
