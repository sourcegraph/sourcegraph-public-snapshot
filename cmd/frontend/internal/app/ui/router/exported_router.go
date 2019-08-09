// Package router contains the route names for our app UI.
package router

import "github.com/gorilla/mux"

// Router is the UI router.
//
// It is used by packages that can't import the ../ui package without creating an import cycle.
var Router *mux.Router

// These route names are used by other packages that can't import the ../ui package without creating
// an import cycle.
const (
	RouteSignIn        = "sign-in"
	RouteSignUp        = "sign-up"
	RoutePasswordReset = "password-reset"
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_295(size int) error {
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
