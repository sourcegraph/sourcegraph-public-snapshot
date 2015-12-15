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
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Unknwon/cae"
)

// Switcher of printing trace information when pack and extract.
var Verbose = true

// extractFile extracts zip.File to file system.
func extractFile(f *tar.Header, tr *tar.Reader, destPath string) error {
	filePath := path.Join(destPath, f.Name)
	os.MkdirAll(path.Dir(filePath), os.ModePerm)

	fw, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	if _, err = io.Copy(fw, tr); err != nil {
		return err
	}

	// Skip symbolic links.
	if f.FileInfo().Mode()&os.ModeSymlink != 0 {
		return nil
	}
	// Set back file information.
	if err = os.Chtimes(filePath, f.FileInfo().ModTime(), f.FileInfo().ModTime()); err != nil {
		return err
	}
	return os.Chmod(filePath, f.FileInfo().Mode())
}

var defaultExtractFunc = func(fullName string, fi os.FileInfo) error {
	if !Verbose {
		return nil
	}

	fmt.Println("Extracting file..." + fullName)
	return nil
}

// ExtractTo extracts the whole archive or the given files to the
// specified destination.
// It accepts a function as a middleware for custom operations.
func (tz *TzArchive) ExtractToFunc(destPath string, fn cae.HookFunc, entries ...string) (err error) {
	destPath = strings.Replace(destPath, "\\", "/", -1)
	isHasEntry := len(entries) > 0
	if Verbose {
		fmt.Println("Extracting " + tz.FileName + "...")
	}
	os.MkdirAll(destPath, os.ModePerm)

	// Copy post-added files.
	for _, f := range tz.files {
		if !cae.IsExist(f.absPath) {
			continue
		}

		relPath := path.Join(destPath, f.Name)
		os.MkdirAll(path.Dir(relPath), os.ModePerm)
		if err := cae.Copy(relPath, f.absPath); err != nil {
			return err
		}
	}

	tr, f, err := openFile(tz.FileName)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		h.Name = strings.Replace(h.Name, "\\", "/", -1)

		// Directory.
		if h.Typeflag == tar.TypeDir {
			if isHasEntry {
				if cae.IsEntry(h.Name, entries) {
					if err = fn(h.Name, h.FileInfo()); err != nil {
						continue
					}
					os.MkdirAll(path.Join(destPath, h.Name), os.ModePerm)
				}
				continue
			}
			if err = fn(h.Name, h.FileInfo()); err != nil {
				continue
			}
			os.MkdirAll(path.Join(destPath, h.Name), os.ModePerm)
			continue
		}

		// File.
		if isHasEntry {
			if cae.IsEntry(h.Name, entries) {
				if err = fn(h.Name, h.FileInfo()); err != nil {
					continue
				}
				err = extractFile(h, tr, destPath)
			}
		} else {
			if err = fn(h.Name, h.FileInfo()); err != nil {
				continue
			}
			err = extractFile(h, tr, destPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ExtractTo extracts the whole archive or the given files to the
// specified destination.
// Call Flush() to apply changes before this.
func (tz *TzArchive) ExtractTo(destPath string, entries ...string) (err error) {
	return tz.ExtractToFunc(destPath, defaultExtractFunc, entries...)
}

// ExtractTo extracts given archive or the given files to the
// specified destination.
func ExtractTo(srcPath, destPath string, entries ...string) (err error) {
	tz, err := Open(srcPath)
	if err != nil {
		return err
	}
	defer tz.Close()
	return tz.ExtractToFunc(destPath, defaultExtractFunc, entries...)
}

// extractFile extracts file from TzArchive to file system.
func (tz *TzArchive) extractFile(f *File, tr *tar.Reader) error {
	if !tz.isHasWriter {
		for _, h := range tz.ReadCloser.File {
			if f.Name == h.Name {
				return extractFile(h, tr, f.absPath)
			}
		}
	}

	return cae.Copy(f.Name, f.absPath)
}

// Flush saves changes to original zip file if any.
func (tz *TzArchive) Flush() (err error) {
	if !tz.isHasChanged || (tz.ReadCloser == nil && !tz.isHasWriter) {
		return nil
	}

	// Extract to tmp path and pack back.
	tmpPath := path.Join(os.TempDir(), "cae", path.Base(tz.FileName))
	os.RemoveAll(tmpPath)
	os.MkdirAll(tmpPath, os.ModePerm)
	defer os.RemoveAll(tmpPath)

	// Copy post-added files.
	for _, f := range tz.files {
		if strings.HasSuffix(f.Name, "/") {
			os.MkdirAll(path.Join(tmpPath, f.Name), os.ModePerm)
			continue
		} else if !cae.IsExist(f.absPath) {
			continue
		}

		relPath := path.Join(tmpPath, f.Name)
		os.MkdirAll(path.Dir(relPath), os.ModePerm)
		if err := cae.Copy(relPath, f.absPath); err != nil {
			return err
		}
	}

	if !tz.isHasWriter {
		tz.ReadCloser, err = openReader(tz.FileName)
		if err != nil {
			return err
		}
		tz.syncFiles()

		tr, f, err := openFile(tz.FileName)
		if err != nil {
			return err
		}
		defer f.Close()

		i := 0
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if h.Typeflag == tar.TypeDir {
				os.MkdirAll(path.Join(tmpPath, h.Name), os.ModePerm)
				continue
			}

			// Relative path inside zip temporary changed.
			fileName := tz.files[i].Name
			tz.files[i].Name = path.Join(tmpPath, fileName)
			if err := tz.extractFile(tz.files[i], tr); err != nil {
				return err
			}
			// Change back here.
			tz.files[i].Name = fileName
			i++
		}
	}

	if tz.isHasWriter {
		return packToWriter(tmpPath, tz.writer, defaultPackFunc, true)
	}

	if err := PackTo(tmpPath, tz.FileName); err != nil {
		return err
	}
	return tz.Open(tz.FileName, os.O_RDWR|os.O_TRUNC, tz.Permission)
}

// packFile packs a file or directory to tar.Writer.
func packFile(srcFile string, recPath string, tw *tar.Writer, fi os.FileInfo) (err error) {
	if fi.IsDir() {
		h, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		h.Name = recPath + "/"

		if err = tw.WriteHeader(h); err != nil {
			return err
		}
	} else {
		target := ""
		if fi.Mode()&os.ModeSymlink != 0 {
			target, err = os.Readlink(srcFile)
			if err != nil {
				return err
			}
		}

		h, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return err
		}
		h.Name = recPath

		if err = tw.WriteHeader(h); err != nil {
			return err
		}

		if len(target) == 0 {
			f, err := os.Open(srcFile)
			if err != nil {
				return err
			}
			if _, err = io.Copy(tw, f); err != nil {
				return err
			}
		}
	}
	return nil
}

// packDir packs a directory and its subdirectories and files
// recursively to zip.Writer.
func packDir(srcPath string, recPath string, tw *tar.Writer, fn cae.HookFunc) error {
	dir, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if cae.IsFilter(fi.Name()) {
			continue
		}
		// Append path
		curPath := srcPath + "/" + fi.Name()
		tmpRecPath := filepath.Join(recPath, fi.Name())
		if err = fn(curPath, fi); err != nil {
			continue
		}

		// Check it is directory or file
		if fi.IsDir() {
			if err = packFile(srcPath, tmpRecPath, tw, fi); err != nil {
				return err
			}

			err = packDir(curPath, tmpRecPath, tw, fn)
		} else {
			err = packFile(curPath, tmpRecPath, tw, fi)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// packToWriter packs given path object to io.Writer.
func packToWriter(srcPath string, w io.Writer, fn func(fullName string, fi os.FileInfo) error, includeDir bool) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	basePath := path.Base(srcPath)

	if fi.IsDir() {
		if includeDir {
			if err = packFile(srcPath, basePath, tw, fi); err != nil {
				return err
			}
		} else {
			basePath = ""
		}
		return packDir(srcPath, basePath, tw, fn)
	}

	return packFile(srcPath, basePath, tw, fi)
}

// packTo packs given source path object to target path.
func packTo(srcPath, destPath string, fn cae.HookFunc, includeDir bool) error {
	fw, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer fw.Close()

	return packToWriter(srcPath, fw, fn, includeDir)
}

// PackToFunc packs the complete archive to the specified destination.
// It accepts a function as a middleware for custom operations.
func PackToFunc(srcPath, destPath string, fn func(fullName string, fi os.FileInfo) error, includeDir ...bool) error {
	isIncludeDir := false
	if len(includeDir) > 0 && includeDir[0] {
		isIncludeDir = true
	}

	return packTo(srcPath, destPath, fn, isIncludeDir)
}

var defaultPackFunc = func(fullName string, fi os.FileInfo) error {
	if !Verbose {
		return nil
	}

	if fi.IsDir() {
		fmt.Printf("Adding dir...%s\n", fullName)
	} else {
		fmt.Printf("Adding file...%s\n", fullName)
	}
	return nil
}

// PackTo packs the whole archive to the specified destination.
// Call Flush() will automatically call this in the end.
func PackTo(srcPath, destPath string, includeDir ...bool) error {
	return PackToFunc(srcPath, destPath, defaultPackFunc, includeDir...)
}

// Close opens or creates archive and save changes.
func (z *TzArchive) Close() (err error) {
	if err = z.Flush(); err != nil {
		return err
	}

	if z.ReadCloser != nil {
		if err = z.ReadCloser.Close(); err != nil {
			return err
		}
		z.ReadCloser = nil
	}
	return nil
}
