// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zoekt

import (
	"cmp"
	"encoding/binary"
	"math"
	"sort"
	"unicode"
	"unicode/utf8"
)

func generateCaseNgrams(g ngram) []ngram {
	asRunes := ngramToRunes(g)

	variants := make([]ngram, 0, 8)
	cur := asRunes
	for {
		for i := 0; i < 3; i++ {
			next := unicode.SimpleFold(cur[i])
			cur[i] = next
			if next != asRunes[i] {
				break
			}
		}

		variants = append(variants, runesToNGram(cur))
		if cur == asRunes {
			break
		}
	}

	return variants
}

func toLower(in []byte) []byte {
	out := make([]byte, 0, len(in))
	var buf [4]byte
	for _, c := range string(in) {
		i := utf8.EncodeRune(buf[:], unicode.ToLower(c))
		out = append(out, buf[:i]...)
	}
	return out
}

// compare 'lower' and 'mixed', where lower is the needle. 'mixed' may
// be larger than 'lower'. Returns whether there was a match, and if
// yes, the byte size of the match.
func caseFoldingEqualsRunes(lower, mixed []byte) (int, bool) {
	matchTotal := 0
	for len(lower) > 0 && len(mixed) > 0 {
		lr, lsz := utf8.DecodeRune(lower)
		lower = lower[lsz:]

		mr, msz := utf8.DecodeRune(mixed)
		mixed = mixed[msz:]
		matchTotal += msz

		if lr != unicode.ToLower(mr) {
			return 0, false
		}
	}

	return matchTotal, len(lower) == 0
}

type ngram uint64

func runesToNGram(b [ngramSize]rune) ngram {
	return ngram(uint64(b[0])<<42 | uint64(b[1])<<21 | uint64(b[2]))
}

func bytesToNGram(b []byte) ngram {
	return runesToNGram([ngramSize]rune{rune(b[0]), rune(b[1]), rune(b[2])})
}

func stringToNGram(s string) ngram {
	return bytesToNGram([]byte(s))
}

func ngramToBytes(n ngram) []byte {
	rs := ngramToRunes(n)
	return []byte{byte(rs[0]), byte(rs[1]), byte(rs[2])}
}

const runeMask = 1<<21 - 1

func ngramToRunes(n ngram) [ngramSize]rune {
	return [ngramSize]rune{rune((n >> 42) & runeMask), rune((n >> 21) & runeMask), rune(n & runeMask)}
}

func (n ngram) String() string {
	rs := ngramToRunes(n)
	return string(rs[:])
}

type runeNgramOff struct {
	ngram ngram
	// index is the original index inside of the returned array of splitNGrams
	index int
}

func (a runeNgramOff) Compare(b runeNgramOff) int {
	if a.ngram == b.ngram {
		return cmp.Compare(a.index, b.index)
	} else if a.ngram < b.ngram {
		return -1
	} else {
		return 1
	}
}

func splitNGrams(str []byte) []runeNgramOff {
	var runeGram [3]rune
	var off [3]uint32
	var runeCount int

	result := make([]runeNgramOff, 0, len(str))
	var i uint32

	for len(str) > 0 {
		r, sz := utf8.DecodeRune(str)
		str = str[sz:]
		runeGram[0] = runeGram[1]
		off[0] = off[1]
		runeGram[1] = runeGram[2]
		off[1] = off[2]
		runeGram[2] = r
		off[2] = uint32(i)
		i += uint32(sz)
		runeCount++
		if runeCount < ngramSize {
			continue
		}

		ng := runesToNGram(runeGram)
		result = append(result, runeNgramOff{
			ngram: ng,
			index: len(result),
		})
	}

	return result
}

const (
	_classLowerChar int = iota
	_classUpperChar
	_classDigit
	_classPunct
	_classOther
	_classSpace
)

func byteClass(c byte) int {
	if c >= 'a' && c <= 'z' {
		return _classLowerChar
	}
	if c >= 'A' && c <= 'Z' {
		return _classUpperChar
	}
	if c >= '0' && c <= '9' {
		return _classDigit
	}

	switch c {
	case ' ', '\n':
		return _classSpace
	case '.', ',', ';', '"', '\'':
		return _classPunct
	default:
		return _classOther
	}
}

func characterClass(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func marshalDocSections(secs []DocumentSection) []byte {
	ints := make([]uint32, 0, len(secs)*2)
	for _, s := range secs {
		ints = append(ints, uint32(s.Start), uint32(s.End))
	}

	return toSizedDeltas(ints)
}

func unmarshalDocSections(data []byte, ds []DocumentSection) []DocumentSection {
	sz, m := binary.Uvarint(data)
	data = data[m:]

	if cap(ds) < int(sz)/2 {
		ds = make([]DocumentSection, 0, sz/2)
	} else {
		ds = ds[:0]
	}

	// Inlining the delta decoding to avoid unnecessary allocations that would come
	// from the straightforward implementation, i.e. packing the result of fromSizedDeltas.
	var last uint32
	for len(data) > 0 {
		var d DocumentSection

		delta, m := binary.Uvarint(data)
		last += uint32(delta)
		data = data[m:]
		d.Start = last

		delta, m = binary.Uvarint(data)
		last += uint32(delta)
		data = data[m:]
		d.End = last

		ds = append(ds, d)
	}
	return ds
}

type ngramSlice []ngram

func (p ngramSlice) Len() int { return len(p) }

func (p ngramSlice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p ngramSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func toSizedDeltas(offsets []uint32) []byte {
	var enc [8]byte

	deltas := make([]byte, 0, len(offsets)*2)

	m := binary.PutUvarint(enc[:], uint64(len(offsets)))
	deltas = append(deltas, enc[:m]...)

	var last uint32
	for _, p := range offsets {
		delta := p - last
		last = p

		m := binary.PutUvarint(enc[:], uint64(delta))
		deltas = append(deltas, enc[:m]...)
	}
	return deltas
}

func fromSizedDeltas(data []byte, ps []uint32) []uint32 {
	sz, m := binary.Uvarint(data)
	data = data[m:]

	if cap(ps) < int(sz) {
		ps = make([]uint32, 0, sz)
	} else {
		ps = ps[:0]
	}

	var last uint32
	for len(data) > 0 {
		delta, m := binary.Uvarint(data)
		offset := last + uint32(delta)
		last = offset
		data = data[m:]
		ps = append(ps, offset)
	}
	return ps
}

func toSizedDeltas16(offsets []uint16) []byte {
	var enc [8]byte

	deltas := make([]byte, 0, len(offsets)*2)

	m := binary.PutUvarint(enc[:], uint64(len(offsets)))
	deltas = append(deltas, enc[:m]...)

	var last uint16
	for _, p := range offsets {
		delta := p - last
		last = p

		m := binary.PutUvarint(enc[:], uint64(delta))
		deltas = append(deltas, enc[:m]...)
	}
	return deltas
}

func fromSizedDeltas16(data []byte, ps []uint16) []uint16 {
	sz, m := binary.Uvarint(data)
	data = data[m:]

	if cap(ps) < int(sz) {
		ps = make([]uint16, 0, sz)
	} else {
		ps = ps[:0]
	}

	var last uint16
	for len(data) > 0 {
		delta, m := binary.Uvarint(data)
		offset := last + uint16(delta)
		last = offset
		data = data[m:]
		ps = append(ps, offset)
	}
	return ps
}

func fromDeltas(data []byte, buf []uint32) []uint32 {
	buf = buf[:0]
	if cap(buf) < len(data)/2 {
		buf = make([]uint32, 0, len(data)/2)
	}

	var last uint32
	for len(data) > 0 {
		delta, m := binary.Uvarint(data)
		offset := last + uint32(delta)
		last = offset
		data = data[m:]
		buf = append(buf, offset)
	}
	return buf
}

type runeOffsetCorrection struct {
	runeOffset, byteOffset uint32
}

// runeOffsetMap converts from rune offsets (with granularity runeOffsetFrequency)
// to byte offsets, by tracking only the points where a span of runes is non-ASCII,
// and otherwise interpolating expected byte offsets as one byte per rune.
//
// Instead of storing [100, 205, 305], it stores [{x: 200, y: 205}].
//
// This is very rarely a slight pessimization on repos where there are frequent
// non-ASCII characters.
type runeOffsetMap []runeOffsetCorrection

// makeRuneOffsetMap converts the mostly-predictable runeOffset input
// into a shorter form tracking the unexpected values.
//
// The input is a sequence of y values that we expect to increase by 100 each,
// so we just store (x, y) points where the expectation is violated.
func makeRuneOffsetMap(off []uint32) runeOffsetMap {
	expected := uint32(0)
	tmp := []runeOffsetCorrection{}
	for runeOffset, byteOffset := range off {
		if byteOffset != expected {
			tmp = append(tmp, runeOffsetCorrection{uint32(runeOffset) * runeOffsetFrequency, byteOffset})
			expected = byteOffset
		}
		expected += runeOffsetFrequency
	}
	// copy the slice to ensure it doesn't waste unused trailing capacity
	out := make([]runeOffsetCorrection, len(tmp))
	copy(out, tmp)
	return runeOffsetMap(out)
}

// lookup converts rune index `off` to a byte offset and a number of additional
// runes to traverse, given the granularity of runeOffsetFrequency.
//
// It does this by finding the nearest point to interpolate from in the map.
func (m runeOffsetMap) lookup(runeOffset uint32) (uint32, uint32) {
	left := runeOffset % runeOffsetFrequency
	runeOffset -= left
	slen := len(m)
	if slen == 0 {
		return runeOffset, left
	}
	// sort.Search finds the *first* index for which the predicate is true,
	// but we want to find the *last* index for which the predicate is true.
	// This involves some work to reverse the index directions.
	idx := sort.Search(slen, func(i int) bool {
		return runeOffset >= m[slen-1-i].runeOffset
	})
	idx = slen - 1 - idx
	// idx is now in the range [-1, len(m))-- -1 indicates that the offset is smaller
	// than the first entry in the map, so no correction is necessary.
	byteOff := runeOffset
	if idx >= 0 {
		byteOff = m[idx].byteOffset + runeOffset - m[idx].runeOffset
	}
	return byteOff, left
}

func (m runeOffsetMap) sizeBytes() int {
	return 8 * len(m)
}

func epsilonEqualsOne(scoreWeight float64) bool {
	return scoreWeight == 1 || math.Abs(scoreWeight-1.0) < 1e-9
}
