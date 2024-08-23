// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsontext

import (
	"github.com/go-json-experiment/json/internal/jsonflags"
	"github.com/go-json-experiment/json/internal/jsonwire"
)

var errInvalidUTF8 = &SyntacticError{str: "invalid UTF-8 within string"}

// AppendQuote appends a double-quoted JSON string literal representing src
// to dst and returns the extended buffer.
// It uses the minimal string representation per RFC 8785, section 3.2.2.2.
// Invalid UTF-8 bytes are replaced with the Unicode replacement character
// and an error is returned at the end indicating the presence of invalid UTF-8.
func AppendQuote[Bytes ~[]byte | ~string](dst []byte, src Bytes) ([]byte, error) {
	return jsonwire.AppendQuote(dst, src, &jsonflags.Flags{})
}

// AppendUnquote appends the decoded interpretation of src as a
// double-quoted JSON string literal to dst and returns the extended buffer.
// The input src must be a JSON string without any surrounding whitespace.
// Invalid UTF-8 bytes are replaced with the Unicode replacement character
// and an error is returned at the end indicating the presence of invalid UTF-8.
// Any trailing bytes after the JSON string literal results in an error.
func AppendUnquote[Bytes ~[]byte | ~string](dst []byte, src Bytes) ([]byte, error) {
	return jsonwire.AppendUnquote(dst, src)
}
