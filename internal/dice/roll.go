package dice

import "github.com/sourcegraph/sourcegraph/internal/random"

func RollPair() int {
	roll1 := random.RandomInt(6)
	roll2 := random.RandomInt(6)
	return roll1 + roll2
}
