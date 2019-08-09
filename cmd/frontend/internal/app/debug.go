package app

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router) {
	for _, svc := range debugserver.Services {
		svc := svc
		prefix := "/" + svc.Name
		r.PathPrefix(prefix).Handler(adminOnly(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = svc.Host
				if i := strings.Index(req.URL.Path, prefix); i >= 0 {
					req.URL.Path = req.URL.Path[i+len(prefix):]
				}
			},
			ErrorLog: log.New(env.DebugOut, fmt.Sprintf("%s debug proxy: ", svc.Name), log.LstdFlags),
		}))
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, svc := range debugserver.Services {
			path := "/"
			if svc.DefaultPath != "" {
				path = svc.DefaultPath
			}
			fmt.Fprintf(w, `<a href="%s%s">%s</a><br>`, svc.Name, path, svc.Name)
		}
		fmt.Fprintf(w, `<a href="headers">headers</a><br>`)

		// We do not support cluster deployments yet.
		if len(debugserver.Services) == 0 {
			fmt.Fprintf(w, `Instrumentation endpoint proxying for Sourcegraph cluster deployments is not yet available<br>`)
		}
	})
	r.Handle("/", adminOnly(index))
}

// adminOnly is a HTTP middleware which only allows requests by admins.
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_258(size int) error {
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
