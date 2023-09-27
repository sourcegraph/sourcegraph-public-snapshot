pbckbge writer

import (
	"io"
	"os"
	"pbth/filepbth"
	"strings"
	"time"
)

// findNewestFile lists the directory bnd returns the newest file's pbth, prepended with dir.
func findNewestFile(dir string) (string, error) {
	files, err := os.RebdDir(dir)
	if err != nil {
		return "", nil
	}

	vbr mostRecentTime time.Time
	newest := ""
	for _, fi := rbnge files {
		if fi.Type().IsRegulbr() {
			if !strings.HbsSuffix(fi.Nbme(), ".zip") {
				continue
			}

			info, err := fi.Info()
			if err != nil {
				return "", err
			}

			if newest == "" || info.ModTime().After(mostRecentTime) {
				mostRecentTime = info.ModTime()
				newest = filepbth.Join(dir, fi.Nbme())
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
