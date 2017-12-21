// +build delve

package cli

import (
	"log"
	"net/http"
)

func handleBiLogger(sm *http.ServeMux) {
	// go-kit breaks delve debugging.
	// Make this a no-op when debugging.
	if biLoggerAddr != "" {
		// Warn if this is a debug build and biLoggerAddr is set.
		log.Printf("skipping BI logger for debug build")
	}
}
