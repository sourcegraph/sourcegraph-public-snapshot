// Copyright 2012 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff_test

import (
	"fmt"
	"github.com/mb0/diff"
)

// Diff on inputs with different representations
type MixedInput struct {
	A []int
	B []string
}

var names map[string]int

func (m *MixedInput) Equal(a, b int) bool {
	return m.A[a] == names[m.B[b]]
}

func ExampleDiff() {
	names = map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	m := &MixedInput{
		[]int{1, 2, 3, 1, 2, 2, 1},
		[]string{"three", "two", "one", "two", "one", "three"},
	}
	changes := diff.Diff(len(m.A), len(m.B), m)
	for _, c := range changes {
		fmt.Println("change at", c.A, c.B)
	}
	// Output:
	// change at 0 0
	// change at 2 2
	// change at 5 4
	// change at 7 5
}

func ExampleGranular() {
	a := "hElLo!"
	b := "hello!"
	changes := diff.Granular(5, diff.ByteStrings(a, b)) // ignore small gaps in differences
	for l := len(changes) - 1; l >= 0; l-- {
		change := changes[l]
		b = b[:change.B] + "|" + b[change.B:change.B+change.Ins] + "|" + b[change.B+change.Ins:]
	}
	fmt.Println(b)
	// Output:
	// h|ell|o!
}
