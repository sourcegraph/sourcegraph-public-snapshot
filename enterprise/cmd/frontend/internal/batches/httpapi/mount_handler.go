package httpapi

import (
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
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MountHandler struct {
	logger      sglog.Logger
	store       *store.Store
	uploadStore uploadstore.Store
	operations  *Operations
}

func NewMountUploadHandler(
	store *store.Store,
	uploadStore uploadstore.Store,
	operations *Operations,
) http.Handler {
	handler := &MountHandler{
		logger:      sglog.Scoped("MountHandler", ""),
		store:       store,
		uploadStore: uploadStore,
		operations:  operations,
	}

	// 🚨 SECURITY: Non-internal installations of this handler will require a user/repo
	// visibility check with the remote code host (if enabled via site configuration).
	return authMiddleware(http.HandlerFunc(handler.handleUpload), store, operations.authMiddleware)
}

func (h *MountHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	batchSpecID := mux.Vars(r)["spec"]
	// 32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	count := r.Form.Get("count")
	if count == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	countNumber, err := strconv.Atoi(count)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	randID, err := unmarshalRandID(batchSpecID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	spec, err := h.store.GetBatchSpec(r.Context(), store.GetBatchSpecOpts{
		RandID: randID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := 0; i < countNumber; i++ {
		if err = h.uploadFile(r, spec, i); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
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

func NewMountRetrievalHandler(
	store *store.Store,
	uploadStore uploadstore.Store,
	operations *Operations,
) http.Handler {
	handler := &MountHandler{
		logger:      sglog.Scoped("MountHandler", ""),
		store:       store,
		uploadStore: uploadStore,
		operations:  operations,
	}

	return http.HandlerFunc(handler.handleRetrieval)
}

func (h *MountHandler) handleRetrieval(w http.ResponseWriter, r *http.Request) {
	batchSpecID := mux.Vars(r)["spec"]
	batchSpecRandID, err := unmarshalRandID(batchSpecID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	batchSpecMountID := mux.Vars(r)["mount"]
	batchSpecMountRandID, err := unmarshalRandID(batchSpecMountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mount, err := h.store.GetBatchSpecMount(r.Context(), store.GetBatchSpecMountOpts{
		RandID: batchSpecMountRandID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	key := filepath.Join(batchSpecRandID, mount.Path, mount.FileName)
	reader, err := h.uploadStore.Get(r.Context(), key)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	if _, err = io.Copy(w, reader); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
