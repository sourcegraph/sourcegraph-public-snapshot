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
	"encoding/binary"
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

func isASCII(in []byte) bool {
	for len(in) > 0 {
		_, sz := utf8.DecodeRune(in)
		if sz > 1 {
			return false
		}
		in = in[sz:]
	}
	return true
}

// compare 'lower' and 'mixed', where 'lower' is thought to be the
// needle.
func caseFoldingEqualsASCII(lower, mixed []byte) bool {
	if len(lower) > len(mixed) {
		return false
	}
	for i, c := range lower {
		d := mixed[i]
		if d >= 'A' && d <= 'Z' {
			d = d - 'A' + 'a'
		}

		if d != c {
			return false
		}
	}
	return true
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
	ngram    ngram
	byteSize uint32 // size of ngram
	byteOff  uint32
	runeOff  uint32
}

func (r runeNgramOff) byteEnd() uint32 {
	return r.byteOff + r.byteSize
}

func splitNGrams(str []byte) []runeNgramOff {
	var runeGram [3]rune
	var off [3]uint32
	var runeCount int

	result := make([]runeNgramOff, 0, len(str))
	var i uint32

	chars := -1
	for len(str) > 0 {
		chars++
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
			ngram:    ng,
			byteSize: i - off[0],
			byteOff:  off[0],
			runeOff:  uint32(chars),
		})
	}
	return result
}

const (
	_classChar  = 0
	_classDigit = iota
	_classPunct = iota
	_classOther = iota
	_classSpace = iota
)

func byteClass(c byte) int {
	if (c >= 'a' && c <= 'z') || c >= 'A' && c <= 'Z' {
		return _classChar
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

func marshalDocSections(secs []DocumentSection) []byte {
	ints := make([]uint32, 0, len(secs)*2)
	for _, s := range secs {
		ints = append(ints, uint32(s.Start), uint32(s.End))
	}

	return toSizedDeltas(ints)
}

func unmarshalDocSections(in []byte, buf []DocumentSection) (secs []DocumentSection) {
	// TODO - ints is unnecessary garbage here.
	ints := fromSizedDeltas(in, nil)
	if cap(buf) >= len(ints)/2 {
		buf = buf[:0]
	} else {
		buf = make([]DocumentSection, 0, len(ints)/2)
	}

	for len(ints) > 0 {
		buf = append(buf, DocumentSection{ints[0], ints[1]})
		ints = ints[2:]
	}
	return buf
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
