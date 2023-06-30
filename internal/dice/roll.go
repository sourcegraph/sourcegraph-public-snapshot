package dice

import "github.com/sourcegraph/sourcegraph/internal/random"

func RollPair(n int) int {
	return random.RandomInt(n) * random.RandomInt(n)
}
