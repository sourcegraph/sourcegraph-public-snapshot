// Package runner implements the job runner service.
package runner

import (
	"net/http"

	"github.com/inconshreveable/log15"
)

// Service is the job runner service. It is an http.Handler.
type Service struct {
	// JobFinish is invoked when a job is finished, and should be used to cleanup.
	JobFinish func()

	// Logger where non-job related logs (e.g. "job starting") will go.
	Log log15.Logger
}

// ServeHTTP handles HTTP requests to run jobs.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: handle job execution requests by (a) decoding into protocol.Request type and (b) streaming back protocol.Response ndjson responses.
	// TODO: run the command, capture output, stream back via protocol.Response
	// TODO: If timeout is hit, send final response and exit immediately?
}
