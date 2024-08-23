/* Copyright 2022 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package golang

import (
	"io"
	_ "unsafe"
)

// In order to correctly capture the subtleties of build tag placement
// and to automatically stay up-to-date with the parsing semantics and
// syntax, we link to the stdlib version of header parsing.
//
// Permalink: https://github.com/golang/go/blob/8ed0e51b5e5cc50985444f39dc56c55e4fa3bcf9/src/go/build/build.go#L1568
//go:linkname parseFileHeader go/build.parseFileHeader
func parseFileHeader(_ []byte) ([]byte, []byte, bool, error)

// readComments is like io.ReadAll, except that it only reads the leading
// block of comments in the file.
//
// Permalink: https://github.com/golang/go/blob/8ed0e51b5e5cc50985444f39dc56c55e4fa3bcf9/src/go/build/read.go#L380
//go:linkname readComments go/build.readComments
func readComments(_ io.Reader) ([]byte, error)
