package writer

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// findNewestFile lists the directory and returns the newest file's path, prepended with dir.
func findNewestFile(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", nil
	}

	var mostRecentTime time.Time
	newest := ""
	for _, fi := range files {
		if fi.Type().IsRegular() {
			if !strings.HasSuffix(fi.Name(), ".zip") {
				continue
			}

			info, err := fi.Info()
			if err != nil {
				return "", err
			}

			if newest == "" || info.ModTime().After(mostRecentTime) {
				mostRecentTime = info.ModTime()
				newest = filepath.Join(dir, fi.Name())
			}
		}
	}

	return newest, nil
}

func copyFile(from string, to string) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer toFile.Close()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		return err
	}

	return nil
}
