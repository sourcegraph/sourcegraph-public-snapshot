package unidecode

import (
	"compress/zlib"
	"io"
	"strings"
)

const (
	dummyLenght = byte(0xff)
)

var (
	transliterations [65536][]rune
	transCount       = rune(len(transliterations))
)

func decodeTransliterations() {
	r, err := zlib.NewReader(strings.NewReader(tableData))
	if err != nil {
		panic(err)
	}
	defer r.Close()
	b := make([]byte, 0, 13) // 13 = longest transliteration, adjust if needed
	lenB := b[:1]
	chr := uint16(0xffff) // char counter, rely on overflow on first pass
	for {
		chr++
		if _, err := io.ReadFull(r, lenB); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if lenB[0] == dummyLenght {
			continue
		}
		b = b[:lenB[0]] // resize, preserving allocation
		if _, err := io.ReadFull(r, b); err != nil {
			panic(err)
		}
		transliterations[int(chr)] = []rune(string(b))
	}
}
