package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MountHandler handles retrieving and uploading of mount files.
type MountHandler struct {
	logger      sglog.Logger
	store       BatchesStore
	uploadStore uploadstore.Store
	operations  *Operations
}

type BatchesStore interface {
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)
	GetBatchSpecMount(context.Context, store.GetBatchSpecMountOpts) (*btypes.BatchSpecMount, error)
	UpsertBatchSpecMount(context.Context, *btypes.BatchSpecMount) error
}

// NewMountHandler creates a new MountHandler.
func NewMountHandler(
	db database.DB,
	store BatchesStore,
	uploadStore uploadstore.Store,
	operations *Operations,
	executor bool,
) http.Handler {
	handler := &MountHandler{
		logger:      sglog.Scoped("MountHandler", "").With(sglog.Bool("executor", executor)),
		store:       store,
		uploadStore: uploadStore,
		operations:  operations,
	}

	// If the handler is being used in the executor, no need to add security. Executor comes with its own security.
	if executor {
		return handler
	}

	// 🚨 SECURITY: Non-internal installations of this handler will require a user/repo
	// visibility check with the remote code host (if enabled via site configuration).
	return authMiddleware(handler, db, operations.authMiddleware)
}

func (h *MountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPost:
		h.upload(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *MountHandler) get(w http.ResponseWriter, r *http.Request) {
	batchSpecID := mux.Vars(r)["spec"]
	batchSpecRandID, err := unmarshalRandID(batchSpecID)
	if err != nil {
		http.Error(w, fmt.Sprintf("batch spec id is malformed: %s", err), http.StatusBadRequest)
		return
	}
	batchSpecMountID := mux.Vars(r)["mount"]
	batchSpecMountRandID, err := unmarshalRandID(batchSpecMountID)
	if err != nil {
		http.Error(w, fmt.Sprintf("mount id is malformed: %s", err), http.StatusBadRequest)
		return
	}

	mount, err := h.store.GetBatchSpecMount(r.Context(), store.GetBatchSpecMountOpts{
		RandID: batchSpecMountRandID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to lookup mount file metadata: %s", err), http.StatusInternalServerError)
		return
	}

	key := filepath.Join(batchSpecRandID, mount.Path, mount.FileName)
	reader, err := h.uploadStore.Get(r.Context(), key)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve file: %s", err), http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	if _, err = io.Copy(w, reader); err != nil {
		http.Error(w, fmt.Sprintf("failed to write file to reponse: %s", err), http.StatusInternalServerError)
		return
	}
}

func (h *MountHandler) upload(w http.ResponseWriter, r *http.Request) {
	batchSpecID := mux.Vars(r)["spec"]
	randID, err := unmarshalRandID(batchSpecID)
	if err != nil {
		http.Error(w, fmt.Sprintf("batch spec id is malformed: %s", err), http.StatusBadRequest)
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
		RandID: randID,
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
	f, headers, err := r.FormFile(fmt.Sprintf("%s_%d", "file", index))
	if err != nil {
		return err
	}
	defer f.Close()

	filePath := r.Form.Get(fmt.Sprintf("filepath_%d", index))
	key := filepath.Join(spec.RandID, filePath, headers.Filename)
	if _, err = h.uploadStore.Upload(r.Context(), key, f); err != nil {
		return err
	}
	if err = h.store.UpsertBatchSpecMount(r.Context(), &btypes.BatchSpecMount{
		BatchSpecID: spec.ID,
		FileName:    headers.Filename,
		Path:        filePath,
		Size:        headers.Size,
	}); err != nil {
		return err
	}
	return nil
}

func unmarshalRandID(id string) (batchSpecRandID string, err error) {
	err = unmarshalID(id, &batchSpecRandID)
	return
}

func unmarshalID(id string, v interface{}) error {
	s, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return err
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return errors.New("invalid randID")
	}
	return json.Unmarshal(s[i+1:], v)
}
