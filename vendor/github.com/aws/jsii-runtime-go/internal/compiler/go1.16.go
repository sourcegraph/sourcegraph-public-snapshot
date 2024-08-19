//go:build go1.16 && !go1.17
// +build go1.16,!go1.17

package compiler

// Version denotes the version of the go compiler that was used for building
// this binary. It is intended for use only in the compiler deprecation warning
// message.
const Version = "go1.16"

// EndOfLifeDate is the date at which this compiler version reached end-of-life.
const EndOfLifeDate = "2022-03-15"
