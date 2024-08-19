// Copyright 2021, 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.16
// +build go1.16

package runfiles

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"time"
)

// Open implements fs.FS.Open.
func (r *Runfiles) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	p, err := r.Rlocation(name)
	if errors.Is(err, ErrEmpty) {
		return emptyFile(name), nil
	}
	if err != nil {
		return nil, pathError("open", name, err)
	}
	return os.Open(p)
}

// Stat implements fs.StatFS.Stat.
func (r *Runfiles) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}
	p, err := r.Rlocation(name)
	if errors.Is(err, ErrEmpty) {
		return emptyFileInfo(name), nil
	}
	if err != nil {
		return nil, pathError("stat", name, err)
	}
	return os.Stat(p)
}

// ReadFile implements fs.ReadFileFS.ReadFile.
func (r *Runfiles) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	p, err := r.Rlocation(name)
	if errors.Is(err, ErrEmpty) {
		return nil, nil
	}
	if err != nil {
		return nil, pathError("open", name, err)
	}
	return os.ReadFile(p)
}

type emptyFile string

func (f emptyFile) Stat() (fs.FileInfo, error) { return emptyFileInfo(f), nil }
func (f emptyFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (emptyFile) Close() error                 { return nil }

type emptyFileInfo string

func (i emptyFileInfo) Name() string     { return string(i) }
func (emptyFileInfo) Size() int64        { return 0 }
func (emptyFileInfo) Mode() fs.FileMode  { return 0444 }
func (emptyFileInfo) ModTime() time.Time { return time.Time{} }
func (emptyFileInfo) IsDir() bool        { return false }
func (emptyFileInfo) Sys() interface{}   { return nil }

func pathError(op, name string, err error) error {
	if err == nil {
		return nil
	}
	var rerr Error
	if errors.As(err, &rerr) {
		// Unwrap the error because we don’t need the failing name
		// twice.
		return &fs.PathError{Op: op, Path: rerr.Name, Err: rerr.Err}
	}
	return &fs.PathError{Op: op, Path: name, Err: err}
}
