package main

import "testing"

func BenchmarkBlurImageWithImaging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		blurImageWithImaging("q2fpH3bvxESO.jpeg")
	}
}
