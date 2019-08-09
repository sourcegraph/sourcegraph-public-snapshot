package handlerutil

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/csrf"
)

// CSRFMiddleware is HTTP middleware that helps prevent cross-site request forgery. To make your
// forms compliant you will have to submit the token via the X-Csrf-Token header, which is made
// available in the client-side context.
func CSRFMiddleware(next http.Handler, isSecure func() bool) http.Handler {
	type handler struct {
		secure bool
		http.Handler
	}

	newHandler := func(secure bool) handler {
		return handler{secure, csrf.Protect(
			[]byte("e953612ddddcdd5ec60d74e07d40218c"),
			// We do not use the name csrf_token since it is a common name. This
			// leads to conflicts between apps on localhost. See
			// https://github.com/sourcegraph/sourcegraph/issues/65
			csrf.CookieName("sg_csrf_token"),
			csrf.Path("/"),
			csrf.Secure(secure),
		)(next)}
	}

	var v atomic.Value
	var mu sync.Mutex

	v.Store(newHandler(isSecure()))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, secure := v.Load().(handler), isSecure()
		if secure != h.secure {
			mu.Lock()
			// Check if other go-routines didn't get there first.
			if h = v.Load().(handler); h.secure != secure {
				h = newHandler(secure)
				v.Store(h)
			}
			mu.Unlock()
		}
		h.ServeHTTP(w, r)
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_389(size int) error {
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
