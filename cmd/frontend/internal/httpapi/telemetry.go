package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/eventlogger"
)

var telemetryHandler http.Handler

func init() {
	if envvar.SourcegraphDotComMode() {
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				// Removed due to our event logging ETL pipeline sunsetting schedule.
				// TODO(Dan): update with new logging URL.
			},
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}
	} else {
		telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tr eventlogger.TelemetryRequest
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				log15.Error("telemetryHandler: Decode(2)", "error", err)
			}
			if tr.UserID != 0 && tr.EventLabel == "SavedSearchEmailNotificationSent" {
				err = usagestats.LogActivity(true, tr.UserID, "", "STAGEVERIFY")
				if err != nil {
					log15.Error("telemetryHandler: usagestats.LogActivity", "error", err)
				}
			}

			fmt.Fprintln(w, "event-level telemetry is disabled")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_357(size int) error {
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
