package httpapi_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFileHandler_ServeHTTP(t *testing.T) {
	mockStore := new(mockBatchesStore)

	batchSpecRandID := "123"
	batchSpecWorkspaceFileRandID := "987"

	modifiedTimeString := "2022-08-15 19:30:25.410972423 +0000 UTC"
	modifiedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modifiedTimeString)
	require.NoError(t, err)

	operations := httpapi.NewOperations(&observation.TestContext)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	creatorID := bt.CreateTestUser(t, db, false).ID
	adminID := bt.CreateTestUser(t, db, true).ID

	tests := []struct {
		name string

		method      string
		path        string
		requestBody func() (io.Reader, string)

		mockInvokes func()

		userID int32

		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:               "Method not allowed",
			method:             http.MethodPatch,
			path:               fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "Get file",
			method: http.MethodGet,
			path:   fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecWorkspaceFile", mock.Anything, store.GetBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(&btypes.BatchSpecWorkspaceFile{Path: "foo/bar", FileName: "hello.txt", Content: []byte("Hello world!")}, nil).
					Once()
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "Hello world!",
		},
		{
			name:                 "Get file missing file id",
			method:               http.MethodGet,
			path:                 fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "path incorrectly structured\n",
		},
		{
			name:   "Failed to find file",
			method: http.MethodGet,
			path:   fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecWorkspaceFile", mock.Anything, store.GetBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(nil, errors.New("failed to find file")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "retrieving file: failed to find file\n",
		},
		{
			name:   "Upload file",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
				mockStore.
					On("UpsertBatchSpecWorkspaceFile", mock.Anything, &btypes.BatchSpecWorkspaceFile{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, Content: []byte("Hello world!"), ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
			},
			userID:             creatorID,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "Upload file as site admin",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
				mockStore.
					On("UpsertBatchSpecWorkspaceFile", mock.Anything, &btypes.BatchSpecWorkspaceFile{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, Content: []byte("Hello world!"), ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
			},
			userID:             adminID,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "Unauthorized upload",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: adminID}, nil).
					Once()
			},
			userID:             creatorID,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:   "Upload has invalid content type",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return nil, "application/json"
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
			},
			userID:               creatorID,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "parsing request: request Content-Type isn't multipart/form-data\n",
		},
		{
			name:   "Upload failed to lookup batch spec",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(nil, errors.New("failed to find batch spec")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "looking up batch spec: failed to find batch spec\n",
		},
		{
			name:   "Upload missing filemod",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.Close()
				return body, w.FormDataContentType()
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
			},
			userID:               creatorID,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "uploading file: missing file modification time\n",
		},
		{
			name:   "Upload missing file",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("filemod", modifiedTimeString)
				w.Close()
				return body, w.FormDataContentType()
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
			},
			userID:               creatorID,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "uploading file: http: no such file\n",
		},
		{
			name:   "Failed to create batch spec workspace file",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
				mockStore.
					On("UpsertBatchSpecWorkspaceFile", mock.Anything, &btypes.BatchSpecWorkspaceFile{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, Content: []byte("Hello world!"), ModifiedAt: modifiedTime}).
					Return(errors.New("failed to insert batch spec file")).
					Once()
			},
			userID:               creatorID,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "uploading file: failed to insert batch spec file\n",
		},
		{
			name:   "File Exists",
			method: http.MethodHead,
			path:   fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(1, nil).
					Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "File Does Not Exists",
			method: http.MethodHead,
			path:   fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(0, nil).
					Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "File Exists Error",
			method: http.MethodHead,
			path:   fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(0, errors.New("failed to count")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "checking file existence: failed to count\n",
		},
		{
			name:                 "Missing file id",
			method:               http.MethodHead,
			path:                 fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "path incorrectly structured\n",
		},
		{
			name:   "File exceeds max limit",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("filemod", modifiedTimeString)
				w.WriteField("filepath", "foo/bar")
				part, _ := w.CreateFormFile("file", "hello.txt")
				io.Copy(part, io.LimitReader(neverEnding('a'), 11<<20))
				w.Close()
				return body, w.FormDataContentType()
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID, UserID: creatorID}, nil).
					Once()
			},
			userID:               creatorID,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "request payload exceeds 10MB limit\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			handler := httpapi.NewFileHandler(db, mockStore, operations)

			var body io.Reader
			var contentType string
			if test.requestBody != nil {
				body, contentType = test.requestBody()
			}
			r := httptest.NewRequest(test.method, test.path, body)
			r.Header.Add("Content-Type", contentType)
			w := httptest.NewRecorder()

			// Setup user
			r = r.WithContext(actor.WithActor(r.Context(), actor.FromUser(test.userID)))

			// In order to get the mux variables from the path, setup mux routes
			router := mux.NewRouter()
			router.HandleFunc("/files/batches/{path:.*}", handler.ServeHTTP)
			router.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			responseBody, err := io.ReadAll(res.Body)
			// There should never be an error when reading the body
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponseBody, string(responseBody))

			// Ensure the mocked store functions get called correctly
			mockStore.AssertExpectations(t)
		})
	}
}

func multipartRequestBody(f file) (io.Reader, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	w.WriteField("filemod", f.modified)
	w.WriteField("filepath", f.path)
	part, _ := w.CreateFormFile("file", f.name)
	io.WriteString(part, f.content)
	w.Close()
	return body, w.FormDataContentType()
}

type file struct {
	name     string
	path     string
	content  string
	modified string
}

type mockBatchesStore struct {
	mock.Mock
}

func (m *mockBatchesStore) CountBatchSpecWorkspaceFiles(ctx context.Context, opts store.ListBatchSpecWorkspaceFileOpts) (int, error) {
	args := m.Called(ctx, opts)
	return args.Int(0), args.Error(1)
}

func (m *mockBatchesStore) GetBatchSpec(ctx context.Context, opts store.GetBatchSpecOpts) (*btypes.BatchSpec, error) {
	args := m.Called(ctx, opts)
	var obj *btypes.BatchSpec
	if args.Get(0) != nil {
		obj = args.Get(0).(*btypes.BatchSpec)
	}
	return obj, args.Error(1)
}

func (m *mockBatchesStore) GetBatchSpecWorkspaceFile(ctx context.Context, opts store.GetBatchSpecWorkspaceFileOpts) (*btypes.BatchSpecWorkspaceFile, error) {
	args := m.Called(ctx, opts)
	var obj *btypes.BatchSpecWorkspaceFile
	if args.Get(0) != nil {
		obj = args.Get(0).(*btypes.BatchSpecWorkspaceFile)
	}
	return obj, args.Error(1)
}

func (m *mockBatchesStore) UpsertBatchSpecWorkspaceFile(ctx context.Context, file *btypes.BatchSpecWorkspaceFile) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

type neverEnding byte

func (b neverEnding) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}
