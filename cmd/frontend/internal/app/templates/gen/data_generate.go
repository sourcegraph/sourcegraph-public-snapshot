// +build generate

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shurcooL/vfsgen"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates"
)

func main() {
	err := vfsgen.Generate(
		constModTimeFS{FileSystem: templates.Data, modTime: time.Date(2018, 1, 1, 1, 1, 1, 1, time.UTC)},
		vfsgen.Options{
			PackageName:  "templates",
			BuildTags:    "dist",
			VariableName: "Data",
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
}

type constModTimeFS struct {
	modTime time.Time
	http.FileSystem
}

func (f constModTimeFS) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	if file == nil {
		return file, err
	}
	return constModTimeFile{File: file, modTime: f.modTime}, err
}

type constModTimeFile struct {
	modTime time.Time
	http.File
}

func (f constModTimeFile) Readdir(count int) ([]os.FileInfo, error) {
	fi, err := f.File.Readdir(count)
	if fi == nil {
		return fi, err
	}
	mfi := make([]os.FileInfo, len(fi))
	for i, fileInfo := range fi {
		mfi[i] = constModTimeFileInfo{FileInfo: fileInfo, modTime: f.modTime}
	}
	return mfi, err
}

func (f constModTimeFile) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if fi == nil {
		return fi, err
	}
	return constModTimeFileInfo{FileInfo: fi, modTime: f.modTime}, err
}

type constModTimeFileInfo struct {
	modTime time.Time
	os.FileInfo
}

func (f constModTimeFileInfo) ModTime() time.Time {
	return f.modTime
}
