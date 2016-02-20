// package fsync keeps two files or directories in sync.
//
//         err := fsync.Sync("~/dst", ".")
//
// After the above code, if err is nil, every file and directory in the current
// directory is copied to ~/dst and has the same permissions. Consequent calls
// will only copy changed or new files.
//
// SyncTo is a helper function which helps you sync a groups of files or
// directories into a signle destination. For instance, calling
//
//     SyncTo("public", "build/app.js", "build/app.css", "images", "fonts")
//
// is equivalient to calling
//
//     Sync("public/app.js", "build/app.js")
//     Sync("public/app.css", "build/app.css")
//     Sync("public/images", "images")
//     Sync("public/fonts", "fonts")
//
// Actually, this is how SyncTo is implemented: consequent calls to Sync.
//
// By default, sync code ignores extra files in the destination that don’t have
// identicals in the source. Setting Delete field of a Syncer to true changes
// this behavior and deletes these extra files.
package fsync

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/spf13/afero"
)

var (
	ErrFileOverDir = errors.New(
		"fsync: trying to overwrite a non-empty directory with a file")
)

// Sync copies files and directories inside src into dst.
func Sync(dst, src string) error {
	return NewSyncer().Sync(dst, src)
}

// SyncTo syncs srcs files and directories into to directory.
func SyncTo(to string, srcs ...string) error {
	return NewSyncer().SyncTo(to, srcs...)
}

// Type Syncer provides functions for syncing files.
type Syncer struct {
	// Set this to true to delete files in the destination that don't exist
	// in the source.
	Delete bool
	// By default, modification times are synced. This can be turned off by
	// setting this to true.
	NoTimes bool
	// TODO add options for not checking content for equality

	SrcFs  afero.Fs
	DestFs afero.Fs
}

// NewSyncer creates a new instance of Syncer with default options.
func NewSyncer() *Syncer {
	return &Syncer{SrcFs: new(afero.OsFs), DestFs: new(afero.OsFs)}
}

// Sync copies files and directories inside src into dst.
func (s *Syncer) Sync(dst, src string) error {
	// make sure src exists
	if _, err := s.SrcFs.Stat(src); err != nil {
		return err
	}
	// return error instead of replacing a non-empty directory with a file
	if b, err := s.checkDir(dst, src); err != nil {
		return err
	} else if b {
		return ErrFileOverDir
	}

	return s.syncRecover(dst, src)
}

// SyncTo syncs srcs files or directories into to directory.
func (s *Syncer) SyncTo(to string, srcs ...string) error {
	for _, src := range srcs {
		dst := path.Join(to, path.Base(src))
		if err := s.Sync(dst, src); err != nil {
			return err
		}
	}
	return nil
}

// syncRecover handles errors and calls sync
func (s *Syncer) syncRecover(dst, src string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	s.sync(dst, src)
	return nil
}

// sync updates dst to match with src, handling both files and directories.
func (s *Syncer) sync(dst, src string) {
	// sync permissions and modification times after handling content
	defer s.syncstats(dst, src)

	// read files info
	dstat, err := s.SrcFs.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	sstat, err := s.SrcFs.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return // src was deleted before we could copy it
	}
	check(err)

	if !sstat.IsDir() {
		// src is a file
		// delete dst if its a directory
		if dstat != nil && dstat.IsDir() {
			check(s.DestFs.RemoveAll(dst))
		}
		if !s.equal(dst, src) {
			// perform copy
			df, err := s.DestFs.Create(dst)
			check(err)
			defer df.Close()
			sf, err := s.SrcFs.Open(src)
			if os.IsNotExist(err) {
				return
			}
			check(err)
			defer sf.Close()
			_, err = io.Copy(df, sf)
			if os.IsNotExist(err) {
				return
			}
			check(err)
		}
		return
	}

	// src is a directory
	// make dst if necessary
	if dstat == nil {
		// dst does not exist; create directory
		check(s.DestFs.MkdirAll(dst, 0755)) // permissions will be synced later
	} else if !dstat.IsDir() {
		// dst is a file; remove and create directory
		check(s.DestFs.Remove(dst))
		check(s.DestFs.MkdirAll(dst, 0755)) // permissions will be synced later
	}

	// go through sf files and sync them
	files, err := afero.ReadDir(src, s.SrcFs)
	if os.IsNotExist(err) {
		return
	}
	check(err)
	// make a map of filenames for quick lookup; used in deletion
	// deletion below
	m := make(map[string]bool, len(files))
	for _, file := range files {
		dst2 := path.Join(dst, file.Name())
		src2 := path.Join(src, file.Name())
		s.sync(dst2, src2)
		m[file.Name()] = true
	}

	// delete files from dst that does not exist in src
	if s.Delete {
		files, err = afero.ReadDir(dst, s.DestFs)
		check(err)
		for _, file := range files {
			if !m[file.Name()] {
				check(s.DestFs.RemoveAll(path.Join(dst, file.Name())))
			}
		}
	}
}

// syncstats makes sure dst has the same pemissions and modification time as src
func (s *Syncer) syncstats(dst, src string) {
	// get file infos; return if not exist and panic if error
	dstat, err1 := s.DestFs.Stat(dst)
	sstat, err2 := s.SrcFs.Stat(src)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		return
	}
	check(err1)
	check(err2)

	// update dst's permission bits
	if dstat.Mode().Perm() != sstat.Mode().Perm() {
		check(s.DestFs.Chmod(dst, sstat.Mode().Perm()))
		return
	}

	// update dst's modification time
	if !s.NoTimes {
		if !dstat.ModTime().Equal(sstat.ModTime()) {
			err := os.Chtimes(dst, sstat.ModTime(), sstat.ModTime())
			check(err)
		}
	}
}

// equal returns true if both files are equal
func (s *Syncer) equal(a, b string) bool {
	// get file infos
	info1, err1 := os.Stat(a)
	info2, err2 := os.Stat(b)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		return false
	}
	check(err1)
	check(err2)

	// check sizes
	if info1.Size() != info2.Size() {
		return false
	}

	// both have the same size, check the contents
	f1, err := os.Open(a)
	check(err)
	defer f1.Close()
	f2, err := os.Open(b)
	check(err)
	defer f2.Close()
	buf1 := make([]byte, 1000)
	buf2 := make([]byte, 1000)
	for {
		// read from both
		n1, err := f1.Read(buf1)
		if err != nil && err != io.EOF {
			panic(err)
		}
		n2, err := f2.Read(buf2)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// compare read bytes
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false
		}

		// end of both files
		if n1 == 0 && n2 == 0 {
			break
		}
	}

	return true
}

// checkDir returns true if dst is a non-empty directory and src is a file
func (s *Syncer) checkDir(dst, src string) (b bool, err error) {
	// read file info
	dstat, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	sstat, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	// return false is dst is not a directory or src is a directory
	if !dstat.IsDir() || sstat.IsDir() {
		return false, nil
	}

	// dst is a directory and src is a file
	// check if dst is non-empty
	// read dst directory

	files, err := afero.ReadDir(dst, s.DestFs)
	if err != nil {
		return false, err
	}
	if len(files) > 0 {
		return true, nil
	}
	return false, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
