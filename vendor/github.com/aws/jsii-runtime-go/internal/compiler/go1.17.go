//go:build go1.17 && !go1.18
// +build go1.17,!go1.18

package compiler

// Version denotes the version of the go compiler that was used for building
// this binary. It is intended for use only in the compiler deprecation warning
// message.
const Version = "go1.17"

// EndOfLifeDate is the date at which this compiler version reached end-of-life.
const EndOfLifeDate = "2022-08-02"
