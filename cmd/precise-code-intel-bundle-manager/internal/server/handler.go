package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

const DefaultMonikerResultPageSize = 100

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
		return db.Exists(getQuery(r, "path"))
	})
}

// GET /dbs/{id:[0-9]+}/definitions
func (s *Server) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		return db.Definitions(getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
	})
}

// GET /dbs/{id:[0-9]+}/references
func (s *Server) handleReferences(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		return db.References(getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
	})
}

// GET /dbs/{id:[0-9]+}/hover
func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		text, hoverRange, exists, err := db.Hover(getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
		if err != nil || !exists {
			return nil, err
		}

		return map[string]interface{}{"text": text, "range": hoverRange}, nil
	})
}

// GET /dbs/{id:[0-9]+}/monikersByPosition
func (s *Server) handleMonikersByPosition(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		return db.MonikersByPosition(getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
	})
}

// GET /dbs/{id:[0-9]+}/monikerResults
func (s *Server) handleMonikerResults(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		var tableName string
		switch getQuery(r, "modelType") {
		case "definition":
			tableName = "definitions"
		case "reference":
			tableName = "references"
		default:
			return nil, errors.New("illegal tableName supplied")
		}

		locations, count, err := db.MonikerResults(
			tableName,
			getQuery(r, "scheme"),
			getQuery(r, "identifier"),
			getQueryInt(r, "skip"),
			getQueryIntDefault(r, "take", DefaultMonikerResultPageSize),
		)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"locations": locations, "count": count}, nil
	})
}

// GET /dbs/{id:[0-9]+}/packageInformation
func (s *Server) handlePackageInformation(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(db *database.Database) (interface{}, error) {
		packageInformationData, exists, err := db.PackageInformation(getQuery(r, "path"), types.ID(getQuery(r, "packageInformationId")))
		if err != nil || !exists {
			return nil, err
		}

		return packageInformationData, nil
	})
}

// doUpload writes the HTTP request body to the path determined by the given
// makeFilename function.
func (s *Server) doUpload(w http.ResponseWriter, r *http.Request, makeFilename func(bundleDir string, id int64) string) {
	defer r.Body.Close()

	targetFile, err := os.OpenFile(makeFilename(s.bundleDir, idFromRequest(r)), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log15.Error("Failed to open target file", "err", err)
		http.Error(w, fmt.Sprintf("failed to open target file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(targetFile, r.Body); err != nil {
		log15.Error("Failed to write payload", "err", err)
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
