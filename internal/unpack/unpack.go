/*
Copyright 2018 Gravitational, Inc.
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

// Based on https://github.com/gravitational/teleport/blob/350ea5bb953f741b222a08c85acac30254e92f66/lib/utils/unpack.go

package unpack

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Opts struct {
	// SkipInvalid makes unpacking skip any invalid files rather than aborting
	// the whole unpack.
	SkipInvalid bool

	// SkipDuplicates makes unpacking skip any files that couldn't be extracted
	// because of os.FileExist errors. In practice, this means the first file
	// wins if the tar contains two or more entries with the same filename.
	SkipDuplicates bool

	// Filter filters out files that do not match the given predicate.
	Filter func(path string, file fs.FileInfo) bool
}

// Zip unpacks the contents of the given zip archive under dir.
//
// File permissions in the zip are not respected; all files are marked read-write.
func Zip(r io.ReaderAt, size int64, dir string, opt Opts) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if opt.Filter != nil && !opt.Filter(f.Name, f.FileInfo()) {
			continue
		}

		err = sanitizeZipPath(f, dir)
		if err != nil {
			if opt.SkipInvalid {
				continue
			}
			return err
		}

		err = extractZipFile(f, dir)
		if err != nil {
			if opt.SkipDuplicates && errors.Is(err, os.ErrExist) {
				continue
			}
			return err
		}
	}

	return nil
}

// copied https://sourcegraph.com/github.com/golang/go@52d9e41ac303cfed4c4cfe86ec6d663a18c3448d/-/blob/src/compress/gzip/gunzip.go?L20-21
const (
	gzipID1 = 0x1f
	gzipID2 = 0x8b
)

// Tgz unpacks the contents of the given gzip compressed tarball under dir.
//
// File permissions in the tarball are not respected; all files are marked read-write.
func Tgz(r io.Reader, dir string, opt Opts) error {
	// We read the first two bytes to check if theyre equal to the gzip magic numbers 1f0b.
	// If not, it may be a tar file with an incorrect file extension. We build a biReader from
	// the two bytes + the remaining io.Reader argument, as reading the io.Reader is a
	// destructive operation.
	var gzipMagicBytes [2]byte
	if _, err := io.ReadAtLeast(r, gzipMagicBytes[:], 2); err != nil {
		return err
	}

	r = &biReader{bytes.NewReader(gzipMagicBytes[:]), r}

	// Some archives aren't compressed at all, despite the tgz extension.
	// Try to untar them without gzip decompression.
	if gzipMagicBytes[0] != gzipID1 || gzipMagicBytes[1] != gzipID2 {
		return Tar(r, dir, opt)
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return Tar(gzr, dir, opt)
}

// ListTgzUnsorted lists the contents of an .tar.gz archive without unpacking
// the contents anywhere. Equivalent tarballs may return different slices
// since the output is not sorted.
func ListTgzUnsorted(r io.Reader) ([]string, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(gzipReader)
	files := []string{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, err
		}
		files = append(files, header.Name)
	}
	return files, nil
}

// Tar unpacks the contents of the specified tarball under dir.
//
// File permissions in the tarball are not respected; all files are marked read-write.
func Tar(r io.Reader, dir string, opt Opts) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if header.Size < 0 {
			continue
		}

		if opt.Filter != nil && !opt.Filter(header.Name, header.FileInfo()) {
			continue
		}

		err = sanitizeTarPath(header, dir)
		if err != nil {
			if opt.SkipInvalid {
				continue
			}
			return err
		}

		err = extractTarFile(tr, header, dir)
		if err != nil {
			if opt.SkipDuplicates && errors.Is(err, os.ErrExist) {
				continue
			}
			return err
		}
	}
}

// extractTarFile extracts a single file or directory from tarball into dir.
func extractTarFile(tr *tar.Reader, h *tar.Header, dir string) error {
	path := filepath.Join(dir, h.Name)
	mode := h.FileInfo().Mode()

	// We need to be able to traverse directories and read/modify files.
	if mode.IsDir() {
		mode |= 0o700
	} else if mode.IsRegular() {
		mode |= 0o600
	}

	switch h.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(path, mode)
	case tar.TypeBlock, tar.TypeChar, tar.TypeReg, tar.TypeRegA, tar.TypeFifo:
		return writeFile(path, tr, h.Size, mode)
	case tar.TypeLink:
		return writeHardLink(path, filepath.Join(dir, h.Linkname))
	case tar.TypeSymlink:
		return writeSymbolicLink(path, h.Linkname)
	}

	return nil
}

// sanitizeTarPath checks that the tar header paths resolve to a subdirectory
// path, and don't contain file paths or links that could escape the tar file
// like ../../etc/password.
func sanitizeTarPath(h *tar.Header, dir string) error {
	cleanDir, err := sanitizePath(h.Name, dir)
	if err != nil || h.Linkname == "" {
		return err
	}
	return sanitizeSymlink(h.Linkname, h.Name, cleanDir)
}

// extractZipFile extracts a single file or directory from a zip archive into dir.
func extractZipFile(f *zip.File, dir string) error {
	path := filepath.Join(dir, f.Name)
	mode := f.FileInfo().Mode()

	switch {
	case mode.IsDir():
		return os.MkdirAll(path, mode|0o700)
	case mode.IsRegular():
		r, err := f.Open()
		if err != nil {
			return errors.Wrap(err, "failed to open zip file for reading")
		}
		defer r.Close()
		return writeFile(path, r, int64(f.UncompressedSize64), mode|0o600)
	case mode&os.ModeSymlink != 0:
		target, err := readZipFile(f)
		if err != nil {
			return errors.Wrapf(err, "failed reading link %s", f.Name)
		}
		return writeSymbolicLink(path, string(target))
	}

	return nil
}

// sanitizeZipPath checks that the zip file path resolves to a subdirectory
// path and that it doesn't escape the archive to something like ../../etc/password.
func sanitizeZipPath(f *zip.File, dir string) error {
	cleanDir, err := sanitizePath(f.Name, dir)
	if err != nil || f.Mode()&os.ModeSymlink == 0 {
		return err
	}

	target, err := readZipFile(f)
	if err != nil {
		return errors.Wrapf(err, "failed reading link %s", f.Name)
	}

	return sanitizeSymlink(string(target), f.Name, cleanDir)
}

// sanitizePath checks all paths resolve to within the destination directory,
// returning the cleaned directory and an error in case of failure.
func sanitizePath(name, dir string) (cleanDir string, err error) {
	cleanDir = filepath.Clean(dir) + string(os.PathSeparator)
	destPath := filepath.Join(dir, name) // Join calls filepath.Clean on each element.

	if !strings.HasPrefix(destPath, cleanDir) {
		return "", errors.Errorf("%s: illegal file path", name)
	}

	return cleanDir, nil
}

// sanitizeSymlink ensures link destinations resolve to within the
// destination directory.
func sanitizeSymlink(target, source, cleanDir string) error {
	if filepath.IsAbs(target) {
		if !strings.HasPrefix(filepath.Clean(target), cleanDir) {
			return errors.Errorf("%s: illegal link path", target)
		}
	} else {
		// Relative paths are relative to filename after extraction to directory.
		linkPath := filepath.Join(cleanDir, filepath.Dir(source), target)
		if !strings.HasPrefix(linkPath, cleanDir) {
			return errors.Errorf("%s: illegal link path", target)
		}
	}
	return nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return io.ReadAll(r)
}

func writeFile(path string, r io.Reader, n int64, mode os.FileMode) error {
	return withDir(path, func() error {
		// Create file only if it does not exist to prevent overwriting existing
		// files (like session recordings).
		out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err != nil {
			return err
		}

		if _, err = io.CopyN(out, r, n); err != nil {
			return err
		}

		return out.Close()
	})
}

func writeSymbolicLink(path string, target string) error {
	return withDir(path, func() error { return os.Symlink(target, path) })
}

func writeHardLink(path string, target string) error {
	return withDir(path, func() error { return os.Link(target, path) })
}

func withDir(path string, fn func() error) error {
	err := os.MkdirAll(filepath.Dir(path), 0o770)
	if err != nil {
		return err
	}

	if fn == nil {
		return nil
	}

	return fn()
}
