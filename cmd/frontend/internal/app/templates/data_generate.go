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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_284(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
