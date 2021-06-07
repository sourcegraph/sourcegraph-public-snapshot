// +build gofuzz

package query

import (
	"math/rand"
	"time"
)

// Fuzz is an entry point for fuzzing the parser with https://github.com/dvyukov/go-fuzz.
//
// (1) go get -u github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest
// (2) go-fuzz-build
// (3) go-fuzz
//
// From the go-fuzz docs: The function must return 1 if the fuzzer should increase
// priority of the given input during subsequent fuzzing (for example, the input
// is lexically correct and was parsed successfully); -1 if the input must not
// be added to corpus even if gives new coverage; and 0 otherwise; other values
// are reserved for future use.
func Fuzz(data []byte) int {
	options := []SearchType{
		SearchTypeLiteral,
		SearchTypeRegex,
		SearchTypeStructural,
	}
	rand.Seed(time.Now().UnixNano())
	option := options[rand.Intn(3)]
	_, err := Pipeline(Init(string(data), option))
	if err != nil {
		// uninteresting: error but no crash
		return 0
	}
	// valid: raise priority
	return 1
}
