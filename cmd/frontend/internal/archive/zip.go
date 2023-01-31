package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type fileInfo struct {
	path  string
	isDir bool
}

// walk archive root dir to discover all files to zip
func archiveFiles(dir string) ([]fileInfo, error) {
	var files []fileInfo
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		files = append(files, fileInfo{
			path:  path,
			isDir: info.IsDir(),
		})
		return nil
	})
	return files, err
}

func createZip(dir string, files []fileInfo) (io.Reader, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, file := range files {
		path := file.path
		if file.isDir {
			path = fmt.Sprintf("%s%c", path, os.PathSeparator)
		}

		// path in zip must be relative to zip root
		w, err := zw.Create(strings.TrimPrefix(path, dir))
		if err != nil {
			return nil, err
		}

		if !file.isDir {
			file, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			if _, err := io.Copy(w, file); err != nil {
				return nil, err
			}
		}

	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func unZip(path string) (string, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer z.Close()

	parentDir := filepath.Dir(path)
	for _, f := range z.File {
		fp := filepath.Join(parentDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fp, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
			return parentDir, err
		}

		src, err := f.Open()
		if err != nil {
			return parentDir, err
		}
		defer src.Close()

		dst, err := os.Create(fp)
		if err != nil {
			return parentDir, err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return parentDir, err
		}

	}

	return parentDir, nil
}
