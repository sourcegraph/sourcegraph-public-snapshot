// Copyright 2012 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff_test

import (
	"github.com/mb0/diff"
	"testing"
)

type testcase struct {
	name string
	a, b []int
	res  []diff.Change
}

var tests = []testcase{
	{"shift",
		[]int{1, 2, 3},
		[]int{0, 1, 2, 3},
		[]diff.Change{{0, 0, 0, 1}},
	},
	{"push",
		[]int{1, 2, 3},
		[]int{1, 2, 3, 4},
		[]diff.Change{{3, 3, 0, 1}},
	},
	{"unshift",
		[]int{0, 1, 2, 3},
		[]int{1, 2, 3},
		[]diff.Change{{0, 0, 1, 0}},
	},
	{"pop",
		[]int{1, 2, 3, 4},
		[]int{1, 2, 3},
		[]diff.Change{{3, 3, 1, 0}},
	},
	{"all changed",
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]int{10, 11, 12, 13, 14},
		[]diff.Change{
			{0, 0, 10, 5},
		},
	},
	{"all same",
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]diff.Change{},
	},
	{"wrap",
		[]int{1},
		[]int{0, 1, 2, 3},
		[]diff.Change{
			{0, 0, 0, 1},
			{1, 2, 0, 2},
		},
	},
	{"snake",
		[]int{0, 1, 2, 3, 4, 5},
		[]int{1, 2, 3, 4, 5, 6},
		[]diff.Change{
			{0, 0, 1, 0},
			{6, 5, 0, 1},
		},
	},
	// note: input is ambiguous
	// first two traces differ from fig.1
	// it still is a lcs and ses path
	{"paper fig. 1",
		[]int{1, 2, 3, 1, 2, 2, 1},
		[]int{3, 2, 1, 2, 1, 3},
		[]diff.Change{
			{0, 0, 1, 1},
			{2, 2, 1, 0},
			{5, 4, 1, 0},
			{7, 5, 0, 1},
		},
	},
}

func TestDiffAB(t *testing.T) {
	for _, test := range tests {
		res := diff.Ints(test.a, test.b)
		if len(res) != len(test.res) {
			t.Error(test.name, "expected length", len(test.res), "for", res)
			continue
		}
		for i, c := range test.res {
			if c != res[i] {
				t.Error(test.name, "expected ", c, "got", res[i])
			}
		}
	}
}

func TestDiffBA(t *testing.T) {
	// interesting: fig.1 Diff(b, a) results in the same path as `diff -d a b`
	tests[len(tests)-1].res = []diff.Change{
		{0, 0, 2, 0},
		{3, 1, 1, 0},
		{5, 2, 0, 1},
		{7, 5, 0, 1},
	}
	for _, test := range tests {
		res := diff.Ints(test.b, test.a)
		if len(res) != len(test.res) {
			t.Error(test.name, "expected length", len(test.res), "for", res)
			continue
		}
		for i, c := range test.res {
			// flip change data also
			rc := diff.Change{c.B, c.A, c.Ins, c.Del}
			if rc != res[i] {
				t.Error(test.name, "expected ", rc, "got", res[i])
			}
		}
	}
}

func diffsEqual(a, b []diff.Change) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGranularStrings(t *testing.T) {
	a := "abcdefghijklmnopqrstuvwxyza"
	b := "AbCdeFghiJklmnOpqrstUvwxyzab"
	// each iteration of i increases granularity and will absorb one more lower-letter-followed-by-upper-letters sequence
	changesI := [][]diff.Change{
		{{0, 0, 1, 1}, {2, 2, 1, 1}, {5, 5, 1, 1}, {9, 9, 1, 1}, {14, 14, 1, 1}, {20, 20, 1, 1}, {27, 27, 0, 1}},
		{{0, 0, 3, 3}, {5, 5, 1, 1}, {9, 9, 1, 1}, {14, 14, 1, 1}, {20, 20, 1, 1}, {27, 27, 0, 1}},
		{{0, 0, 6, 6}, {9, 9, 1, 1}, {14, 14, 1, 1}, {20, 20, 1, 1}, {27, 27, 0, 1}},
		{{0, 0, 10, 10}, {14, 14, 1, 1}, {20, 20, 1, 1}, {27, 27, 0, 1}},
		{{0, 0, 15, 15}, {20, 20, 1, 1}, {27, 27, 0, 1}},
		{{0, 0, 21, 21}, {27, 27, 0, 1}},
		{{0, 0, 27, 28}},
	}
	for i := 0; i < len(changesI); i++ {
		diffs := diff.Granular(i, diff.ByteStrings(a, b))
		if !diffsEqual(diffs, changesI[i]) {
			t.Errorf("expected %v, got %v", diffs, changesI[i])
		}
	}
}

func TestDiffRunes(t *testing.T) {
	a := []rune("brown fox jumps over the lazy dog")
	b := []rune("brwn faax junps ovver the lay dago")
	res := diff.Runes(a, b)
	echange := []diff.Change{
		{2, 2, 1, 0},
		{7, 6, 1, 2},
		{12, 12, 1, 1},
		{18, 18, 0, 1},
		{27, 28, 1, 0},
		{31, 31, 0, 2},
		{32, 34, 1, 0},
	}
	for i, c := range res {
		t.Log(c)
		if c != echange[i] {
			t.Error("expected", echange[i], "got", c)
		}
	}
}

func TestDiffByteStrings(t *testing.T) {
	a := "brown fox jumps over the lazy dog"
	b := "brwn faax junps ovver the lay dago"
	res := diff.ByteStrings(a, b)
	echange := []diff.Change{
		{2, 2, 1, 0},
		{7, 6, 1, 2},
		{12, 12, 1, 1},
		{18, 18, 0, 1},
		{27, 28, 1, 0},
		{31, 31, 0, 2},
		{32, 34, 1, 0},
	}
	for i, c := range res {
		t.Log(c)
		if c != echange[i] {
			t.Error("expected", echange[i], "got", c)
		}
	}
}

type ints struct{ a, b []int }

func (d *ints) Equal(i, j int) bool { return d.a[i] == d.b[j] }
func BenchmarkDiff(b *testing.B) {
	t := tests[len(tests)-1]
	d := &ints{t.a, t.b}
	n, m := len(d.a), len(d.b)
	for i := 0; i < b.N; i++ {
		diff.Diff(n, m, d)
	}
}

func BenchmarkInts(b *testing.B) {
	t := tests[len(tests)-1]
	d1 := t.a
	d2 := t.b
	for i := 0; i < b.N; i++ {
		diff.Ints(d1, d2)
	}
}

func BenchmarkDiffRunes(b *testing.B) {
	d1 := []rune("1231221")
	d2 := []rune("321213")
	for i := 0; i < b.N; i++ {
		diff.Runes(d1, d2)
	}
}

func BenchmarkDiffBytes(b *testing.B) {
	d1 := []byte("lorem ipsum dolor sit amet consectetur")
	d2 := []byte("lorem lovesum daenerys targaryen ami consecteture")
	for i := 0; i < b.N; i++ {
		diff.Bytes(d1, d2)
	}
}

func BenchmarkDiffByteStrings(b *testing.B) {
	d1 := "lorem ipsum dolor sit amet consectetur"
	d2 := "lorem lovesum daenerys targaryen ami consecteture"
	for i := 0; i < b.N; i++ {
		diff.ByteStrings(d1, d2)
	}
}
