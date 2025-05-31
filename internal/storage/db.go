//  Mock DB and file storage
package storage

import "sync"

var blurredImages = map[string]bool{}
var mu sync.Mutex

func Exists(imageID string) bool {
	mu.Lock()
	defer mu.Unlock()
	return blurredImages[imageID]
}

func MarkAsBlurred(imageID string) {
	mu.Lock()
	defer mu.Unlock()
	blurredImages[imageID] = true
}
