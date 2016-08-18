package langp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Server represents all of the Language Processor REST API methods that must
// be implemented by a language processor.
type Server interface {
	Prepare(r *RepoRev) error
	DefSpecToPosition(d *DefSpec) (*Position, error)
	PositionToDefSpec(p *Position) (*DefSpec, error)
	Definition(p *Position) (*Range, error)
	Hover(p *Position) (*Hover, error)
	LocalRefs(p *Position) (*RefLocations, error)
	ExternalRefs(r *RepoRev) (*ExternalRefs, error)
	DefSpecRefs(d *DefSpec) (*RefLocations, error)
	ExportedSymbols(r *RepoRev) (*ExportedSymbols, error)
}

// NewServer returns a new HTTP handler which decodes and encodes JSON
// responses according to the Language Processor REST API by invoking the
// given methods.
//
// Additionally, the server handles sanitization according to the protocol
// (e.g. if a request is missing required JSON fields). It does not handle
// anything more complex such as workspace preparation.
func NewServer(s Server) http.Handler {
	srv := &server{s: s}
	mux := http.NewServeMux()
	mux.Handle("/prepare", handler("/prepare", srv.servePrepare))
	mux.Handle("/definition", handler("/definition", srv.serveDefinition))
	mux.Handle("/exported-symbols", handler("/exported-symbols", srv.serveExportedSymbols))
	mux.Handle("/external-refs", handler("/external-refs", srv.serveExternalRefs))
	mux.Handle("/position-to-defspec", handler("/position-to-defspec", srv.servePositionToDefSpec))
	mux.Handle("/defspec-to-position", handler("/defspec-to-position", srv.serveDefSpecToPosition))
	mux.Handle("/defspec-refs", handler("/defspec-refs", srv.serveDefSpecRefs))
	mux.Handle("/hover", handler("/hover", srv.serveHover))
	mux.Handle("/local-refs", handler("/local-refs", srv.serveLocalRefs))
	return mux
}

type server struct {
	s Server
}

type handlerFunc func(body []byte) (interface{}, error)

// handler returns an HTTP handler which serves the given method at the
// specified path.
func handler(path string, m handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Confirm the path because http.ServeMux is fuzzy.
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// All Language Processor methods are POST and have no query
		// parameters.
		if r.Method != "POST" || len(r.URL.Query()) > 0 {
			resp := &Error{ErrorMsg: "expected POST with no query parameters"}
			writeResponse(w, http.StatusBadRequest, resp, path, nil)
			return
		}

		// Handle the request.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		resp, err := m(body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		writeResponse(w, http.StatusOK, resp, path, body)
	})
}

func writeResponse(w http.ResponseWriter, status int, v interface{}, path string, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(v)
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}
	_, err = io.Copy(w, bytes.NewReader(respBody))
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}

	// TODO: configurable logging
	log.Printf("POST %s -> %d %s\n\tBody:     %s\n\tResponse: %s\n", path, status, http.StatusText(status), string(body), string(respBody))
}

func (s *server) servePrepare(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	return map[string]string{}, s.s.Prepare(&r)
}

func (s *server) serveDefinition(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var p Position
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if p.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if p.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}
	return s.s.Definition(&p)
}

func (s *server) serveExportedSymbols(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	return s.s.ExportedSymbols(&r)
}

func (s *server) serveExternalRefs(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	return s.s.ExternalRefs(&r)
}

func (s *server) servePositionToDefSpec(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var p Position
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if p.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if p.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}
	return s.s.PositionToDefSpec(&p)
}

func (s *server) serveDefSpecToPosition(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var d DefSpec
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, err
	}
	if d.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if d.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if d.Unit == "" {
		return nil, fmt.Errorf("Unit field must be set")
	}
	if d.UnitType == "" {
		return nil, fmt.Errorf("UnitType field must be set")
	}
	if d.Path == "" {
		return nil, fmt.Errorf("Path field must be set")
	}
	return s.s.DefSpecToPosition(&d)
}

func (s *server) serveDefSpecRefs(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var d DefSpec
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, err
	}
	if d.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if d.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if d.Unit == "" {
		return nil, fmt.Errorf("Unit field must be set")
	}
	if d.UnitType == "" {
		return nil, fmt.Errorf("UnitType field must be set")
	}
	if d.Path == "" {
		return nil, fmt.Errorf("Path field must be set")
	}
	return s.s.DefSpecRefs(&d)
}

func (s *server) serveHover(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var p Position
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if p.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if p.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}
	return s.s.Hover(&p)
}

func (s *server) serveLocalRefs(body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var p Position
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if p.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if p.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}
	return s.s.LocalRefs(&p)
}
