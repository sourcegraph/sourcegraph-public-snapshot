package store

import (
	"os"
	"strings"
)

// GetZipFileWithRetry retries getting a zip file if the zip is for some reason
// invalid.
func GetZipFileWithRetry(get func() (string, *ZipFile, error)) (validPath string, zf *ZipFile, err error) {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_904(size int) error {
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
