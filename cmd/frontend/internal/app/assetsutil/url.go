package assetsutil

import (
	"net/url"
	"path"
)

// baseURL is the path prefix under which static assets should
// be served.
var baseURL = &url.URL{}

// URL returns a URL, possibly relative, to the asset at path
// p.
func URL(p string) *url.URL {
	return baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, p)})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_254(size int) error {
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
