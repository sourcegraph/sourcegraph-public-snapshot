package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MountHandler handles retrieving and uploading of mount files.
type MountHandler struct {
	logger     sglog.Logger
	store      BatchesStore
	operations *Operations
}

type BatchesStore interface {
	CountBatchSpecMounts(context.Context, store.ListBatchSpecMountsOpts) (int, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)
	GetBatchSpecMount(context.Context, store.GetBatchSpecMountOpts) (*btypes.BatchSpecMount, error)
	UpsertBatchSpecMount(context.Context, *btypes.BatchSpecMount) error
}

// NewMountHandler creates a new MountHandler.
func NewMountHandler(
	//db database.DB,
	store BatchesStore,
	operations *Operations,
	executor bool,
) http.Handler {
	handler := &MountHandler{
		logger:     sglog.Scoped("MountHandler", "").With(sglog.Bool("executor", executor)),
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

func (h *MountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *MountHandler) get(w http.ResponseWriter, r *http.Request) {
	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	batchSpecID := mux.Vars(r)["spec"]
	if batchSpecID == "" {
		http.Error(w, fmt.Sprintf("batch spec id not provided"), http.StatusBadRequest)
		return
	}
	batchSpecMountID := mux.Vars(r)["mount"]
	if batchSpecMountID == "" {
		http.Error(w, fmt.Sprintf("mount id not provided"), http.StatusBadRequest)
		return
	}

	mount, err := h.store.GetBatchSpecMount(r.Context(), store.GetBatchSpecMountOpts{
		RandID: batchSpecMountID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to lookup mount file metadata: %s", err), http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(mount.Content); err != nil {
		http.Error(w, fmt.Sprintf("failed to write file to reponse: %s", err), http.StatusInternalServerError)
		return
	}
}

func (h *MountHandler) exists(w http.ResponseWriter, r *http.Request) {
	// For now batchSpecID is only validation. When moving to the blob store, will need this to do queries.
	batchSpecID := mux.Vars(r)["spec"]
	if batchSpecID == "" {
		http.Error(w, fmt.Sprintf("batch spec id not provided"), http.StatusBadRequest)
		return
	}
	batchSpecMountID := mux.Vars(r)["mount"]
	if batchSpecMountID == "" {
		http.Error(w, fmt.Sprintf("mount id not provided"), http.StatusBadRequest)
		return
	}

	count, err := h.store.CountBatchSpecMounts(r.Context(), store.ListBatchSpecMountsOpts{RandID: batchSpecMountID})
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

func (h *MountHandler) upload(w http.ResponseWriter, r *http.Request) {
	batchSpecID := mux.Vars(r)["spec"]
	if batchSpecID == "" {
		http.Error(w, fmt.Sprintf("batch spec id not provided"), http.StatusBadRequest)
		return
	}

	// max memory: 32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %s", err), http.StatusBadRequest)
		return
	}

	count := r.Form.Get("count")
	if count == "" {
		http.Error(w, fmt.Sprintf("count was not provided"), http.StatusBadRequest)
		return
	}
	countNumber, err := strconv.Atoi(count)
	if err != nil {
		http.Error(w, fmt.Sprintf("count is not a number: %s", err), http.StatusBadRequest)
		return
	}

	spec, err := h.store.GetBatchSpec(r.Context(), store.GetBatchSpecOpts{
		RandID: batchSpecID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to lookup batch spec: %s", err), http.StatusInternalServerError)
		return
	}

	// TODO: could probably use some goroutines here
	for i := 0; i < countNumber; i++ {
		if err = h.uploadFile(r, spec, i); err != nil {
			http.Error(w, fmt.Sprintf("failed to upload file: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func (h *MountHandler) uploadFile(r *http.Request, spec *btypes.BatchSpec, index int) error {
	modtime := r.Form.Get(fmt.Sprintf("filemod_%d", index))
	if modtime == "" {
		return errors.New("missing file modification time")
	}
	modified, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modtime)
	if err != nil {
		return err
	}

	f, headers, err := r.FormFile(fmt.Sprintf("file_%d", index))
	if err != nil {
		return err
	}
	defer f.Close()

	filePath := r.Form.Get(fmt.Sprintf("filepath_%d", index))
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err = h.store.UpsertBatchSpecMount(r.Context(), &btypes.BatchSpecMount{
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
