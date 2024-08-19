package cast

import (
	"slices"
	"testing"
	"testing/quick"
)

type myStringType string

func TestStrings(t *testing.T) {
	err := quick.Check(func(input []string) bool {
		roundtripped := ToStrings(FromStrings[myStringType](input))
		return slices.Equal(input, roundtripped)
	}, nil)

	if err != nil {
		t.Fatal(err)
	}
}
