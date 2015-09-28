// Go Port:
// Copyright (c) 2013 Dmitry Chestnykh <dmitry@codingrobots.com>
//
// Original:
// Copyright (c) 2008 Ryan Grove <ryan@wonko.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//   * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//   * Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//   * Neither the name of this project nor the names of its contributors may be
//     used to endorse or promote products derived from this software without
//     specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package cssmin minifies CSS. It's a port of Ryan Grove's cssmin from Ruby.
package cssmin

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

var (
	rcomments      = regexp.MustCompile(`\/\*[\s\S]*?\*\/`)
	rwhitespace    = regexp.MustCompile(`\s+`)
	rbmh           = regexp.MustCompile(`"\\"\}\\""`)
	runspace1      = regexp.MustCompile(`(?:^|\})[^\{:]+\s+:+[^\{]*\{`)
	runspace2      = regexp.MustCompile(`\s+([!\{\};:>+\(\)\],])`)
	runspace3      = regexp.MustCompile(`([!\{\}:;>+\(\[,])\s+`)
	rsemicolons    = regexp.MustCompile(`([^;\}])\}`)
	runits         = regexp.MustCompile(`(?i)([\s:])([+-]?0)(?:%|em|ex|px|in|cm|mm|pt|pc)`)
	rfourzero      = regexp.MustCompile(`:(?:0 )+0;`)
	rleadzero      = regexp.MustCompile(`(:|\s)0+\.(\d+)`)
	rrgb           = regexp.MustCompile(`rgb\s*\(\s*([0-9,\s]+)\s*\)`)
	rdigits        = regexp.MustCompile(`\d+`)
	rcompresshex   = regexp.MustCompile(`(?i)([^"'=\s])(\s?)\s*#([0-9a-f]){6}`)
	rhexval        = regexp.MustCompile(`[0-9a-f]{2}`)
	remptyrules    = regexp.MustCompile(`[^\}]+\{;\}\n`)
	rmediaspace    = regexp.MustCompile(`\band\(`)
	rredsemicolons = regexp.MustCompile(`;+\}`)
)

func Minify(css []byte) (minified []byte) {
	// Remove comments.
	css = rcomments.ReplaceAll(css, []byte{})

	// Compress all runs of whitespace to a single space to make things easier
	// to work with.
	css = rwhitespace.ReplaceAll(css, []byte(" "))

	// Replace box model hacks with placeholders.
	css = rbmh.ReplaceAll(css, []byte("___BMH___"))

	// Remove unnecessary spaces, but be careful not to turn "p :link {...}"
	// into "p:link{...}".
	css = runspace1.ReplaceAllFunc(css, func(match []byte) []byte {
		return bytes.Replace(match, []byte(":"), []byte("___PSEUDOCLASSCOLON___"), -1)
	})
	css = runspace2.ReplaceAll(css, []byte("$1"))
	css = bytes.Replace(css, []byte("___PSEUDOCLASSCOLON___"), []byte(":"), -1)
	css = runspace3.ReplaceAll(css, []byte("$1"))

	// Add missing semicolons.
	css = rsemicolons.ReplaceAll(css, []byte("$1;}"))

	// Replace 0(%, em, ex, px, in, cm, mm, pt, pc) with just 0.
	css = runits.ReplaceAll(css, []byte("$1$2"))

	// Replace 0 0 0 0; with 0.
	css = rfourzero.ReplaceAll(css, []byte(":0;"))

	// Replace background-position:0; with background-position:0 0;
	css = bytes.Replace(css, []byte("background-position:0;"), []byte("background-position:0 0;"), -1)

	// Replace 0.6 with .6, but only when preceded by : or a space.
	css = rleadzero.ReplaceAll(css, []byte("$1.$2"))

	// Convert rgb color values to hex values.
	css = rrgb.ReplaceAllFunc(css, func(match []byte) (out []byte) {
		out = []byte{'#'}
		for _, v := range rdigits.FindAll(match, -1) {
			d, err := strconv.Atoi(string(v))
			if err != nil {
				return match
			}
			out = append(out, []byte(fmt.Sprintf("%02x", d))...)
		}
		return out
	})

	// Compress color hex values, making sure not to touch values used in IE
	// filters, since they would break.
	css = rcompresshex.ReplaceAllFunc(css, func(match []byte) (out []byte) {
		vals := rhexval.FindAll(match, -1)
		if len(vals) != 3 {
			return match
		}
		compressible := true
		for _, v := range vals {
			if v[0] != v[1] {
				compressible = false
			}
		}
		if !compressible {
			return match
		}
		out = append(out, match[:bytes.IndexByte(match, '#')+1]...)
		return append(out, vals[0][0], vals[1][0], vals[2][0])
	})

	// Remove empty rules.
	css = remptyrules.ReplaceAll(css, []byte{})

	// Re-insert box model hacks.
	css = bytes.Replace(css, []byte("___BMH___"), []byte(`"\"}\""`), -1)

	// Put the space back in for media queries
	css = rmediaspace.ReplaceAll(css, []byte("and ("))

	// Prevent redundant semicolons.
	css = rredsemicolons.ReplaceAll(css, []byte("}"))

	return bytes.TrimSpace(css)
}
