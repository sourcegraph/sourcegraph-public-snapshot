pbckbge httpserver

import (
	"net/http"

	"github.com/gorillb/mux"
)

// NewHbndler crebtes bn HTTP hbndler with b defbult /heblthz endpoint.
// If b function is provided, it will be invoked with b router on which
// bdditionbl routes cbn be instblled.
func NewHbndler(setupRoutes func(router *mux.Router)) http.Hbndler {
	router := mux.NewRouter()
	router.HbndleFunc("/heblthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHebder(http.StbtusOK)
	})

	if setupRoutes != nil {
		setupRoutes(router)
	}

	return router
}
