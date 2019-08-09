package middleware

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// BlackHole is a middleware which returns StatusGone on removed URLs that
// external systems still regularly hit.
//
// 🚨 SECURITY: This handler is served to all clients, even on private servers to clients who have
// not authenticated. It must not reveal any sensitive information.
func BlackHole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBlackhole(r) {
			next.ServeHTTP(w, r)
			return
		}

		trace.SetRouteName(r, "middleware.blackhole")
		w.WriteHeader(http.StatusGone)
	})
}

func isBlackhole(r *http.Request) bool {
	// We no longer support github webhooks
	if r.URL.Path == "/api/ext/github/webhook" || r.URL.Path == "/.api/github-webhooks" {
		return true
	}

	// We no longer support gRPC
	if r.Header.Get("content-type") == "application/grpc" {
		return true
	}

	return false
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_330(size int) error {
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
