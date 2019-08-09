package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var httpTrace, _ = strconv.ParseBool(env.Get("HTTP_TRACE", "false", "dump HTTP requests (including body) to stderr"))

// Trace is an HTTP middleware that dumps the HTTP request body (to stderr) if the env var
// `HTTP_TRACE=1`.
func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpTrace {
			data, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Println("HTTP_TRACE: unable to print request:", err)
			}
			log.Println("====================================================================== HTTP_TRACE: HTTP request")
			log.Println(string(data))
			log.Println("===============================================================================================")
		}
		next.ServeHTTP(w, r)
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_334(size int) error {
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
