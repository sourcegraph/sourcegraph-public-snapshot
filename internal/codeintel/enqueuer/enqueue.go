package enqueuer

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/inconshreveable/log15"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type ClientError struct {
	err error
}

func (e *ClientError) Error() string {
	return e.err.Error()
}

func clientError(message string, vals ...interface{}) error {
	return &ClientError{err: fmt.Errorf(message, vals...)}
}

type enqueuePayload struct {
	ID string `json:"id"`
}

type Enqueuer struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
}

func NewEnqueuer(db db.DB, bundleManagerClient bundles.BundleManagerClient) *Enqueuer {
	return &Enqueuer{
		db:                  db,
		bundleManagerClient: bundleManagerClient,
	}
}

// POST /upload
func (s *Enqueuer) HandleEnqueue(w http.ResponseWriter, r *http.Request) {
	payload, err := s.handleEnqueueErr(w, r)
	if err != nil {
		if cerr, ok := err.(*ClientError); ok {
			http.Error(w, cerr.Error(), http.StatusBadRequest)
			return
		}

		if err == ErrMetadataExceedsBuffer {
			http.Error(w, "Could not read indexer name from metaData vertex. Please supply it explicitly.", http.StatusBadRequest)
			return
		}

		log15.Error("Failed to enqueue payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to enqueue payload: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if payload != nil {
		w.WriteHeader(http.StatusAccepted)
		writeJSON(w, payload)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// UploadArgs are common arguments required to enqueue an upload for both single-payload
// and multipart uploads.
type UploadArgs struct {
	Commit       string
	Root         string
	RepositoryID int
	Indexer      string
}

// handleEnqueueErr dispatches to the correct handler function based on query args. Running the
// `src lsif upload` command will cause one of two sequences of requests to occur. For uploads that
// are small enough repos (that can be uploaded in one-shot), only one request will be made:
//
//    - POST `/upload?repositoryId,commit,root,indexerName`
//
// For larger uploads, the requests are broken up into a setup request, a serires of upload requests,
// and a finalization request:
//
//   - POST `/upload?repositoryId,commit,root,indexerName,multiPart=true,numParts={n}`
//   - POST `/upload?uploadId={id},index={i}`
//   - POST `/upload?uploadId={id},done=true`
//
// See the functions the following functions for details on how each request is handled:
//
//   - handleEnqueueSinglePayload
//   - handleEnqueueMultipartSetup
//   - handleEnqueueMultipartUpload
//   - handleEnqueueMultipartFinalize
func (s *Enqueuer) handleEnqueueErr(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	uploadArgs := UploadArgs{
		Commit:       getQuery(r, "commit"),
		Root:         sanitizeRoot(getQuery(r, "root")),
		RepositoryID: getQueryInt(r, "repositoryId"),
		Indexer:      getQuery(r, "indexerName"),
	}

	if !hasQuery(r, "multiPart") && !hasQuery(r, "uploadId") {
		return s.handleEnqueueSinglePayload(r, uploadArgs)
	}

	if hasQuery(r, "multiPart") {
		if numParts := getQueryInt(r, "numParts"); numParts <= 0 {
			return nil, clientError("illegal number of parts: %d", numParts)
		} else {
			return s.handleEnqueueMultipartSetup(r, uploadArgs, numParts)
		}
	}

	if !hasQuery(r, "uploadId") {
		return nil, clientError("no uploadId supplied")
	}

	upload, exists, err := s.db.GetUploadByID(r.Context(), getQueryInt(r, "uploadId"))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, clientError("upload not found")
	}

	if hasQuery(r, "index") {
		if partIndex := getQueryInt(r, "index"); partIndex < 0 || partIndex >= upload.NumParts {
			return nil, clientError("illegal part index: index %d is outside the range [0, %d)", partIndex, upload.NumParts)
		} else {
			return s.handleEnqueueMultipartUpload(r, upload, partIndex)
		}
	}

	if hasQuery(r, "done") {
		return s.handleEnqueueMultipartFinalize(r, upload)
	}

	return nil, clientError("no index supplied")
}

// handleEnqueueSinglePayload handles a non-multipart upload. This creates an upload record
// with state 'queued', proxies the data to the bundle manager, and returns the generated ID.
func (s *Enqueuer) handleEnqueueSinglePayload(r *http.Request, uploadArgs UploadArgs) (_ interface{}, err error) {
	// Newer versions of src-cli will do this same check before uploading the file. However,
	// older versions of src-cli will not guarantee that the index name query parameter is
	// sent. Requiring it now will break valid workflows. We only need ot maintain backwards
	// compatibility on single-payload uploads, as everything else is as new as the version
	// of src-cli that always sends the indexer name.
	if uploadArgs.Indexer == "" {
		// Tee all reads from the body into a buffer so that we don't destructively consume
		// any data from the body payload.
		var buf bytes.Buffer
		teeReader := io.TeeReader(r.Body, &buf)

		gzipReader, err := gzip.NewReader(teeReader)
		if err != nil {
			return nil, err
		}

		name, err := readIndexerName(gzipReader)
		if err != nil {
			return nil, err
		}
		uploadArgs.Indexer = name

		// Replace the body of the request with a reader that will produce all of the same
		// content: all of the data that was already read from r.Body, plus the remaining
		// content from r.Body.
		r.Body = ioutil.NopCloser(io.MultiReader(bytes.NewReader(buf.Bytes()), r.Body))
	}

	tx, err := s.db.Transact(r.Context())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	id, err := tx.InsertUpload(r.Context(), db.Upload{
		Commit:        uploadArgs.Commit,
		Root:          uploadArgs.Root,
		RepositoryID:  uploadArgs.RepositoryID,
		Indexer:       uploadArgs.Indexer,
		State:         "queued",
		NumParts:      1,
		UploadedParts: []int{0},
	})
	if err != nil {
		return nil, err
	}
	if err := s.bundleManagerClient.SendUpload(r.Context(), id, r.Body); err != nil {
		return nil, err
	}

	// older versions of src-cli expect a string
	return enqueuePayload{fmt.Sprintf("%d", id)}, nil
}

// handleEnqueueMultipartSetup handles the first request in a multipart upload. This creates a
// new upload record with state 'uploading' and returns the generated ID to be used in subsequent
// requests for the same upload.
func (s *Enqueuer) handleEnqueueMultipartSetup(r *http.Request, uploadArgs UploadArgs, numParts int) (interface{}, error) {
	id, err := s.db.InsertUpload(r.Context(), db.Upload{
		Commit:        uploadArgs.Commit,
		Root:          uploadArgs.Root,
		RepositoryID:  uploadArgs.RepositoryID,
		Indexer:       uploadArgs.Indexer,
		State:         "uploading",
		NumParts:      numParts,
		UploadedParts: nil,
	})
	if err != nil {
		return nil, err
	}

	// older versions of src-cli expect a string
	return enqueuePayload{fmt.Sprintf("%d", id)}, nil
}

// handleEnqueueMultipartUpload handles a partial upload in a multipart upload. This proxies the
// data to the bundle manager and marks the part index in the upload record.
func (s *Enqueuer) handleEnqueueMultipartUpload(r *http.Request, upload db.Upload, partIndex int) (_ interface{}, err error) {
	tx, err := s.db.Transact(r.Context())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := tx.AddUploadPart(r.Context(), upload.ID, partIndex); err != nil {
		return nil, err
	}
	if err := s.bundleManagerClient.SendUploadPart(r.Context(), upload.ID, partIndex, r.Body); err != nil {
		return nil, err
	}

	return nil, nil
}

//.handleEnqueueMultipartFinalize handles the final request of a multipart upload. This transitions the
// upload from 'uploading' to 'queued', then instructs the bundle manager to concatenate all of the part
// files together.
func (s *Enqueuer) handleEnqueueMultipartFinalize(r *http.Request, upload db.Upload) (_ interface{}, err error) {
	if len(upload.UploadedParts) != upload.NumParts {
		return nil, clientError("upload is missing %d parts", upload.NumParts-len(upload.UploadedParts))
	}

	tx, err := s.db.Transact(r.Context())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := tx.MarkQueued(r.Context(), upload.ID); err != nil {
		return nil, err
	}
	if err := s.bundleManagerClient.StitchParts(r.Context(), upload.ID); err != nil {
		return nil, err
	}

	return nil, nil
}
