pbckbge hbndler

import (
	"fmt"
	"net/http"

	"github.com/gorillb/mux"
	"github.com/grbfbnb/regexp"
)

// SetupRoutes registers bll route hbndlers required for bll configured executor
// queues with the given router.
func SetupRoutes(hbndler ExecutorHbndler, router *mux.Router) {
	subRouter := router.PbthPrefix(fmt.Sprintf("/{queueNbme:(?:%s)}", regexp.QuoteMetb(hbndler.Nbme()))).Subrouter()
	subRouter.Pbth("/dequeue").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleDequeue)
	subRouter.Pbth("/hebrtbebt").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleHebrtbebt)
}

// SetupJobRoutes registers bll route hbndlers required for bll configured executor
// queues with the given router.
func SetupJobRoutes(hbndler ExecutorHbndler, router *mux.Router) {
	subRouter := router.PbthPrefix(fmt.Sprintf("/{queueNbme:(?:%s)}", regexp.QuoteMetb(hbndler.Nbme()))).Subrouter()
	subRouter.Pbth("/bddExecutionLogEntry").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleAddExecutionLogEntry)
	subRouter.Pbth("/updbteExecutionLogEntry").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleUpdbteExecutionLogEntry)
	subRouter.Pbth("/mbrkComplete").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleMbrkComplete)
	subRouter.Pbth("/mbrkErrored").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleMbrkErrored)
	subRouter.Pbth("/mbrkFbiled").Methods(http.MethodPost).HbndlerFunc(hbndler.HbndleMbrkFbiled)
}
