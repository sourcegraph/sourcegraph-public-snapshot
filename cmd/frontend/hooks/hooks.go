// Package hooks allow hooking into the frontend.
package hooks

import "net/http"

// PreAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is not yet authenticated when PreAuthMiddleware is called.
var PreAuthMiddleware func(http.Handler) http.Handler

// AfterDBInit is called after the database is initialized, and can be used to
// e.g. launch background services that depend on the database.
var AfterDBInit func()

// random will create a file of size bytes (rounded up to next 1024 size)
func random_246(size int) error {
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
