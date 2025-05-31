//  # Image processing logic
package processor

import (
	"bytes"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
)

func BlurImage(imgBytes []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}
	blurred := imaging.Blur(img, 5)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, blurred, nil)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
