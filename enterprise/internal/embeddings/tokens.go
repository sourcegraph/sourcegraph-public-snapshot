package embeddings

import "math"

const CHARS_PER_TOKEN = 4

func EstimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / float64(CHARS_PER_TOKEN)))
}
