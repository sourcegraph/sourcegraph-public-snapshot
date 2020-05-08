package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/ext"
	pkgerrors "github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/reader"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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
	file, err := os.Open(paths.UploadFilename(s.BundleDir, idFromRequest(r)))
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
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		exists, err := db.Exists(ctx, getQuery(r, "path"))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.Exists")
		}
		return exists, nil
	})
}

// GET /dbs/{id:[0-9]+}/definitions
func (s *Server) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		definitions, err := db.Definitions(ctx, getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.Definitions")
		}
		return definitions, nil
	})
}

// GET /dbs/{id:[0-9]+}/references
func (s *Server) handleReferences(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		references, err := db.References(ctx, getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.References")
		}
		return references, nil
	})
}

// GET /dbs/{id:[0-9]+}/hover
func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		text, hoverRange, exists, err := db.Hover(ctx, getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.Hover")
		}
		if !exists {
			return nil, nil
		}

		return map[string]interface{}{"text": text, "range": hoverRange}, nil
	})
}

// GET /dbs/{id:[0-9]+}/monikersByPosition
func (s *Server) handleMonikersByPosition(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		monikerLocations, err := db.MonikersByPosition(ctx, getQuery(r, "path"), getQueryInt(r, "line"), getQueryInt(r, "character"))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.MonikersByPosition")
		}
		return monikerLocations, nil
	})
}

// GET /dbs/{id:[0-9]+}/monikerResults
func (s *Server) handleMonikerResults(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		var tableName string
		switch getQuery(r, "modelType") {
		case "definition":
			tableName = "definitions"
		case "reference":
			tableName = "references"
		default:
			return nil, errors.New("illegal tableName supplied")
		}

		skip := getQueryInt(r, "skip")
		if skip < 0 {
			return nil, errors.New("illegal skip supplied")
		}

		take := getQueryInt(r, "take")
		if take < 0 {
			return nil, errors.New("illegal take supplied")
		}

		locations, count, err := db.MonikerResults(
			ctx,
			tableName,
			getQuery(r, "scheme"),
			getQuery(r, "identifier"),
			skip,
			take,
		)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.MonikerResults")
		}

		return map[string]interface{}{"locations": locations, "count": count}, nil
	})
}

// GET /dbs/{id:[0-9]+}/packageInformation
func (s *Server) handlePackageInformation(w http.ResponseWriter, r *http.Request) {
	s.dbQuery(w, r, func(ctx context.Context, db database.Database) (interface{}, error) {
		packageInformationData, exists, err := db.PackageInformation(
			ctx,
			getQuery(r, "path"),
			types.ID(getQuery(r, "packageInformationId")),
		)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.PackageInformation")
		}
		if !exists {
			return nil, nil
		}

		return packageInformationData, nil
	})
}

// doUpload writes the HTTP request body to the path determined by the given
// makeFilename function.
func (s *Server) doUpload(w http.ResponseWriter, r *http.Request, makeFilename func(bundleDir string, id int64) string) {
	defer r.Body.Close()

	targetFile, err := os.OpenFile(makeFilename(s.BundleDir, idFromRequest(r)), os.O_WRONLY|os.O_CREATE, 0666)
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

type dbQueryHandlerFn func(ctx context.Context, db database.Database) (interface{}, error)

// dbQuery invokes the given handler with the database instance chosen from the
// route's id value and serializes the resulting value to the response writer. If an
// error occurs it will be written to the body of a 500-level response.
func (s *Server) dbQuery(w http.ResponseWriter, r *http.Request, handler dbQueryHandlerFn) {
	if err := s.dbQueryErr(w, r, handler); err != nil {
		http.Error(w, fmt.Sprintf("failed to handle query: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// queryBundleErr invokes the given handler with the database instance chosen from the
// route's id value and serializes the resulting value to the response writer. If an
// error occurs it will be returned.
func (s *Server) dbQueryErr(w http.ResponseWriter, r *http.Request, handler dbQueryHandlerFn) (err error) {
	ctx := r.Context()
	filename := paths.DBFilename(s.BundleDir, idFromRequest(r))
	cached := true

	span, ctx := ot.StartSpanFromContext(ctx, "dbQuery")
	span.SetTag("filename", filename)
	defer func() {
		span.SetTag("cached", cached)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	openDatabase := func() (database.Database, error) {
		cached = false

		// TODO - What is the behavior if the db is missing? Should we stat first or clean up after?
		sqliteReader, err := reader.NewSQLiteReader(filename, serializer.NewDefaultSerializer())
		if err != nil {
			return nil, pkgerrors.Wrap(err, "reader.NewSQLiteReader")
		}

		database, err := database.OpenDatabase(ctx, filename, s.wrapReader(sqliteReader), s.DocumentDataCache, s.ResultChunkDataCache)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "database.OpenDatabase")
		}

		return s.wrapDatabase(database), nil
	}

	cacheHandler := func(db database.Database) error {
		payload, err := handler(ctx, db)
		if err != nil {
			return err
		}

		writeJSON(w, payload)
		return nil
	}

	return s.DatabaseCache.WithDatabase(filename, openDatabase, cacheHandler)
}

func (s *Server) wrapReader(innerReader reader.Reader) reader.Reader {
	return reader.NewObserved(innerReader, s.ObservationContext, "precise_code_intel_bundle_manager")
}

func (s *Server) wrapDatabase(innerDatabase database.Database) database.Database {
	return database.NewObserved(innerDatabase, s.ObservationContext)
}
