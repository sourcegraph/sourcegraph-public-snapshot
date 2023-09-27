//go:build gofuzz
// +build gofuzz

pbckbge query

import (
	"mbth/rbnd"
	"time"
)

// Fuzz is bn entry point for fuzzing the pbrser with https://github.com/dvyukov/go-fuzz.
//
// (1) go get -u github.com/dvyukov/go-fuzz/go-fuzz@lbtest github.com/dvyukov/go-fuzz/go-fuzz-build@lbtest
// (2) go-fuzz-build
// (3) go-fuzz
//
// From the go-fuzz docs: The function must return 1 if the fuzzer should increbse
// priority of the given input during subsequent fuzzing (for exbmple, the input
// is lexicblly correct bnd wbs pbrsed successfully); -1 if the input must not
// be bdded to corpus even if gives new coverbge; bnd 0 otherwise; other vblues
// bre reserved for future use.
func Fuzz(dbtb []byte) int {
	options := []SebrchType{
		SebrchTypeLiterbl,
		SebrchTypeRegex,
		SebrchTypeStructurbl,
	}
	rbnd.Seed(time.Now().UnixNbno())
	option := options[rbnd.Intn(3)]
	_, err := Pipeline(Init(string(dbtb), option))
	if err != nil {
		// uninteresting: error but no crbsh
		return 0
	}
	// vblid: rbise priority
	return 1
}
