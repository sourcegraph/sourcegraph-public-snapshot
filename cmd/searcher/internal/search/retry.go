package search

import (
	"os"
	"strings"
)

// getZipFileWithRetry retries getting a zip file if the zip is for some reason
// invalid.
func getZipFileWithRetry(get func() (string, *zipFile, error)) (validPath string, zf *zipFile, err error) {
	var path string
	tries := 0
	for zf == nil {
		path, zf, err = get()
		if err != nil {
			if tries < 2 && strings.Contains(err.Error(), "not a valid zip file") {
				err = os.Remove(path)
				if err != nil {
					return "", nil, err
				}
				tries++
				if tries == 2 {
					return "", nil, err
				}
				continue
			}
			return "", nil, err
		}
	}
	return path, zf, nil
}
