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

// Package tz enables you to transparently read or write TAR.GZ compressed archives and the files inside them.
package tz

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Unknwon/cae"
)

// A File represents a file or directory entry in archive.
type File struct {
	*tar.Header
	absPath string
}

// A TzArchive represents a file archive, compressed with Tar and Gzip.
type TzArchive struct {
	*ReadCloser
	FileName   string
	NumFiles   int
	Flag       int
	Permission os.FileMode

	files        []*File
	isHasChanged bool

	// For supporting flushing to io.Writer.
	writer      io.Writer
	isHasWriter bool
}

// OpenFile is the generalized open call; most users will use Open
// instead. It opens the named tar.gz file with specified flag
// (O_RDONLY etc.) if applicable. If successful,
// methods on the returned TzArchive can be used for I/O.
// If there is an error, it will be of type *PathError.
func OpenFile(fileName string, flag int, perm os.FileMode) (*TzArchive, error) {
	tz := new(TzArchive)
	err := tz.Open(fileName, flag, perm)
	return tz, err
}

// Create creates the named tar.gz file, truncating
// it if it already exists. If successful, methods on the returned
// TzArchive can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func Create(fileName string) (*TzArchive, error) {
	os.MkdirAll(path.Dir(fileName), os.ModePerm)
	return OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens the named tar.gz file for reading. If successful, methods on
// the returned TzArchive can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func Open(fileName string) (*TzArchive, error) {
	return OpenFile(fileName, os.O_RDONLY, 0)
}

// New accepts a variable that implemented interface io.Writer
// for write-only purpose operations.
func New(w io.Writer) *TzArchive {
	return &TzArchive{
		writer:      w,
		isHasWriter: true,
	}
}

// List returns a string slice of files' name in TzArchive.
// Specify prefixes will be used as filters.
func (tz *TzArchive) List(prefixes ...string) []string {
	isHasPrefix := len(prefixes) > 0
	names := make([]string, 0, tz.NumFiles)
	for _, f := range tz.files {
		if isHasPrefix && !cae.HasPrefix(f.Name, prefixes) {
			continue
		}
		names = append(names, f.Name)
	}
	return names
}

// AddEmptyDir adds a raw directory entry to TzArchive,
// it returns false if same directory enry already existed.
func (tz *TzArchive) AddEmptyDir(dirPath string) bool {
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	for _, f := range tz.files {
		if dirPath == f.Name {
			return false
		}
	}

	dirPath = strings.TrimSuffix(dirPath, "/")
	if strings.Contains(dirPath, "/") {
		// Auto add all upper level directories.
		tz.AddEmptyDir(path.Dir(dirPath))
	}
	tz.files = append(tz.files, &File{
		Header: &tar.Header{
			Name: dirPath + "/",
		},
	})
	tz.updateStat()
	return true
}

// AddDir adds a directory and subdirectories entries to TzArchive.
func (tz *TzArchive) AddDir(dirPath, absPath string) error {
	dir, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	tz.AddEmptyDir(dirPath)

	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		curPath := strings.Replace(absPath+"/"+fi.Name(), "\\", "/", -1)
		tmpRecPath := strings.Replace(filepath.Join(dirPath, fi.Name()), "\\", "/", -1)
		if fi.IsDir() {
			if err = tz.AddDir(tmpRecPath, curPath); err != nil {
				return err
			}
		} else {
			if err = tz.AddFile(tmpRecPath, curPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateStat should be called after every change for rebuilding statistic.
func (tz *TzArchive) updateStat() {
	tz.NumFiles = len(tz.files)
	tz.isHasChanged = true
}

// AddFile adds a file entry to TzArchive.
func (tz *TzArchive) AddFile(fileName, absPath string) error {
	if cae.IsFilter(absPath) {
		return nil
	}

	si, err := os.Lstat(absPath)
	if err != nil {
		return err
	}

	target := ""
	if si.Mode()&os.ModeSymlink != 0 {
		target, err = os.Readlink(absPath)
		if err != nil {
			return err
		}
	}

	file := new(File)
	file.Header, err = tar.FileInfoHeader(si, target)
	if err != nil {
		return err
	}
	file.Name = fileName
	file.absPath = absPath

	tz.AddEmptyDir(path.Dir(fileName))

	isExist := false
	for _, f := range tz.files {
		if fileName == f.Name {
			f = file
			isExist = true
			break
		}
	}
	if !isExist {
		tz.files = append(tz.files, file)
	}

	tz.updateStat()
	return nil
}

// DeleteIndex deletes an entry in the archive by its index.
func (tz *TzArchive) DeleteIndex(idx int) error {
	if idx >= tz.NumFiles {
		return errors.New("index out of range of number of files")
	}

	tz.files = append(tz.files[:idx], tz.files[idx+1:]...)
	return nil
}

// DeleteName deletes an entry in the archive by its name.
func (tz *TzArchive) DeleteName(name string) error {
	for i, f := range tz.files {
		if f.Name == name {
			return tz.DeleteIndex(i)
		}
	}
	return errors.New("entry with given name not found")
}
