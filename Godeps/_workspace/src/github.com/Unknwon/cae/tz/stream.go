// Copyright 2014 Unknown
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package tz

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// A StreamArchive represents a streamable archive.
type StreamArchive struct {
	*tar.Writer
	gw *gzip.Writer
}

func (s *StreamArchive) Close() (err error) {
	if err = s.Writer.Close(); err != nil {
		return err
	}
	return s.gw.Close()
}

// NewStreamArachive returns a new streamable archive with given io.Writer.
// It's caller's responsibility to close io.Writer and streamer after operation.
func NewStreamArachive(w io.Writer) *StreamArchive {
	s := &StreamArchive{}
	s.gw = gzip.NewWriter(w)
	s.Writer = tar.NewWriter(s.gw)
	return s
}

// StreamFile streams a file or directory entry into StreamArchive.
func (s *StreamArchive) StreamFile(relPath string, fi os.FileInfo, data []byte) error {
	if fi.IsDir() {
		fh, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		fh.Name = relPath + "/"
		if err = s.Writer.WriteHeader(fh); err != nil {
			return err
		}
	} else {
		target := ""
		if fi.Mode()&os.ModeSymlink != 0 {
			target = string(data)
		}

		fh, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return err
		}
		fh.Name = filepath.Join(relPath, fi.Name())
		if err = s.Writer.WriteHeader(fh); err != nil {
			return err
		}

		if _, err = s.Writer.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// StreamReader streams data from io.Reader to StreamArchive.
func (s *StreamArchive) StreamReader(relPath string, fi os.FileInfo, r io.Reader) (err error) {
	fh, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	fh.Name = filepath.Join(relPath, fi.Name())
	if err = s.Writer.WriteHeader(fh); err != nil {
		return err
	}
	_, err = io.Copy(s.Writer, r)
	return err
}
