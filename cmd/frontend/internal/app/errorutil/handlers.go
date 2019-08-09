// Package errorutil exports a HTTP Middleware for HTTP handlers which return
// errors.
package errorutil

import (
	"fmt"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Handler is a wrapper func for app HTTP handlers that enables app
// error pages.
func Handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error: func(w http.ResponseWriter, req *http.Request, status int, err error) {
			if status < 200 || status >= 400 {
				var spanURL string
				if span := opentracing.SpanFromContext(req.Context()); span != nil {
					ext.Error.Set(span, true)
					span.SetTag("err", err)
					spanURL = trace.SpanURL(span)
				}
				log15.Error("App HTTP handler error response", "method", req.Method, "request_uri", req.URL.RequestURI(), "status_code", status, "error", err, "trace", spanURL)
			}

			w.Header().Set("cache-control", "no-cache")

			var body string
			if env.InsecureDev {
				body = fmt.Sprintf("Error: HTTP %d %s\n\nError: %s", status, http.StatusText(status), err.Error())
			} else {
				body = fmt.Sprintf("Error: HTTP %d: %s", status, http.StatusText(status))
			}
			http.Error(w, body, status)
		},
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_262(size int) error {
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
