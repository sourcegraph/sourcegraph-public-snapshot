// +build !delve

package cli

import (
	"io/ioutil"
	"net/http"
	"os"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/util/conn"
)

func handleBiLogger(sm *http.ServeMux) {
	if biLoggerAddr == "" {
		return
	}
	logger := gokitlog.NewLogfmtLogger(os.Stdout)
	biLogger := conn.NewDefaultManager("tcp", biLoggerAddr, logger)
	sm.HandleFunc("/.bi-logger/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Log("component", "bi-logger", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if _, err := biLogger.Write(append(body, '\n')); err != nil {
			logger.Log("component", "bi-logger", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("OK"))
	})
}
