package embeddings

func QuantizeFloats(input []float32) []int8 {
	output := make([]int8, len(input))
	for i, val := range input {
		// All the inputs should be in the range [-1, 1],
		// so we can use the full range of int8
		output[i] = int8(val * 127.0)
	}
	return output
}

func DequantizeFloats(input []int8) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val) / 127.0
	}
	return output
}
