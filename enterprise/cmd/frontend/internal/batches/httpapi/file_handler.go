package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FileHandler handles retrieving and uploading of files.
type FileHandler struct {
	logger     sglog.Logger
	store      BatchesStore
	operations *Operations
}

type BatchesStore interface {
	CountBatchSpecWorkspaceFiles(context.Context, store.ListBatchSpecWorkspaceFileOpts) (int, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)
	GetBatchSpecWorkspaceFile(context.Context, store.GetBatchSpecWorkspaceFileOpts) (*btypes.BatchSpecWorkspaceFile, error)
	UpsertBatchSpecWorkspaceFile(context.Context, *btypes.BatchSpecWorkspaceFile) error
}

// NewFileHandler creates a new FileHandler.
func NewFileHandler(
	//db database.DB,
	store BatchesStore,
	operations *Operations,
	executor bool,
) http.Handler {
	handler := &FileHandler{
		logger:     sglog.Scoped("FileHandler", "").With(sglog.Bool("executor", executor)),
		store:      store,
		operations: operations,
	}

	// If the handler is being used in the executor, no need to add security. Executor comes with its own security.
	if executor {
		return handler
	}

	// 🚨 SECURITY: TODO
	return handler
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodHead:
		h.exists(w, r)
	case http.MethodPost:
		h.upload(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *FileHandler) get(w http.ResponseWriter, r *http.Request) {
	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPathParts(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := h.store.GetBatchSpecWorkspaceFile(r.Context(), store.GetBatchSpecWorkspaceFileOpts{
		RandID: fileID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to lookup file metadata: %s", err), http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(file.Content); err != nil {
		http.Error(w, fmt.Sprintf("failed to write file to reponse: %s", err), http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) exists(w http.ResponseWriter, r *http.Request) {
	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPathParts(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := h.store.CountBatchSpecWorkspaceFiles(r.Context(), store.ListBatchSpecWorkspaceFileOpts{RandID: fileID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check if file exists: %s", err), http.StatusInternalServerError)
		return
	}

	// Either the count is 1 or zero.
	if count == 1 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func getPathParts(r *http.Request) (string, string, error) {
	path := mux.Vars(r)["path"]
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", errors.New("path incorrectly structured")
	}

	batchSpecRandID := parts[0]
	if batchSpecRandID == "" {
		return "", "", errors.New("spec ID not provided")
	}

	batchSpecWorkspaceFileID := parts[1]
	if batchSpecWorkspaceFileID == "" {
		return "", "", errors.New("file ID not provided")
	}

	return batchSpecRandID, batchSpecWorkspaceFileID, nil
}

const maxUploadSize = 10 << 20 // 10MB
const maxMemory = 1 << 20      // 1MB

func (h *FileHandler) upload(w http.ResponseWriter, r *http.Request) {
	specID := mux.Vars(r)["path"]
	if specID == "" {
		http.Error(w, "spec ID not provided", http.StatusBadRequest)
		return
	}

	// ParseMultipartForm parses the whole request body and store the max size into memory. The rest of the body is
	// stored in temporary files on disk. We need to do this since we are using Postgres and the column is bytea.
	//
	// When this is moved to use the blob store (MinIO/S3/GCS), we can stream the parts instead.
	// See example: https://sourcegraph.com/github.com/rfielding/uploader@master/-/blob/uploader.go?L167

	// Prevent client from uploading files that are too large. This will also be enforced on the src-cli side as well.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			http.Error(w, "request payload exceeds 10MB limit", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("failed to parse multipart form: %s", err), http.StatusInternalServerError)
		}
		return
	}

	spec, err := h.store.GetBatchSpec(r.Context(), store.GetBatchSpecOpts{RandID: specID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to lookup batch spec: %s", err), http.StatusInternalServerError)
		return
	}

	if err = h.uploadFile(r, spec); err != nil {
		http.Error(w, fmt.Sprintf("failed to upload file: %s", err), http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) uploadFile(r *http.Request, spec *btypes.BatchSpec) error {
	modtime := r.Form.Get("filemod")
	if modtime == "" {
		return errors.New("missing file modification time")
	}
	modified, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modtime)
	if err != nil {
		return err
	}

	f, headers, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer f.Close()

	filePath := r.Form.Get("filepath")
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err = h.store.UpsertBatchSpecWorkspaceFile(r.Context(), &btypes.BatchSpecWorkspaceFile{
		BatchSpecID: spec.ID,
		FileName:    headers.Filename,
		Path:        filePath,
		Size:        headers.Size,
		Content:     content,
		ModifiedAt:  modified,
	}); err != nil {
		return err
	}
	return nil
}
