package devdoc

import (
	"net/http"
	"strings"

	"src.sourcegraph.com/sourcegraph/sgx/cli"

	"github.com/sourcegraph/mux"

	"gopkg.in/inconshreveable/log15.v2"
)

// flags are CLI flags exposed on the `src serve` subcommand.
type flags struct {
	Host   string `long:"devdoc.host" description:"hostname to serve developer doc site on (empty for any)"`
	Prefix string `long:"devdoc.prefix" description:"URL path prefix to mount developer doc site ('/' for root)" default:"/.docs/"`
}

var activeFlags flags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("Developer documentation site", "Developer documentation site", &activeFlags)
	})

	cli.ServeMuxFuncs = append(cli.ServeMuxFuncs, func(sm *http.ServeMux) {
		// Ensure that the prefix starts and ends with "/".
		if !strings.HasPrefix(activeFlags.Prefix, "/") {
			activeFlags.Prefix = "/" + activeFlags.Prefix
		}
		if !strings.HasSuffix(activeFlags.Prefix, "/") {
			activeFlags.Prefix = activeFlags.Prefix + "/"
		}

		log15.Debug("Developer documentation site running", "at", activeFlags.Host+activeFlags.Prefix)

		// Attach devdoc handler to main server's ServeMux.
		router := NewRouter(mux.NewRouter().PathPrefix(activeFlags.Prefix).Subrouter())
		sm.Handle(activeFlags.Host+activeFlags.Prefix, New(router))
	})
}
