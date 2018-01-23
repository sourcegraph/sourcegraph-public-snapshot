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

// +build !linux

package zoekt

import (
	"fmt"
	"os"
)

// NewIndexFile returns a new index file. The index file takes
// ownership of the passed in file, and may close it.
func NewIndexFile(f *os.File) (IndexFile, error) {
	return &indexFileFromOS{f}, nil
}

type indexFileFromOS struct {
	f *os.File
}

func (f *indexFileFromOS) Read(off, sz uint32) ([]byte, error) {
	r := make([]byte, sz)
	_, err := f.f.ReadAt(r, int64(off))
	return r, err
}

func (f indexFileFromOS) Size() (uint32, error) {
	fi, err := f.f.Stat()
	if err != nil {
		return 0, err
	}

	sz := fi.Size()

	if sz >= maxUInt32 {
		return 0, fmt.Errorf("overflow")
	}

	return uint32(sz), nil
}

func (f indexFileFromOS) Close() {
	f.f.Close()
}

func (f indexFileFromOS) Name() string {
	return f.f.Name()
}
