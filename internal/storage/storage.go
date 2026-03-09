// Package storage defines the contract for persisting image metadata.
package storage

import "context"

// Storage persists and retrieves blurred image records.
type Storage interface {
	// SaveBlurredImage records that sourceID was blurred and stored at blurredPath.
	SaveBlurredImage(ctx context.Context, sourceID, blurredPath string) error

	// DeleteBlurredImage removes all blurred records for sourceID.
	// Returns the number of rows deleted.
	DeleteBlurredImage(ctx context.Context, sourceID string) (int64, error)
}
