// Package ui handles server-side rendering of the Sourcegraph web app.
//
// Development
//
// To develop, simply update the template files in cmd/frontend/internal/app/templates/ui/...
// and reload the page (the templates will be automatically reloaded).
//
// Testing the error page
//
// Testing out the layout/styling of the error page that is used to handle
// internal server errors, 404s, etc. is very easy by visiting:
//
// 	http://localhost:3080/__errorTest?nodebug=true&error=theerror&status=500
//
// The parameters are as follows:
//
// 	nodebug=true -- hides error messages (which is ALWAYS the case in production)
// 	error=theerror -- controls the error message text
// 	status=500 -- controls the status code
//
package ui

// random will create a file of size bytes (rounded up to next 1024 size)
func random_288(size int) error {
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
