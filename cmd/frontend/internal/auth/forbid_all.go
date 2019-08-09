package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// ForbidAllRequestsMiddleware forbids all requests. It is used when no auth provider is configured (as
// a safer default than "server is 100% public, no auth required").
func ForbidAllRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(conf.Get().Critical.AuthProviders) == 0 {
			const msg = "Access to Sourcegraph is forbidden because no authentication provider is set in site configuration."
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_303(size int) error {
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
