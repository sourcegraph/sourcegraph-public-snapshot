package embeddings

import "math"

type Int8Embedding []int8

// Dot computes the dot product between two embeddings. The embeddings must have the
// same length. When the embeddings are normalized, this produces equivalent rankings
// to cosine similarity.
func (e Int8Embedding) Dot(other Int8Embedding) int32 {
	similarity := int32(0)

	count := len(e)
	if count > len(other) {
		// Do this ahead of time so the compiler doesn't need to bounds check
		// every time we index into query.
		panic("mismatched vector lengths")
	}

	i := 0
	for ; i+3 < count; i += 4 {
		m0 := int32(e[i]) * int32(other[i])
		m1 := int32(e[i+1]) * int32(other[i+1])
		m2 := int32(e[i+2]) * int32(other[i+2])
		m3 := int32(e[i+3]) * int32(other[i+3])
		similarity += (m0 + m1 + m2 + m3)
	}

	for ; i < count; i++ {
		similarity += int32(e[i]) * int32(other[i])
	}

	return similarity
}

// Dequantize converts an int8 embedding back into a float32 embedding.
// Converting to an int8 embedding is lossy, so converting to float32
// should only be done for compatibility reasons because it provides
// a false sense of precision.
func (e Int8Embedding) Dequantize() Float32Embedding {
	output := make([]float32, len(e))
	for i, val := range e {
		output[i] = float32(val) / 127.0
	}
	return output
}

type Float32Embedding []float32

// Dot computes the dot product between two embeddings. The embeddings must have the
// same length. When the embeddings are normalized, this produces equivalent rankings
// to cosine similarity.
func (e Float32Embedding) Dot(other Float32Embedding) float32 {
	similarity := float32(0)

	count := len(e)
	if count > len(other) {
		// Do this ahead of time so the compiler doesn't need to bounds check
		// every time we index into query.
		panic("mismatched vector lengths")
	}

	i := 0
	for ; i+3 < count; i += 4 {
		m0 := e[i] * other[i]
		m1 := e[i+1] * other[i+1]
		m2 := e[i+2] * other[i+2]
		m3 := e[i+3] * other[i+3]

		similarity += (m0 + m1 + m2 + m3)
	}

	for ; i < count; i++ {
		similarity += e[i] * other[i]
	}

	return similarity
}

// Quantize reduces the precision of the vectors from float32
// to int8. It uses a simple linear mapping from [-1, 1] to
// [-127, 127].
//
// When compared against rankings from the float32 embeddings, this
// quantization function yielded rankings where the average change in rank was
// only 1.2%. 93 of the top 100 rows  were unchanged, and 950 of the top 1000
// were unchanged.
func (e Float32Embedding) Quantize() Int8Embedding {
	output := make([]int8, len(e))
	for i, val := range e {
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
	return output
}
