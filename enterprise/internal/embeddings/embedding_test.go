package embeddings

import (
	"testing"
	"testing/quick"
)

func simpleDotInt8(a, b []int8) int32 {
	similarity := int32(0)
	for i := 0; i < len(a); i++ {
		similarity += int32(a[i]) * int32(b[i])
	}
	return similarity
}

func simpleDotFloat32(a, b []float32) float32 {
	similarity := float32(0)
	for i := 0; i < len(a); i++ {
		similarity += a[i] * b[i]
	}
	return similarity
}

func TestDot(t *testing.T) {
	t.Run("int8", func(t *testing.T) {
		f := func(a, b Int8Embedding) bool {
			if len(a) > len(b) {
				a = a[:len(b)]
			} else if len(a) < len(b) {
				b = b[:len(a)]
			}

			want := simpleDotInt8(a, b)
			got := a.Dot(b)
			return want == got
		}
		err := quick.Check(f, nil)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("float32", func(t *testing.T) {
		f := func(a, b Float32Embedding) bool {
			if len(a) > len(b) {
				a = a[:len(b)]
			} else if len(a) < len(b) {
				b = b[:len(a)]
			}

			want := simpleDotFloat32(a, b)
			got := a.Dot(b)
			return want == got
		}
		err := quick.Check(f, nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}
