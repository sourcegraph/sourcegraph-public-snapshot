package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	sgdb "github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/api"
	sgapi "github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/tomnomnom/linkheader"
)

const DefaultUploadPageSize = 50

// NOTE: the stuff below is pretty rough and I'm not planning on putting too much
// effort into this while we're doing the port. This is an internal API so it's
// allowed to be a bit shoddy during this transitionary period. I'm not even sure
// if HTTP is the right transport for the long term.

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
	upload, exists, err := s.db.GetUploadByID(context.Background(), int(idFromRequest(r)))
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
	getTipCommit := func(repositoryID int) (string, error) {
		repo, err := sgdb.Repos.Get(context.Background(), sgapi.RepoID(repositoryID))
		if err != nil {
			return "", err
		}

		cmd := gitserver.DefaultClient.Command("git", "rev-parse", "HEAD")
		cmd.Repo = gitserver.Repo{Name: repo.Name}
		out, err := cmd.CombinedOutput(context.Background())
		if err != nil {
			return "", err
		}
		return string(bytes.TrimSpace(out)), nil
	}

	exists, err := s.db.DeleteUploadByID(context.Background(), int(idFromRequest(r)), getTipCommit)
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
	q := r.URL.Query()
	term := q.Get("query")
	state := q.Get("state")
	visibleAtTip, _ := strconv.ParseBool(q.Get("visibleAtTip"))
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil {
		limit = DefaultUploadPageSize
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	uploads, totalCount, err := s.db.GetUploadsByRepo(context.Background(), id, state, term, visibleAtTip, limit, offset)
	if err != nil {
		log15.Error("Failed to list uploads", "error", err)
		http.Error(w, fmt.Sprintf("failed to list uploads: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if offset+len(uploads) < totalCount {
		url := r.URL
		q.Set("limit", strconv.FormatInt(int64(limit), 10))
		q.Set("offset", strconv.FormatInt(int64(offset+len(uploads)), 10))
		url.RawQuery = q.Encode()
		link := linkheader.Link{
			URL: url.String(),
			Rel: "next",
		}
		w.Header().Set("Link", link.String())
	}

	writeJSON(w, map[string]interface{}{"uploads": uploads, "totalCount": totalCount})
}

// POST /upload
func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repositoryID, _ := strconv.Atoi(q.Get("repositoryId"))
	commit := q.Get("commit")
	root := sanitizeRoot(q.Get("root"))
	indexerName := q.Get("indexerName")

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

	// TODO(efritz) - write tracing code
	tracingContext := "{}"

	if indexerName == "" {
		if indexerName, err = readIndexerNameFromFile(f); err != nil {
			log15.Error("Failed to read indexer name from upload", "error", err)
			http.Error(w, fmt.Sprintf("failed to read indexer name from upload: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}

	id, closer, err := s.db.Enqueue(context.Background(), commit, root, tracingContext, repositoryID, indexerName)
	if err == nil {
		err = closer.CloseTx(s.bundleManagerClient.SendUpload(context.Background(), id, f))
	}
	if err != nil {
		log15.Error("Failed to enqueue payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to enqueue payload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, map[string]interface{}{"id": id})
}

// GET /exists
func (s *Server) handleExists(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repositoryID, _ := strconv.Atoi(q.Get("repositoryId"))
	commit := q.Get("commit")
	file := q.Get("path")

	dumps, err := s.api.FindClosestDumps(repositoryID, commit, file)
	if err != nil {
		log15.Error("Failed to handle exists request", "error", err)
		http.Error(w, fmt.Sprintf("failed to handle exists request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"uploads": dumps})
}

// GET /definitions
func (s *Server) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	file := q.Get("path")
	line, _ := strconv.Atoi(q.Get("line"))
	character, _ := strconv.Atoi(q.Get("character"))
	uploadID, _ := strconv.Atoi(q.Get("uploadId"))

	defs, err := s.api.Definitions(file, line, character, uploadID)
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
	q := r.URL.Query()
	repositoryID, _ := strconv.Atoi(q.Get("repositoryId"))
	commit := q.Get("commit")
	limit, _ := strconv.Atoi(q.Get("limit"))

	cursor, err := api.DecodeCursorFromRequest(q, s.db, s.bundleManagerClient)
	if err != nil {
		if err == api.ErrMissingDump {
			http.Error(w, "no such dump", http.StatusNotFound)
			return
		}

		log15.Error("Failed to prepare cursor", "error", err)
		http.Error(w, fmt.Sprintf("failed to prepare cursor: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	locations, newCursor, hasNewCursor, err := s.api.References(repositoryID, commit, limit, cursor)
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
		url := r.URL
		q.Set("cursor", api.EncodeCursor(newCursor))
		url.RawQuery = q.Encode()
		link := linkheader.Link{
			URL: url.String(),
			Rel: "next",
		}
		w.Header().Set("Link", link.String())
	}

	writeJSON(w, map[string]interface{}{"locations": outers})
}

// GET /hover
func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	file := q.Get("path")
	line, _ := strconv.Atoi(q.Get("line"))
	character, _ := strconv.Atoi(q.Get("character"))
	uploadID, _ := strconv.Atoi(q.Get("uploadId"))

	text, rn, exists, err := s.api.Hover(file, line, character, uploadID)
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

	states, err := s.db.GetStates(context.Background(), payload.IDs)
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
	id, prunable, err := s.db.DeleteOldestDump(context.Background())
	if err != nil {
		log15.Error("Failed to prune upload", "error", err)
		http.Error(w, fmt.Sprintf("failed to prune upload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if !prunable {
		writeJSON(w, nil)
	} else {
		writeJSON(w, map[string]interface{}{"id": id})
	}
}
