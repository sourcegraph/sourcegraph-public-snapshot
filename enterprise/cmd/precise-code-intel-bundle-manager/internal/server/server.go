package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/mxk/go-flowrate/flowrate"
	"github.com/sourcegraph/codeintelutils"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const addr = ":3187"

type Server struct {
	bundleDir          string
	observationContext *observation.Context
}

func New(bundleDir string, observationContext *observation.Context) (goroutine.BackgroundRoutine, error) {
	server := &Server{
		bundleDir:          bundleDir,
		observationContext: observationContext,
	}

	return httpserver.NewFromAddr(addr, httpserver.NewHandler(server.setupRoutes), httpserver.Options{})
}

func (s *Server) setupRoutes(router *mux.Router) {
	router.Path("/uploads/{id:[0-9]+}").Methods("GET").HandlerFunc(s.handleGetUpload)
	router.Path("/uploads/{id:[0-9]+}").Methods("POST").HandlerFunc(s.handlePostUpload)
	router.Path("/uploads/{id:[0-9]+}/{index:[0-9]+}").Methods("POST").HandlerFunc(s.handlePostUploadPart)
	router.Path("/uploads/{id:[0-9]+}/stitch").Methods("POST").HandlerFunc(s.handlePostUploadStitch)
	router.Path("/uploads/{id:[0-9]+}").Methods("DELETE").HandlerFunc(s.handleDeleteUpload)
}

// GET /uploads/{id:[0-9]+}
func (s *Server) handleGetUpload(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(paths.UploadFilename(s.bundleDir, idFromRequest(r)))
	if err != nil {
		http.Error(w, "Upload not found.", http.StatusNotFound)
		return
	}
	defer file.Close()

	// If there was a transient error while the worker was trying to access the upload
	// file, it retries but indicates the number of bytes that it has received. We can
	// fast-forward the file to this position and only give the worker the data that it
	// still needs. This technique saves us from having to pre-chunk the file as we must
	// do in the reverse direction.
	if _, err := file.Seek(int64(getQueryInt(r, "seek")), io.SeekStart); err != nil {
		log15.Error("Failed to seek upload file", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(limitTransferRate(w), file); err != nil {
		log15.Error("Failed to write payload to client", "err", err)
	}
}

// POST /uploads/{id:[0-9]+}
func (s *Server) handlePostUpload(w http.ResponseWriter, r *http.Request) {
	_ = s.doUpload(w, r, paths.UploadFilename)
}

// POST /uploads/{id:[0-9]+}/{index:[0-9]+}
func (s *Server) handlePostUploadPart(w http.ResponseWriter, r *http.Request) {
	makeFilename := func(bundleDir string, id int64) string {
		return paths.UploadPartFilename(bundleDir, id, indexFromRequest(r))
	}

	_ = s.doUpload(w, r, makeFilename)
}

// POST /uploads/{id:[0-9]+}/stitch
func (s *Server) handlePostUploadStitch(w http.ResponseWriter, r *http.Request) {
	id := idFromRequest(r)
	filename := paths.UploadFilename(s.bundleDir, id)
	makePartFilename := func(index int) string {
		return paths.UploadPartFilename(s.bundleDir, id, int64(index))
	}

	if err := codeintelutils.StitchFiles(filename, makePartFilename, false, false); err != nil {
		log15.Error("Failed to stitch multipart upload", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = writeFileSize(w, filename)
}

// DELETE /uploads/{id:[0-9]+}
func (s *Server) handleDeleteUpload(w http.ResponseWriter, r *http.Request) {
	s.deleteUpload(w, r)
}

// doUpload writes the HTTP request body to the path determined by the given
// makeFilename function.
func (s *Server) doUpload(w http.ResponseWriter, r *http.Request, makeFilename func(bundleDir string, id int64) string) bool {
	filename := makeFilename(s.bundleDir, idFromRequest(r))

	if err := writeToFile(filename, r.Body); err != nil {
		log15.Error("Failed to write payload", "err", err)
		http.Error(w, fmt.Sprintf("failed to write payload: %s", err.Error()), http.StatusInternalServerError)
		return false
	}

	return writeFileSize(w, filename)
}

func writeToFile(filename string, r io.Reader) (err error) {
	targetFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := targetFile.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	_, err = io.Copy(targetFile, r)
	return err
}

func writeFileSize(w http.ResponseWriter, filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		log15.Error("Failed to stat file", "err", err)
		http.Error(w, fmt.Sprintf("failed to stat file: %s", err.Error()), http.StatusInternalServerError)
		return false
	}

	payload := map[string]int{
		"size": int(fi.Size()),
	}

	writeJSON(w, payload)
	return true
}

func (s *Server) deleteUpload(w http.ResponseWriter, r *http.Request) {
	if err := os.Remove(paths.UploadFilename(s.bundleDir, idFromRequest(r))); err != nil {
		log15.Warn("Failed to delete upload file", "err", err)
	}
}

// limitTransferRate applies a transfer limit to the given writer.
//
// In the case that the remote server is running on the same host as this service, an unbounded
// transfer rate can end up being so fast that we harm our own network connectivity. In order to
// prevent the disruption of other in-flight requests, we cap the transfer rate of w to 1Gbps.
func limitTransferRate(w io.Writer) io.Writer {
	return flowrate.NewWriter(w, 1000*1000*1000)
}

func getQueryInt(r *http.Request, name string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(name))
	return value
}

// idFromRequest returns the database id from the request URL's path. This method
// must only be called from routes containing the `id:[0-9]+` pattern, as the error
// return from ParseInt is not checked.
func idFromRequest(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id
}

func indexFromRequest(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["index"], 10, 64)
	return id
}

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "err", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "err", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}
