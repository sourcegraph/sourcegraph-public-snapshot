package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Server struct {
	host                 string
	port                 int
	bundleDir            string
	databaseCache        *database.DatabaseCache
	documentDataCache    *database.DocumentDataCache
	resultChunkDataCache *database.ResultChunkDataCache
}

type ServerOpts struct {
	Host                     string
	Port                     int
	BundleDir                string
	DatabaseCacheSize        int64
	DocumentDataCacheSize    int64
	ResultChunkDataCacheSize int64
}

func New(opts ServerOpts) (*Server, error) {
	databaseCache, err := database.NewDatabaseCache(opts.DatabaseCacheSize)
	if err != nil {
		return nil, err
	}

	documentDataCache, err := database.NewDocumentDataCache(opts.DocumentDataCacheSize)
	if err != nil {
		return nil, err
	}

	resultChunkDataCache, err := database.NewResultChunkDataCache(opts.ResultChunkDataCacheSize)
	if err != nil {
		return nil, err
	}

	return &Server{
		host:                 opts.Host,
		port:                 opts.Port,
		bundleDir:            opts.BundleDir,
		databaseCache:        databaseCache,
		documentDataCache:    documentDataCache,
		resultChunkDataCache: resultChunkDataCache,
	}, nil
}

func (s *Server) Start() error {
	addr := net.JoinHostPort(s.host, strconv.FormatInt(int64(s.port), 10))
	handler := ot.Middleware(s.handler())
	server := &http.Server{Addr: addr, Handler: handler}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// NOTE: the stuff below is pretty rough and I'm not planning on putting too much
// effort into this while we're doing the port. This is an internal API so it's
// allowed to be a bit shoddy during this transitionary period. I'm not even sure
// if HTTP is the right transport for the long term.

func (s *Server) handler() http.Handler {
	mux := mux.NewRouter()
	mux.Path("/uploads/{id:[0-9]+}").Methods("GET").HandlerFunc(s.handleGetUpload)
	mux.Path("/uploads/{id:[0-9]+}").Methods("POST").HandlerFunc(s.handlePostUpload)
	mux.Path("/dbs/{id:[0-9]+}").Methods("POST").HandlerFunc(s.handlePostDatabase)
	mux.Path("/dbs/{id:[0-9]+}/exists").Methods("GET").HandlerFunc(s.handleExists)
	mux.Path("/dbs/{id:[0-9]+}/definitions").Methods("GET").HandlerFunc(s.handleDefinitions)
	mux.Path("/dbs/{id:[0-9]+}/references").Methods("GET").HandlerFunc(s.handleReferences)
	mux.Path("/dbs/{id:[0-9]+}/hover").Methods("GET").HandlerFunc(s.handleHover)
	mux.Path("/dbs/{id:[0-9]+}/monikersByPosition").Methods("GET").HandlerFunc(s.handleMonikersByPosition)
	mux.Path("/dbs/{id:[0-9]+}/monikerResults").Methods("GET").HandlerFunc(s.handleMonikerResults)
	mux.Path("/dbs/{id:[0-9]+}/packageInformation").Methods("GET").HandlerFunc(s.handlePackageInformation)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

// GET /uploads/{id:[0-9]+}
func (s *Server) handleGetUpload(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(paths.UploadFilename(s.bundleDir, idFromRequest(r)))
	if err != nil {
		http.Error(w, "Upload not found.", http.StatusNotFound)
		return
	}
	defer file.Close()

	copyAll(w, file)
}

// POST /uploads/{id:[0-9]+}
func (s *Server) handlePostUpload(w http.ResponseWriter, r *http.Request) {
	s.doUpload(w, r, paths.UploadFilename)
}

// POST /dbs/{id:[0-9]+}
func (s *Server) handlePostDatabase(w http.ResponseWriter, r *http.Request) {
	s.doUpload(w, r, paths.DBFilename)
}

// GET /dbs/{id:[0-9]+}/exists
func (s *Server) handleExists(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")

		return db.Exists(path)
	})
}

// GET /dbs/{id:[0-9]+}/definitions
func (s *Server) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")
		line, _ := strconv.Atoi(q.Get("line"))
		character, _ := strconv.Atoi(q.Get("character"))

		return db.Definitions(path, line, character)
	})
}

// GET /dbs/{id:[0-9]+}/references
func (s *Server) handleReferences(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")
		line, _ := strconv.Atoi(q.Get("line"))
		character, _ := strconv.Atoi(q.Get("character"))

		return db.References(path, line, character)
	})
}

// GET /dbs/{id:[0-9]+}/hover
func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")
		line, _ := strconv.Atoi(q.Get("line"))
		character, _ := strconv.Atoi(q.Get("character"))

		text, hoverRange, exists, err := db.Hover(path, line, character)
		if err != nil || !exists {
			return nil, err
		}

		return map[string]interface{}{"text": text, "range": hoverRange}, nil
	})
}

// GET /dbs/{id:[0-9]+}/monikersByPosition
func (s *Server) handleMonikersByPosition(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")
		line, _ := strconv.Atoi(q.Get("line"))
		character, _ := strconv.Atoi(q.Get("character"))

		return db.MonikersByPosition(path, line, character)
	})
}

// GET /dbs/{id:[0-9]+}/monikerResults
func (s *Server) handleMonikerResults(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		modelType := q.Get("modelType")
		scheme := q.Get("scheme")
		identifier := q.Get("identifier")
		skip, _ := strconv.Atoi(q.Get("skip"))
		take, err := strconv.Atoi(q.Get("take"))
		if err != nil {
			take = 100
		}

		var tableName string
		if modelType == "definition" {
			tableName = "definitions"
		} else if modelType == "reference" {
			tableName = "references"
		} else {
			return nil, fmt.Errorf("illegal tableName supplied")
		}

		locations, count, err := db.MonikerResults(tableName, scheme, identifier, skip, take)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"locations": locations, "count": count}, nil
	})
}

// GET /dbs/{id:[0-9]+}/packageInformation
func (s *Server) handlePackageInformation(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		q := r.URL.Query()
		path := q.Get("path")
		packageInformationID := types.ID(q.Get("packageInformationId"))

		packageInformationData, exists, err := db.PackageInformation(path, packageInformationID)
		if err != nil || !exists {
			return nil, err
		}

		return packageInformationData, nil
	})
}

// idFromRequest returns the database id from the request URL's path. This method
// must only be called from routes containing the `id:[0-9]+` pattern, as the error
// return from ParseInt is not checked.
func idFromRequest(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id
}

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "error", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "error", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}

// doUpload writes the HTTP request body to the path determined by the given
// makeFilename function.
func (s *Server) doUpload(w http.ResponseWriter, r *http.Request, makeFilename func(bundleDir string, id int64) string) {
	defer r.Body.Close()

	targetFile, err := os.OpenFile(makeFilename(s.bundleDir, idFromRequest(r)), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log15.Error("Failed to open target file", "error", err)
		http.Error(w, fmt.Sprintf("failed to open target file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(targetFile, r.Body); err != nil {
		log15.Error("Failed to write payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to write payload: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// dbQuery invokes the given handler with the database instance chosen from the
// route's id value and serializes the resulting value to the response writer.
func (s *Server) dbQuery(w http.ResponseWriter, r *http.Request, handler func(db *database.Database) (interface{}, error)) {
	filename := paths.DBFilename(s.bundleDir, idFromRequest(r))

	openDatabase := func() (*database.Database, error) {
		return database.OpenDatabase(filename, s.documentDataCache, s.resultChunkDataCache)
	}

	cacheHandler := func(db *database.Database) error {
		payload, err := handler(db)
		if err != nil {
			return err
		}

		writeJSON(w, payload)
		return nil
	}

	if err := s.databaseCache.WithDatabase(filename, openDatabase, cacheHandler); err != nil {
		http.Error(w, fmt.Sprintf("failed to handle query: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}
