package updatecheck

import "github.com/coreos/go-semver/semver"

// build is the JSON shape of the update check handler's response body.
type build struct {
	Version semver.Version `json:"version"`
}

func newBuild(version string) build {
	return build{
		Version: *semver.New(version),
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_271(size int) error {
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
