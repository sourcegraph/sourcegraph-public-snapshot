package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

const DefaultUploadPageSize = 50
const DefaultReferencesPageSize = 100

func (s *Server) handler() http.Handler {
	mux := mux.NewRouter()
	mux.Path("/uploads/{id:[0-9]+}").Methods("GET").HandlerFunc(s.handleGetUploadByID)
	mux.Path("/uploads/{id:[0-9]+}").Methods("DELETE").HandlerFunc(s.handleDeleteUploadByID)
	mux.Path("/uploads/repository/{id:[0-9]+}").Methods("GET").HandlerFunc(s.handleGetUploadsByRepo)
	mux.Path("/upload").Methods("POST").HandlerFunc(s.handleEnqueue)
	mux.Path("/exists").Methods("GET").HandlerFunc(s.handleExists)
	mux.Path("/definitions").Methods("GET").HandlerFunc(s.handleDefinitions)
	mux.Path("/references").Methods("GET").HandlerFunc(s.handleReferences)
	mux.Path("/hover").Methods("GET").HandlerFunc(s.handleHover)
	mux.Path("/uploads").Methods("POST").HandlerFunc(s.handleUploads)
	mux.Path("/prune").Methods("POST").HandlerFunc(s.handlePrune)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

// GET /uploads/{id:[0-9]+}
func (s *Server) handleGetUploadByID(w http.ResponseWriter, r *http.Request) {
	upload, exists, err := s.db.GetUploadByID(r.Context(), int(idFromRequest(r)))
	if err != nil {
		log15.Error("Failed to retrieve upload", "error", err)
		http.Error(w, fmt.Sprintf("failed to retrieve upload: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	writeJSON(w, upload)
}

// DELETE /uploads/{id:[0-9]+}
func (s *Server) handleDeleteUploadByID(w http.ResponseWriter, r *http.Request) {
	exists, err := s.db.DeleteUploadByID(r.Context(), int(idFromRequest(r)), func(repositoryID int) (string, error) {
		tipCommit, err := gitserver.Head(s.db, repositoryID)
		if err != nil {
			return "", errors.Wrap(err, "gitserver.Head")
		}
		return tipCommit, nil
	})
	if err != nil {
		log15.Error("Failed to delete upload", "error", err)
		http.Error(w, fmt.Sprintf("failed to delete upload: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /uploads/repository/{id:[0-9]+}
func (s *Server) handleGetUploadsByRepo(w http.ResponseWriter, r *http.Request) {
	id := int(idFromRequest(r))
	limit := getQueryIntDefault(r, "limit", DefaultUploadPageSize)
	offset := getQueryInt(r, "offset")

	uploads, totalCount, err := s.db.GetUploadsByRepo(
		r.Context(),
		id,
		getQuery(r, "state"),
		getQuery(r, "query"),
		getQueryBool(r, "visibleAtTip"),
		limit,
		offset,
	)
	if err != nil {
		log15.Error("Failed to list uploads", "error", err)
		http.Error(w, fmt.Sprintf("failed to list uploads: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if offset+len(uploads) < totalCount {
		w.Header().Set("Link", makeNextLink(r.URL, map[string]interface{}{
			"limit":  limit,
			"offset": offset + len(uploads),
		}))
	}

	writeJSON(w, map[string]interface{}{"uploads": uploads, "totalCount": totalCount})
}

// POST /upload
func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	f, err := ioutil.TempFile("", "upload-")
	if err != nil {
		log15.Error("Failed to open target file", "error", err)
		http.Error(w, fmt.Sprintf("failed to open target file: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		log15.Error("Failed to write payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to write payload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	indexerName := getQuery(r, "indexerName")
	if indexerName == "" {
		if indexerName, err = readIndexerNameFromFile(f); err != nil {
			log15.Error("Failed to read indexer name from upload", "error", err)
			http.Error(w, fmt.Sprintf("failed to read indexer name from upload: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}

	tx, err := s.db.Transact(r.Context())
	if err != nil {
		log15.Error("Failed to start transaction", "error", err)
		http.Error(w, fmt.Sprintf("failed to start transaction: %s", err.Error()), http.StatusInternalServerError)
	}
	id, err := tx.Enqueue(
		r.Context(),
		getQuery(r, "commit"),
		sanitizeRoot(getQuery(r, "root")),
		getQueryInt(r, "repositoryId"),
		indexerName,
	)
	if err == nil {
		err = tx.Done(s.bundleManagerClient.SendUpload(r.Context(), id, f))
	}
	if err != nil {
		log15.Error("Failed to enqueue payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to enqueue payload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, map[string]interface{}{"id": fmt.Sprintf("%d", id)})
}

// GET /exists
func (s *Server) handleExists(w http.ResponseWriter, r *http.Request) {
	dumps, err := s.codeIntelAPI.FindClosestDumps(
		r.Context(),
		getQueryInt(r, "repositoryId"),
		getQuery(r, "commit"),
		getQuery(r, "path"),
	)
	if err != nil {
		log15.Error("Failed to handle exists request", "error", err)
		http.Error(w, fmt.Sprintf("failed to handle exists request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"uploads": dumps})
}

// GET /definitions
func (s *Server) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	defs, err := s.codeIntelAPI.Definitions(
		r.Context(),
		getQuery(r, "path"),
		getQueryInt(r, "line"),
		getQueryInt(r, "character"),
		getQueryInt(r, "uploadId"),
	)
	if err != nil {
		if err == api.ErrMissingDump {
			http.Error(w, "no such dump", http.StatusNotFound)
			return
		}

		log15.Error("Failed to handle definitions request", "error", err)
		http.Error(w, fmt.Sprintf("failed to handle definitions request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	outers, err := serializeLocations(defs)
	if err != nil {
		log15.Error("Failed to resolve locations", "error", err)
		http.Error(w, fmt.Sprintf("failed to resolve locations: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"locations": outers})
}

// GET /references
func (s *Server) handleReferences(w http.ResponseWriter, r *http.Request) {
	cursor, err := api.DecodeOrCreateCursor(
		getQuery(r, "path"),
		getQueryInt(r, "line"),
		getQueryInt(r, "character"),
		getQueryInt(r, "uploadId"),
		getQuery(r, "cursor"),
		s.db,
		s.bundleManagerClient,
	)
	if err != nil {
		if err == api.ErrMissingDump {
			http.Error(w, "no such dump", http.StatusNotFound)
			return
		}

		log15.Error("Failed to prepare cursor", "error", err)
		http.Error(w, fmt.Sprintf("failed to prepare cursor: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	limit := getQueryIntDefault(r, "limit", DefaultReferencesPageSize)
	if limit <= 0 {
		http.Error(w, "illegal limit", http.StatusBadRequest)
		return
	}

	locations, newCursor, hasNewCursor, err := s.codeIntelAPI.References(
		r.Context(),
		getQueryInt(r, "repositoryId"),
		getQuery(r, "commit"),
		limit,
		cursor,
	)
	if err != nil {
		log15.Error("Failed to handle references request", "error", err)
		http.Error(w, fmt.Sprintf("failed to handle references request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	outers, err := serializeLocations(locations)
	if err != nil {
		log15.Error("Failed to resolve locations", "error", err)
		http.Error(w, fmt.Sprintf("failed to resolve locations: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if hasNewCursor {
		w.Header().Set("Link", makeNextLink(r.URL, map[string]interface{}{
			"cursor": api.EncodeCursor(newCursor),
		}))
	}

	writeJSON(w, map[string]interface{}{"locations": outers})
}

// GET /hover
func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	text, rn, exists, err := s.codeIntelAPI.Hover(
		r.Context(),
		getQuery(r, "path"),
		getQueryInt(r, "line"),
		getQueryInt(r, "character"),
		getQueryInt(r, "uploadId"),
	)
	if err != nil {
		if err == api.ErrMissingDump {
			http.Error(w, "no such dump", http.StatusNotFound)
			return
		}

		log15.Error("Failed to handle hover request", "error", err)
		http.Error(w, fmt.Sprintf("failed to handle hover request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if !exists {
		writeJSON(w, nil)
	} else {
		writeJSON(w, map[string]interface{}{"text": text, "range": rn})
	}
}

// POST /uploads
func (s *Server) handleUploads(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		IDs []int `json:"ids"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log15.Error("Failed to read request body", "error", err)
		http.Error(w, fmt.Sprintf("failed to read request body: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	states, err := s.db.GetStates(r.Context(), payload.IDs)
	if err != nil {
		log15.Error("Failed to retrieve upload states", "error", err)
		http.Error(w, fmt.Sprintf("failed to retrieve upload states: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	pairs := []interface{}{}
	for k, v := range states {
		pairs = append(pairs, []interface{}{k, v})
	}

	writeJSON(w, map[string]interface{}{"type": "map", "value": pairs})
}

// POST /prune
func (s *Server) handlePrune(w http.ResponseWriter, r *http.Request) {
	id, prunable, err := s.db.DeleteOldestDump(r.Context())
	if err != nil {
		log15.Error("Failed to prune upload", "error", err)
		http.Error(w, fmt.Sprintf("failed to prune upload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if prunable {
		writeJSON(w, map[string]interface{}{"id": id})
	} else {
		writeJSON(w, map[string]interface{}{"id": nil})
	}
}
