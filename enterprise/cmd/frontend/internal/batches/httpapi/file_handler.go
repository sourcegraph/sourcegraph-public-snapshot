package httpapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FileHandler handles retrieving and uploading of files.
type FileHandler struct {
	logger     sglog.Logger
	db         database.DB
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
func NewFileHandler(db database.DB, store BatchesStore, operations *Operations) *FileHandler {
	return &FileHandler{
		logger:     sglog.Scoped("FileHandler", "Batch Changes mounted file REST API handler"),
		db:         db,
		store:      store,
		operations: operations,
	}
}

// Get retrieves the workspace file.
func (h *FileHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody, statusCode, err := h.get(r)

		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)

		if responseBody != nil {
			if _, err := io.Copy(w, responseBody); err != nil {
				h.logger.Error("failed to write payload to client", sglog.Error(err))
			}
		}
	})
}

func (h *FileHandler) get(r *http.Request) (_ io.Reader, statusCode int, err error) {
	ctx, _, endObservation := h.operations.get.With(r.Context(), &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("statusCode", statusCode),
		}})
	}()

	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPathParts(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	file, err := h.store.GetBatchSpecWorkspaceFile(ctx, store.GetBatchSpecWorkspaceFileOpts{RandID: fileID})
	if err != nil {
		if errors.Is(err, store.ErrNoResults) {
			return nil, http.StatusNotFound, errors.New("workspace file does not exist")
		}
		return nil, http.StatusInternalServerError, errors.Wrap(err, "retrieving file")
	}

	return bytes.NewReader(file.Content), http.StatusOK, nil
}

// Exists checks if the workspace file exists.
func (h *FileHandler) Exists() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, statusCode, err := h.exists(r)

		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)
	})
}

func (h *FileHandler) exists(r *http.Request) (_ io.Reader, statusCode int, err error) {
	ctx, _, endObservation := h.operations.exists.With(r.Context(), &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("statusCode", statusCode),
		}})
	}()

	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPathParts(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	count, err := h.store.CountBatchSpecWorkspaceFiles(ctx, store.ListBatchSpecWorkspaceFileOpts{RandID: fileID})
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "checking file existence")
	}

	// Either the count is 1 or zero.
	if count == 1 {
		return nil, http.StatusOK, nil
	} else {
		return nil, http.StatusNotFound, nil
	}
}

func getPathParts(r *http.Request) (string, string, error) {
	batchSpecRandID := mux.Vars(r)["spec"]
	if batchSpecRandID == "" {
		return "", "", errors.New("spec ID not provided")
	}

	batchSpecWorkspaceFileRandID := mux.Vars(r)["file"]
	if batchSpecWorkspaceFileRandID == "" {
		return "", "", errors.New("file ID not provided")
	}

	return batchSpecRandID, batchSpecWorkspaceFileRandID, nil
}

const maxUploadSize = 10 << 20 // 10MB

// Upload uploads a workspace file associated with a batch spec.
func (h *FileHandler) Upload() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		_, statusCode, err := h.upload(r)

		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)
	})
}

const maxMemory = 1 << 20 // 1MB

func (h *FileHandler) upload(r *http.Request) (_ io.Reader, statusCode int, err error) {
	ctx, _, endObservation := h.operations.upload.With(r.Context(), &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("statusCode", statusCode),
		}})
	}()

	specID := mux.Vars(r)["spec"]
	if specID == "" {
		return nil, http.StatusBadRequest, errors.New("spec ID not provided")
	}

	spec, err := h.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{RandID: specID})
	if err != nil {
		if errors.Is(err, store.ErrNoResults) {
			return nil, http.StatusNotFound, errors.New("batch spec does not exist")
		}
		return nil, http.StatusInternalServerError, errors.Wrap(err, "looking up batch spec")
	}

	// 🚨 SECURITY: Only site-admins or the creator of batch spec can upload files.
	if !isSiteAdminOrSameUser(ctx, h.logger, h.db, spec.UserID) {
		return nil, http.StatusUnauthorized, nil
	}

	// ParseMultipartForm parses the whole request body and stores the max size into memory. The rest of the body is
	// stored in temporary files on disk. The reason for parsing the whole request in one go is because data cannot be
	// "streamed" or "appended" to the bytea type column. Data for the bytea column must be inserted in one go.
	//
	// When we move to using a blob store (MinIO/S3/GCS), we can stream the parts instead. This means we won't need to
	// parse the entire request body up front. We will be able to iterate over and write the parts/chunks one at a time
	// - thus avoiding putting everything into memory.
	// See example: https://sourcegraph.com/github.com/rfielding/uploader@master/-/blob/uploader.go?L167
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		// TODO: starting in Go 1.19, if the request payload is too large the custom error MaxBytesError is returned here
		if strings.Contains(err.Error(), "request body too large") {
			return nil, http.StatusBadRequest, errors.New("request payload exceeds 10MB limit")
		} else {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "parsing request")
		}
	}

	if err = h.uploadBatchSpecWorkspaceFile(ctx, r, spec); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "uploading file")
	}

	return nil, http.StatusOK, err
}

func isSiteAdminOrSameUser(ctx context.Context, logger sglog.Logger, db database.DB, userId int32) bool {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		logger.Error("failed to get up current user", sglog.Error(err))
		return false
	}

	return user != nil && (user.SiteAdmin || user.ID == userId)
}

func (h *FileHandler) uploadBatchSpecWorkspaceFile(ctx context.Context, r *http.Request, spec *btypes.BatchSpec) error {
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
	if err = h.store.UpsertBatchSpecWorkspaceFile(ctx, &btypes.BatchSpecWorkspaceFile{
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

func defaultFileHandlerFunc(_ *http.Request) (io.Reader, int, error) {
	return nil, http.StatusMethodNotAllowed, nil
}
