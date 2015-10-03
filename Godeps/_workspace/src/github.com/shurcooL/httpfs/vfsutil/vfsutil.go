// Package vfsutil implements some I/O utility functions for http.FileSystem.
package vfsutil

import (
	"io/ioutil"
	"net/http"
	"os"
)

// ReadDir reads the contents of the directory associated with file and
// returns a slice of FileInfo values in directory order.
func ReadDir(fs http.FileSystem, name string) ([]os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(0)
}

// Stat returns the FileInfo structure describing file.
func Stat(fs http.FileSystem, name string) (os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}

// openStat performs Open and Stat and returns results, or first error encountered.
// The caller is responsible for closing the returned file when done.
func openStat(fs http.FileSystem, name string) (http.File, os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, fi, nil
}

// ReadFile reads the file named by path from fs and returns the contents.
func ReadFile(fs http.FileSystem, path string) ([]byte, error) {
	rc, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}
