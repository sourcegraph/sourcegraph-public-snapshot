package embeddings

import (
	"testing"
	"testing/quick"
)

func TestQuantize(t *testing.T) {
	t.Run("buf doesn't affect output", func(t *testing.T) {
		a := func(input []float32, buf []int8) []int8 {
			return Quantize(input, buf)
		}

		b := func(input []float32, buf []int8) []int8 {
			return Quantize(input, nil)
		}
		quick.CheckEqual(a, b, nil)
	})
}
