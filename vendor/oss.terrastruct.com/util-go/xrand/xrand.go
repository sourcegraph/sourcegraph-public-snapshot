// Package xrand provides helpers for generating useful random values.
package xrand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	math_rand "math/rand"
	"strings"
	"time"
	"unicode/utf8"
)

func init() {
	math_rand.Seed(time.Now().UnixNano())
}

// 0 is a valid unicode codepoint but certain programs
// won't allow 0 to avoid clashing with c strings. e.g. pq.
func randRune() rune {
	for {
		if Bool() {
			// Generate plain ASCII half the time.
			if math_rand.Int31n(100) == 0 {
				// Generate newline 1% of the time.
				return '\n'
			}
			return math_rand.Int31n(128) + 1
		}
		r := math_rand.Int31n(utf8.MaxRune+1) + 1
		if utf8.ValidRune(r) {
			return r
		}
	}
}

func String(n int, exclude []rune) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		r := randRune()
		excluded := false
		for _, xr := range exclude {
			if r == xr {
				excluded = true
				break
			}
		}
		if excluded {
			i--
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func Bool() bool {
	return math_rand.Intn(2) == 0
}

func Base64(n int) string {
	var b strings.Builder
	b.Grow(n)
	wc := base64.NewEncoder(base64.URLEncoding, &b)

	_, err := io.CopyN(wc, rand.Reader, int64(n))
	if err != nil {
		panic(fmt.Sprintf("error encoding base64: %v", err))
	}

	err = wc.Close()
	if err != nil {
		panic(fmt.Sprintf("error encoding base64: %v", err))
	}

	return b.String()
}
