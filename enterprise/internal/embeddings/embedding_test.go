package embeddings

import (
	"fmt"
	"math"
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
		cleanInput := func(input []float32) []float32 {
			for i, f := range input {
				if math.IsInf(float64(f), 0) {
					input[i] = 0.9
				} else if math.IsNaN(float64(f)) {
					input[i] = 0.2
				} else if f > 1.0 {
					input[i] = 1.0
				} else if f < -1.0 {
					input[i] = -1.0
				}
			}
			return input
		}

		f := func(a, b Float32Embedding) bool {
			if len(a) > len(b) {
				a = a[:len(b)]
			} else if len(a) < len(b) {
				b = b[:len(a)]
			}

			a = cleanInput(a)
			b = cleanInput(b)

			want := simpleDotFloat32(a, b)
			got := a.Dot(b)

			if want == got {
				return true
			}

			// There is no guarantee that reordering/instruction merging
			// will return the exact same results, so test with an epsilon.
			if math.Abs(float64(want-got))/float64(want) < 0.001 {
				return true
			}

			fmt.Printf("got: %.10f, want: %.10f\n", got, want)
			return false
		}
		err := quick.Check(f, nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}
