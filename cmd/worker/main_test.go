package main

import "testing"

func BenchmarkBlurImage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BlurImage("q2fpH3bvxESO.jpeg")
	}

}

func BenchmarkLLm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// BlurImageWithImaging("q2fpH3bvxESO.jpeg")
		x := 5
		if x < 5 {
			b.Fatal("expected 5, got ", x)
		}
	}

}

func TestFDjdsfldskfsd(t *testing.T) {
	t.Parallel()
	s := 5
	if s != 5 {
		t.Error("expected 5, got ", s)
	}
}

//func TestWorking(t *testing.T)
