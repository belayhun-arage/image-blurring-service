// Package processor defines the contract for image processing operations.
package processor

// Processor applies transformations to images on disk.
type Processor interface {
	// Blur reads the image at imageID (relative to the configured assets
	// directory), applies a Gaussian blur, writes the result to the same
	// directory, and returns the full path of the blurred file.
	Blur(imageID string) (blurredPath string, err error)
}
