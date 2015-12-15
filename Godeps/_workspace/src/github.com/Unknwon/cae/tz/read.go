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
	"strings"
)

// A ReadCloser represents a caller closable file reader.
type ReadCloser struct {
	f    *os.File
	File []*tar.Header
}

// Close closes the tar.gz file, rendering it unusable for I/O.
func (rc *ReadCloser) Close() error {
	return rc.f.Close()
}

// openFile opens a tar.gz file with gzip and tar decoders.
func openFile(name string) (*tar.Reader, *os.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	return tar.NewReader(gr), f, nil
}

// openReader opens the tar.gz file specified by name and return a tar.Reader.
func openReader(name string) (*ReadCloser, error) {
	tr, f, err := openFile(name)
	if err != nil {
		return nil, err
	}

	r := new(ReadCloser)
	if err := r.init(tr); err != nil {
		return nil, err
	}
	r.f = f
	return r, nil
}

// init initializes a new ReadCloser.
func (rc *ReadCloser) init(r *tar.Reader) error {
	defer rc.Close()

	rc.File = make([]*tar.Header, 0, 10)
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		rc.File = append(rc.File, h)
	}
	return nil
}

// syncFiles syncs file information from file system to memroy object.
func (tz *TzArchive) syncFiles() {
	tz.files = make([]*File, tz.NumFiles)
	for i, f := range tz.File {
		tz.files[i] = &File{}
		tz.files[i].Header = f
		tz.files[i].Name = strings.Replace(f.Name, "\\", "/", -1)
		if f.FileInfo().IsDir() && !strings.HasSuffix(tz.files[i].Name, "/") {
			tz.files[i].Name += "/"
		}
	}
}

// Open is the generalized open call; most users will use Open
// instead. It opens the named tar.gz file with specified flag
// (O_RDONLY etc.) if applicable. If successful,
// methods on the returned TzArchive can be used for I/O.
// If there is an error, it will be of type *PathError.
func (tz *TzArchive) Open(name string, flag int, perm os.FileMode) error {
	if flag&os.O_CREATE != 0 {
		fw, err := os.Create(name)
		if err != nil {
			return err
		}

		gw := gzip.NewWriter(fw)
		tw := tar.NewWriter(gw)
		if err = tw.Close(); err != nil {
			return err
		} else if err = gw.Close(); err != nil {
			return err
		} else if err = fw.Close(); err != nil {
			return err
		}
	}

	rc, err := openReader(name)
	if err != nil {
		return err
	}

	tz.ReadCloser = rc
	tz.FileName = name
	tz.NumFiles = len(rc.File)
	tz.Flag = flag
	tz.Permission = perm
	tz.isHasChanged = false

	tz.syncFiles()
	return nil
}
