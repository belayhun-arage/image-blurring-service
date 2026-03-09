package processor

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

const (
	blurSigma  = 8.0
	randChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randLength = 12
)

// ImagingProcessor blurs images using the disintegration/imaging library.
type ImagingProcessor struct {
	assetsDir string
	mu        sync.Mutex
}

// NewImagingProcessor creates an ImagingProcessor rooted at assetsDir.
func NewImagingProcessor(assetsDir string) *ImagingProcessor {
	return &ImagingProcessor{assetsDir: assetsDir}
}

// Blur applies a Gaussian blur to the image identified by imageID and writes
// the result to assetsDir. It guards against path traversal attacks.
func (p *ImagingProcessor) Blur(imageID string) (string, error) {
	originalPath, err := p.safePath(imageID)
	if err != nil {
		return "", err
	}

	img, err := imaging.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("open image %q: %w", imageID, err)
	}

	blurred := imaging.Blur(img, blurSigma)

	ext := filepath.Ext(imageID)
	blurredName := p.randomName(randLength) + "_blurred" + ext
	blurredPath := filepath.Join(p.assetsDir, blurredName)

	if err := imaging.Save(blurred, blurredPath); err != nil {
		return "", fmt.Errorf("save blurred image: %w", err)
	}

	return blurredPath, nil
}

// safePath resolves imageID relative to assetsDir and rejects any path that
// escapes the directory (e.g. "../../etc/passwd").
func (p *ImagingProcessor) safePath(imageID string) (string, error) {
	absAssets, err := filepath.Abs(p.assetsDir)
	if err != nil {
		return "", fmt.Errorf("resolve assets directory: %w", err)
	}
	absTarget, err := filepath.Abs(filepath.Join(p.assetsDir, imageID))
	if err != nil {
		return "", fmt.Errorf("resolve image path: %w", err)
	}
	if !strings.HasPrefix(absTarget, absAssets+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid imageID %q: path traversal detected", imageID)
	}
	return absTarget, nil
}

func (p *ImagingProcessor) randomName(n int) string {
	b := make([]byte, n)
	p.mu.Lock()
	for i := range b {
		b[i] = randChars[rand.Intn(len(randChars))]
	}
	p.mu.Unlock()
	return string(b)
}

// Compile-time check that ImagingProcessor satisfies Processor.
var _ Processor = (*ImagingProcessor)(nil)
