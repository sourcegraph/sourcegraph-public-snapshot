package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grafana/regexp"
)

// SetupRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupRoutes(handler ExecutorHandler, router *mux.Router) {
	subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}", regexp.QuoteMeta(handler.Name()))).Subrouter()
	subRouter.Path("/dequeue").Methods(http.MethodPost).HandlerFunc(handler.HandleDequeue)
	subRouter.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(handler.HandleHeartbeat)
}

// SetupJobRoutes registers all route handlers required for all configured executor
// queues with the given router.
func SetupJobRoutes(handler ExecutorHandler, router *mux.Router) {
	subRouter := router.PathPrefix(fmt.Sprintf("/{queueName:(?:%s)}", regexp.QuoteMeta(handler.Name()))).Subrouter()
	subRouter.Path("/addExecutionLogEntry").Methods(http.MethodPost).HandlerFunc(handler.HandleAddExecutionLogEntry)
	subRouter.Path("/updateExecutionLogEntry").Methods(http.MethodPost).HandlerFunc(handler.HandleUpdateExecutionLogEntry)
	subRouter.Path("/markComplete").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkComplete)
	subRouter.Path("/markErrored").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkErrored)
	subRouter.Path("/markFailed").Methods(http.MethodPost).HandlerFunc(handler.HandleMarkFailed)
}
