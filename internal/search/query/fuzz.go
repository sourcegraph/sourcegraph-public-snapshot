// +build gofuzz

package query

// Fuzz is an entry point for fuzzing the parser with https://github.com/dvyukov/go-fuzz.
// Run go-fuzz-build and then go-fuzz in this directory.
//
// From the go-fuzz docs: The function must return 1 if the fuzzer should increase
// priority of the given input during subsequent fuzzing (for example, the input
// is lexically correct and was parsed successfully); -1 if the input must not
// be added to corpus even if gives new coverage; and 0 otherwise; other values
// are reserved for future use.
func Fuzz(data []byte) int {
	_, err := ProcessAndOr(string(data))
	if err != nil {
		return 0
	}
	return 1
}
