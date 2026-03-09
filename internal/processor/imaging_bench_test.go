package processor

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// makePNG writes a synthetic RGBA image of the given dimensions to dir/name.
func makePNG(tb testing.TB, dir, name string, width, height int) {
	tb.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		tb.Fatal(err)
	}
}

// BenchmarkImagingProcessor_Blur measures blur throughput across image sizes.
// Blurred output files are removed after each iteration to avoid disk pressure.
func BenchmarkImagingProcessor_Blur(b *testing.B) {
	sizes := []struct {
		name          string
		width, height int
	}{
		{"256x256", 256, 256},
		{"512x512", 512, 512},
		{"1024x1024", 1024, 1024},
	}

	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			dir := b.TempDir()
			const imgName = "input.png"
			makePNG(b, dir, imgName, tc.width, tc.height)

			p := NewImagingProcessor(dir)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				blurredPath, err := p.Blur(imgName)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				os.Remove(blurredPath)
				b.StartTimer()
			}
		})
	}
}

// BenchmarkImagingProcessor_safePath measures path validation overhead.
func BenchmarkImagingProcessor_safePath(b *testing.B) {
	p := NewImagingProcessor("/tmp/assets")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.safePath("photo.jpg") //nolint:errcheck
	}
}
