package langp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

type contextKey int

const (
	methodNameKey    contextKey = 1
	authorizationKey contextKey = 2
	GitHubTokenKey   contextKey = 3
)

// Server represents all of the Language Processor REST API methods that must
// be implemented by a language processor.
type Server interface {
	Prepare(ctx context.Context, r *RepoRev) error
	Definition(ctx context.Context, p *Position) (*Range, error)
	Hover(ctx context.Context, p *Position) (*Hover, error)
	LocalRefs(ctx context.Context, p *Position) (*RefLocations, error)
	ExternalRefs(ctx context.Context, r *RepoRev) (*ExternalRefs, error)
	Symbols(ctx context.Context, opt *SymbolsQuery) (*Symbols, error)
	ExportedSymbols(ctx context.Context, r *RepoRev) (*ExportedSymbols, error)
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
	mux.Handle("/symbols", handler("/symbols", srv.serveSymbols))
	mux.Handle("/exported-symbols", handler("/exported-symbols", srv.serveExportedSymbols))
	mux.Handle("/external-refs", handler("/external-refs", srv.serveExternalRefs))
	mux.Handle("/hover", handler("/hover", srv.serveHover))
	mux.Handle("/local-refs", handler("/local-refs", srv.serveLocalRefs))
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok\n"))
	}))
	return mux
}

type server struct {
	s Server
}

type handlerFunc func(ctx context.Context, body []byte) (interface{}, error)

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

		parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			log.Println("could not extract opentracing headers:", err)
		}
		span := opentracing.StartSpan("LP Serve: "+path, opentracing.ChildOf(parentSpanCtx))
		defer span.Finish()

		// Handle the request.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), authorizationKey, r.Header.Get("Authorization")))
		r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
		r = r.WithContext(context.WithValue(r.Context(), methodNameKey, path))
		resp, err := m(r.Context(), body)
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
}

func (s *server) servePrepare(ctx context.Context, body []byte) (interface{}, error) {
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
	return map[string]string{}, s.s.Prepare(ctx, &r)
}

func (s *server) serveDefinition(ctx context.Context, body []byte) (interface{}, error) {
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
	return s.s.Definition(ctx, &p)
}

func (s *server) serveSymbols(ctx context.Context, body []byte) (interface{}, error) {
	// Decode the user request and ensure that required fields are specified.
	var opt SymbolsQuery
	if err := json.Unmarshal(body, &opt); err != nil {
		return nil, err
	}
	if opt.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if opt.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	return s.s.Symbols(ctx, &opt)
}

func (s *server) serveExportedSymbols(ctx context.Context, body []byte) (interface{}, error) {
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
	return s.s.ExportedSymbols(ctx, &r)
}

func (s *server) serveExternalRefs(ctx context.Context, body []byte) (interface{}, error) {
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
	return s.s.ExternalRefs(ctx, &r)
}

func (s *server) serveHover(ctx context.Context, body []byte) (interface{}, error) {
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
	return s.s.Hover(ctx, &p)
}

func (s *server) serveLocalRefs(ctx context.Context, body []byte) (interface{}, error) {
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
	return s.s.LocalRefs(ctx, &p)
}
