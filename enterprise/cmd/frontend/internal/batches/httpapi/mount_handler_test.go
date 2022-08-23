package httpapi_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	uploadstoremocks "github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMountHandler_ServeHTTP(t *testing.T) {
	mockStore := new(mockBatchesStore)
	mockUploadStore := uploadstoremocks.NewMockStore()

	batchSpecRandID := "123"
	batchSpecMarshalledID := "ZG9lcy1ub3QtbWF0dGVyOiIxMjMi"
	batchSpecMountRandID := "987"
	batchSpecMountMarshalledID := "ZG9lcy1ub3QtbWF0dGVyOiI5ODci"

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
			path:               fmt.Sprintf("/batches/mount/%s/%s", batchSpecMarshalledID, batchSpecMountMarshalledID),
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "Get file",
			isExecutor: true,
			method:     http.MethodGet,
			path:       fmt.Sprintf("/batches/mount/%s/%s", batchSpecMarshalledID, batchSpecMountMarshalledID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecMount", mock.Anything, store.GetBatchSpecMountOpts{RandID: batchSpecMountRandID}).
					Return(&btypes.BatchSpecMount{Path: "foo/bar", FileName: "hello.txt"}, nil).
					Once()
				mockUploadStore.GetFunc.SetDefaultReturn(io.NopCloser(strings.NewReader("Hello world!")), nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "Hello world!",
		},
		{
			name:                 "Get file malformed spec id",
			isExecutor:           true,
			method:               http.MethodGet,
			path:                 fmt.Sprintf("/batches/mount/%s/%s", "foo", batchSpecMountMarshalledID),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "batch spec id is malformed: illegal base64 data at input byte 0\n",
		},
		{
			name:                 "Get file malformed mount id",
			isExecutor:           true,
			method:               http.MethodGet,
			path:                 fmt.Sprintf("/batches/mount/%s/%s", batchSpecMarshalledID, "foo"),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "mount id is malformed: illegal base64 data at input byte 0\n",
		},
		{
			name:       "Get file failed to find mount",
			isExecutor: true,
			method:     http.MethodGet,
			path:       fmt.Sprintf("/batches/mount/%s/%s", batchSpecMarshalledID, batchSpecMountMarshalledID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecMount", mock.Anything, store.GetBatchSpecMountOpts{RandID: batchSpecMountRandID}).
					Return(nil, errors.New("failed to find mount")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to lookup mount file metadata: failed to find mount\n",
		},
		{
			name:       "Get file failed to get from uploadstore",
			isExecutor: true,
			method:     http.MethodGet,
			path:       fmt.Sprintf("/batches/mount/%s/%s", batchSpecMarshalledID, batchSpecMountMarshalledID),
			mockInvokes: func() {
				mockStore.
					On("GetBatchSpecMount", mock.Anything, store.GetBatchSpecMountOpts{RandID: batchSpecMountRandID}).
					Return(&btypes.BatchSpecMount{Path: "foo/bar", FileName: "hello.txt"}, nil).
					Once()
				mockUploadStore.GetFunc.SetDefaultReturn(nil, errors.New("failed to find file"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to retrieve file: failed to find file\n",
		},
		{
			name:       "Upload file",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockUploadStore.UploadFunc.SetDefaultReturn(0, nil)
				mockStore.
					On("UpsertBatchSpecMount", mock.Anything, &btypes.BatchSpecMount{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Upload multiple files",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(
					file{name: "hello.txt", path: "foo/bar", content: "Hello!", modified: modifiedTimeString},
					file{name: "world.txt", path: "faz/baz", content: "World!", modified: modifiedTimeString},
				)
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockUploadStore.UploadFunc.SetDefaultReturn(0, nil)
				mockStore.
					On("UpsertBatchSpecMount", mock.Anything, &btypes.BatchSpecMount{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 6, ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
				mockStore.
					On("UpsertBatchSpecMount", mock.Anything, &btypes.BatchSpecMount{BatchSpecID: 1, FileName: "world.txt", Path: "faz/baz", Size: 6, ModifiedAt: modifiedTime}).
					Return(nil).
					Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Upload file malformed batch spec id",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", "foo"),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "batch spec id is malformed: illegal base64 data at input byte 0\n",
		},
		{
			name:       "Upload file invalid content type",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				return nil, "application/json"
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "failed to parse multipart form: request Content-Type isn't multipart/form-data\n",
		},
		{
			name:       "Upload file count not provided",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.Close()
				return body, w.FormDataContentType()
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "count was not provided\n",
		},
		{
			name:       "Upload file count is not a number",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("count", "a")
				w.Close()
				return body, w.FormDataContentType()
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "count is not a number: strconv.Atoi: parsing \"a\": invalid syntax\n",
		},
		{
			name:       "Upload file failed to lookup batch spec",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
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
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("count", "1")
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
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				w.WriteField("count", "1")
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
			name:       "Upload file failed to upload",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockUploadStore.UploadFunc.SetDefaultReturn(0, errors.New("failed to upload"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to upload file: failed to upload\n",
		},
		{
			name:       "Upload file failed to insert batch spec mount",
			isExecutor: true,
			method:     http.MethodPost,
			path:       fmt.Sprintf("/batches/mount/%s", batchSpecMarshalledID),
			requestBody: func() (io.Reader, string) {
				return multipartRequestBody(file{name: "hello.txt", path: "foo/bar", content: "Hello world!", modified: modifiedTimeString})
			},
			mockInvokes: func() {
				mockStore.On("GetBatchSpec", mock.Anything, store.GetBatchSpecOpts{RandID: batchSpecRandID}).
					Return(&btypes.BatchSpec{ID: 1, RandID: batchSpecRandID}, nil).
					Once()
				mockUploadStore.UploadFunc.SetDefaultReturn(0, nil)
				mockStore.
					On("UpsertBatchSpecMount", mock.Anything, &btypes.BatchSpecMount{BatchSpecID: 1, FileName: "hello.txt", Path: "foo/bar", Size: 12, ModifiedAt: modifiedTime}).
					Return(errors.New("failed to insert batch spec mount")).
					Once()
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "failed to upload file: failed to insert batch spec mount\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			handler := httpapi.NewMountHandler(mockStore, mockUploadStore, nil, test.isExecutor)

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
			// GET route
			router.HandleFunc("/batches/mount/{spec}/{mount}", handler.ServeHTTP)
			// POST route
			router.HandleFunc("/batches/mount/{spec}", handler.ServeHTTP)
			router.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			responseBody, err := io.ReadAll(res.Body)
			// There should never be an error when reading the body
			require.NoError(t, err)
			assert.Equal(t, test.expectedResponseBody, string(responseBody))

			// Ensure the mocked store functions get called correctly
			mockStore.AssertExpectations(t)
		})
	}
}

func multipartRequestBody(files ...file) (io.Reader, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	w.WriteField("count", strconv.Itoa(len(files)))

	for i, f := range files {
		w.WriteField(fmt.Sprintf("filemod_%d", i), f.modified)
		w.WriteField(fmt.Sprintf("filepath_%d", i), f.path)
		part, _ := w.CreateFormFile(fmt.Sprintf("file_%d", i), f.name)
		io.WriteString(part, f.content)
	}
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

func (m *mockBatchesStore) GetBatchSpec(ctx context.Context, opts store.GetBatchSpecOpts) (*btypes.BatchSpec, error) {
	args := m.Called(ctx, opts)
	var obj *btypes.BatchSpec
	if args.Get(0) != nil {
		obj = args.Get(0).(*btypes.BatchSpec)
	}
	return obj, args.Error(1)
}

func (m *mockBatchesStore) GetBatchSpecMount(ctx context.Context, opts store.GetBatchSpecMountOpts) (*btypes.BatchSpecMount, error) {
	args := m.Called(ctx, opts)
	var obj *btypes.BatchSpecMount
	if args.Get(0) != nil {
		obj = args.Get(0).(*btypes.BatchSpecMount)
	}
	return obj, args.Error(1)
}

func (m *mockBatchesStore) UpsertBatchSpecMount(ctx context.Context, mount *btypes.BatchSpecMount) error {
	args := m.Called(ctx, mount)
	return args.Error(0)
}
