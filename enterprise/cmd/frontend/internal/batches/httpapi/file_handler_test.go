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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFileHandler_ServeHTTP(t *testing.T) {
	mockStore := new(mockBatchesStore)

	batchSpecRandID := "123"
	batchSpecWorkspaceFileRandID := "987"

	modifiedTimeString := "2022-08-15 19:30:25.410972423 +0000 UTC"
	modifiedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", modifiedTimeString)
	require.NoError(t, err)

	tests := []struct {
		name       string
		isExecutor bool

		method      string
		path        string
		requestBody func() (io.Reader, string)

		mockInvokes func()

		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:               "Method not allowed",
			isExecutor:         true,
			method:             http.MethodPatch,
			path:               fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "Get file",
			isExecutor: true,
			method:     http.MethodGet,
			path:       fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
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
			isExecutor:           true,
			method:               http.MethodGet,
			path:                 fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "path incorrectly structured\n",
		},
		{
			name:       "Get file failed to find file",
			isExecutor: true,
			method:     http.MethodGet,
			path:       fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecWorkspaceFile", mock.Anything, store.GetBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(nil, errors.New("failed to find file")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to lookup file metadata: failed to find file\n",
		},
		{
			name:       "Upload file",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockStore.
					On("UpsertBatchSpecWorkspaceFile", mock.Anything, &btypes.BatchSpecWorkspaceFile{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, Content: []byte("Hello world!"), ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Upload file invalid content type",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return nil, "application/json"
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "failed to parse multipart form: request Content-Type isn't multipart/form-data\n",
		},
		{
			name:       "Upload file failed to lookup batch spec",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(nil, errors.New("failed to find batch spec")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to lookup batch spec: failed to find batch spec\n",
		},
		{
			name:       "Upload file missing filemod",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.Close()
				return body, w.FormDataContentType()
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to upload file: missing file modification time\n",
		},
		{
			name:       "Upload file missing file",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("filemod", modifiedTimeString)
				w.Close()
				return body, w.FormDataContentType()
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to upload file: http: no such file\n",
		},
		{
			name:       "Upload file failed to insert batch spec workspace file",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockStore.
					On("UpsertBatchSpecWorkspaceFile", mock.Anything, &btypes.BatchSpecWorkspaceFile{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, Content: []byte("Hello world!"), ModifiedAt: modifiedTime}).
					Return(errors.New("failed to insert batch spec file")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to upload file: failed to insert batch spec file\n",
		},
		{
			name:       "File Exists",
			isExecutor: true,
			method:     http.MethodHead,
			path:       fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(1, nil).
					Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "File Does Exists",
			isExecutor: true,
			method:     http.MethodHead,
			path:       fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(0, nil).
					Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "File Exists Error",
			isExecutor: true,
			method:     http.MethodHead,
			path:       fmt.Sprintf("/files/batches/%s/%s", batchSpecRandID, batchSpecWorkspaceFileRandID),
			mockInvokes: func() {
				mockStore.
					On("CountBatchSpecWorkspaceFiles", mock.Anything, store.ListBatchSpecWorkspaceFileOpts{RandID: batchSpecWorkspaceFileRandID}).
					Return(0, errors.New("failed to count")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to check if file exists: failed to count\n",
		},
		{
			name:                 "File Exists Missing file id",
			isExecutor:           true,
			method:               http.MethodHead,
			path:                 fmt.Sprintf("/files/batches/%s", batchSpecRandID),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "path incorrectly structured\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			handler := httpapi.NewFileHandler(mockStore, nil, test.isExecutor)

			var body io.Reader
			var contentType string
			if test.requestBody != nil {
				body, contentType = test.requestBody()
			}
			r := httptest.NewRequest(test.method, test.path, body)
			r.Header.Add("Content-Type", contentType)
			w := httptest.NewRecorder()

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
