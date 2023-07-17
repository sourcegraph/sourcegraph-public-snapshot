package embeddings

import "math"

// Quantize reduces the precision of the vectors from float32
// to int8. It uses a simple linear mapping from [-1, 1] to
// [-127, 127].
//
// When compared against rankings from the float32 embeddings, this
// quantization function yielded rankings where the average change in rank was
// only 1.2%. 93 of the top 100 rows  were unchanged, and 950 of the top 1000
// were unchanged.
//
// When buf is large enough to fit the output, it will be used instead of
// an allocation.
func Quantize(input []float32, buf []int8) []int8 {
	output := buf
	if len(input) > len(buf) {
		output = make([]int8, len(input))
	}
	for i, val := range input {
		// All our inputs should be in [-1, 1],
		// but double check just in case.
		if val > 1 {
			val = 1
		} else if val < -1 {
			val = -1
		}

		// All the inputs should be in the range [-1, 1], so we can use the
		// full range of int8. Rounding instead of truncating is a little more
		// expensive, but it yields values closer to the originals, giving
		// better accuracy.
		output[i] = int8(math.Round(float64(val) * 127.0))
	}
	return output[:len(input)]
}

func Dequantize(input []int8) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val) / 127.0
	}
	return output
}
