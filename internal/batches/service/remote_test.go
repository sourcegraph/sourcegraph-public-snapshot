package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockclient "github.com/sourcegraph/src-cli/internal/api/mock"
	"github.com/sourcegraph/src-cli/internal/batches/service"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestService_UpsertBatchChange(t *testing.T) {
	client := new(mockclient.Client)
	mockRequest := new(mockclient.Request)
	svc := service.New(&service.Opts{Client: client})

	tests := []struct {
		name string

		mockInvokes func()

		requestName        string
		requestNamespaceID string

		expectedID   string
		expectedName string
		expectedErr  error
	}{
		{
			name: "New Batch Change",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"name":      "my-change",
					"namespace": "my-namespace",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						json.Unmarshal([]byte(`{"upsertEmptyBatchChange":{"id":"123", "name":"my-change"}}`), &args[1])
					}).
					Return(true, nil).
					Once()
			},
			requestName:        "my-change",
			requestNamespaceID: "my-namespace",
			expectedID:         "123",
			expectedName:       "my-change",
		},
		{
			name: "Failed to upsert batch change",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"name":      "my-change",
					"namespace": "my-namespace",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Return(false, errors.New("did not get a good response code")).
					Once()
			},
			requestName:        "my-change",
			requestNamespaceID: "my-namespace",
			expectedErr:        errors.New("did not get a good response code"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			id, name, err := svc.UpsertBatchChange(context.Background(), test.requestName, test.requestNamespaceID)
			assert.Equal(t, test.expectedID, id)
			assert.Equal(t, test.expectedName, name)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
		})
	}
}

func TestService_CreateBatchSpecFromRaw(t *testing.T) {
	client := new(mockclient.Client)
	mockRequest := new(mockclient.Request)
	svc := service.New(&service.Opts{Client: client})

	tests := []struct {
		name string

		mockInvokes func()

		requestBatchSpec        string
		requestNamespaceID      string
		requestAllowIgnored     bool
		requestAllowUnsupported bool
		requestNoCache          bool
		requestBatchChange      string

		expectedID  string
		expectedErr error
	}{
		{
			name: "Create batch spec",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"batchSpec":        "abc",
					"namespace":        "some-namespace",
					"allowIgnored":     false,
					"allowUnsupported": false,
					"noCache":          false,
					"batchChange":      "123",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						json.Unmarshal([]byte(`{"createBatchSpecFromRaw":{"id":"xyz"}}`), &args[1])
					}).
					Return(true, nil).
					Once()
			},
			requestBatchSpec:        "abc",
			requestNamespaceID:      "some-namespace",
			requestAllowIgnored:     false,
			requestAllowUnsupported: false,
			requestNoCache:          false,
			requestBatchChange:      "123",
			expectedID:              "xyz",
		},
		{
			name: "Failed to create batch spec",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"batchSpec":        "abc",
					"namespace":        "some-namespace",
					"allowIgnored":     false,
					"allowUnsupported": false,
					"noCache":          false,
					"batchChange":      "123",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Return(false, errors.New("did not get a good response code")).
					Once()
			},
			requestBatchSpec:        "abc",
			requestNamespaceID:      "some-namespace",
			requestAllowIgnored:     false,
			requestAllowUnsupported: false,
			requestNoCache:          false,
			requestBatchChange:      "123",
			expectedErr:             errors.New("did not get a good response code"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			id, err := svc.CreateBatchSpecFromRaw(
				context.Background(),
				test.requestBatchSpec,
				test.requestNamespaceID,
				test.requestAllowIgnored,
				test.requestAllowUnsupported,
				test.requestNoCache,
				test.requestBatchChange,
			)
			assert.Equal(t, test.expectedID, id)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
		})
	}
}

func TestService_UploadBatchSpecWorkspaceFiles(t *testing.T) {
	tests := []struct {
		name  string
		steps []batches.Step

		setup       func(workingDir string) error
		mockInvokes func(client *mockclient.Client)

		expectedError error
	}{
		{
			name: "Upload single file",
			steps: []batches.Step{{
				Mount: []batches.Mount{{
					Path: "./hello.txt",
				}},
			}},
			setup: func(workingDir string) error {
				return writeTempFile(workingDir, "hello.txt", "hello world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Once()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				entry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello world!",
				}
				requestMatcher := multipartFormRequestMatcher(entry)
				client.On("Do", mock.MatchedBy(requestMatcher)).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Deduplicate files",
			steps: []batches.Step{{
				Mount: []batches.Mount{
					{
						Path: "./hello.txt",
					},
					{
						Path: "./hello.txt",
					},
					{
						Path: "./hello.txt",
					},
				},
			}},
			setup: func(workingDir string) error {
				return writeTempFile(workingDir, "hello.txt", "hello world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Once()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				entry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello world!",
				}
				requestMatcher := multipartFormRequestMatcher(entry)
				client.On("Do", mock.MatchedBy(requestMatcher)).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Upload multiple files",
			steps: []batches.Step{{
				Mount: []batches.Mount{
					{
						Path: "./hello.txt",
					},
					{
						Path: "./world.txt",
					},
				},
			}},
			setup: func(workingDir string) error {
				if err := writeTempFile(workingDir, "hello.txt", "hello"); err != nil {
					return err
				}
				return writeTempFile(workingDir, "world.txt", "world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Twice()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				helloEntry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello",
				}
				client.
					On("Do", mock.MatchedBy(multipartFormRequestMatcher(helloEntry))).
					Return(resp, nil).
					Once()

				worldEntry := &multipartFormEntry{
					fileName: "world.txt",
					content:  "world!",
				}
				client.
					On("Do", mock.MatchedBy(multipartFormRequestMatcher(worldEntry))).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Upload directory",
			steps: []batches.Step{{
				Mount: []batches.Mount{
					{
						Path: "./",
					},
				},
			}},
			setup: func(workingDir string) error {
				if err := writeTempFile(workingDir, "hello.txt", "hello"); err != nil {
					return err
				}
				return writeTempFile(workingDir, "world.txt", "world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Twice()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				helloEntry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello",
				}
				client.
					On("Do", mock.MatchedBy(multipartFormRequestMatcher(helloEntry))).
					Return(resp, nil).
					Once()

				worldEntry := &multipartFormEntry{
					fileName: "world.txt",
					content:  "world!",
				}
				client.
					On("Do", mock.MatchedBy(multipartFormRequestMatcher(worldEntry))).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Upload subdirectory",
			steps: []batches.Step{{
				Mount: []batches.Mount{
					{
						Path: "./scripts",
					},
				},
			}},
			setup: func(workingDir string) error {
				dir := filepath.Join(workingDir, "scripts")
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					return err
				}
				return writeTempFile(dir, "hello.txt", "hello world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Once()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				entry := &multipartFormEntry{
					path:     "scripts",
					fileName: "hello.txt",
					content:  "hello world!",
				}
				client.On("Do", mock.MatchedBy(multipartFormRequestMatcher(entry))).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Upload files and directory",
			steps: []batches.Step{{
				Mount: []batches.Mount{
					{
						Path: "./hello.txt",
					},
					{
						Path: "./world.txt",
					},
					{
						Path: "./scripts",
					},
				},
			}},
			setup: func(workingDir string) error {
				if err := writeTempFile(workingDir, "hello.txt", "hello"); err != nil {
					return err
				}
				if err := writeTempFile(workingDir, "world.txt", "world!"); err != nil {
					return err
				}
				dir := filepath.Join(workingDir, "scripts")
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					return err
				}
				return writeTempFile(dir, "something-else.txt", "this is neat")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Times(3)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				helloEntry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello",
				}
				client.On("Do", mock.MatchedBy(multipartFormRequestMatcher(helloEntry))).
					Return(resp, nil).
					Once()
				worldEntry := &multipartFormEntry{
					fileName: "world.txt",
					content:  "world!",
				}
				client.On("Do", mock.MatchedBy(multipartFormRequestMatcher(worldEntry))).
					Return(resp, nil).
					Once()
				somethingElseEntry := &multipartFormEntry{
					path:     "scripts",
					fileName: "something-else.txt",
					content:  "this is neat",
				}
				client.On("Do", mock.MatchedBy(multipartFormRequestMatcher(somethingElseEntry))).
					Return(resp, nil).
					Once()
			},
		},
		{
			name: "Bad status code",
			steps: []batches.Step{{
				Mount: []batches.Mount{{
					Path: "./hello.txt",
				}},
			}},
			setup: func(workingDir string) error {
				return writeTempFile(workingDir, "hello.txt", "hello world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Return(req, nil).
					Once()

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader([]byte("failed to upload file"))),
				}
				client.On("Do", mock.Anything).
					Return(resp, nil).
					Once()
			},
			expectedError: errors.New("failed to upload file"),
		},
		{
			name: "File exceeds limit",
			steps: []batches.Step{{
				Mount: []batches.Mount{{
					Path: "./hello.txt",
				}},
			}},
			setup: func(workingDir string) error {
				f, err := os.Create(filepath.Join(workingDir, "hello.txt"))
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err = io.Copy(f, io.LimitReader(neverEnding('a'), 11<<20)); err != nil {
					return err
				}
				return nil
			},
			mockInvokes: func(client *mockclient.Client) {
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Once()

				client.On("Do", mock.Anything).
					Return(nil, errors.New("file exceeds limit")).
					Once()
			},
			expectedError: errors.New("file exceeds limit"),
		},
		{
			name: "Long mount path",
			steps: []batches.Step{{
				Mount: []batches.Mount{{
					Path: "foo/../bar/../baz/../hello.txt",
				}},
			}},
			setup: func(workingDir string) error {
				dir := filepath.Join(workingDir, "foo")
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					return err
				}
				dir = filepath.Join(workingDir, "bar")
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					return err
				}
				dir = filepath.Join(workingDir, "baz")
				if err := os.Mkdir(dir, os.ModePerm); err != nil {
					return err
				}
				return writeTempFile(workingDir, "hello.txt", "hello world!")
			},
			mockInvokes: func(client *mockclient.Client) {
				// Body will get set with the body argument to NewHTTPRequest
				req := httptest.NewRequest(http.MethodPost, "http://fake.com/.api/files/batch-changes/123", nil)
				client.On("NewHTTPRequest", mock.Anything, http.MethodPost, ".api/files/batch-changes/123", mock.Anything).
					Run(func(args mock.Arguments) {
						req.Body = args[3].(*io.PipeReader)
					}).
					Return(req, nil).
					Once()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte{})),
				}
				entry := &multipartFormEntry{
					fileName: "hello.txt",
					content:  "hello world!",
				}
				requestMatcher := multipartFormRequestMatcher(entry)
				client.On("Do", mock.MatchedBy(requestMatcher)).
					Return(resp, nil).
					Once()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// TODO: use TempDir when https://github.com/golang/go/issues/51442 is cherry-picked into 1.18 or upgrade to 1.19+
			//tempDir := t.TempDir()
			workingDir, err := os.MkdirTemp("", test.name)
			require.NoError(t, err)
			t.Cleanup(func() {
				os.RemoveAll(workingDir)
			})

			if test.setup != nil {
				err := test.setup(workingDir)
				require.NoError(t, err)
			}

			client := new(mockclient.Client)
			svc := service.New(&service.Opts{Client: client})

			if test.mockInvokes != nil {
				test.mockInvokes(client)
			}

			err = svc.UploadBatchSpecWorkspaceFiles(context.Background(), workingDir, "123", test.steps)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
		})
	}
}

func writeTempFile(dir string, name string, content string) error {
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = io.WriteString(f, content); err != nil {
		return err
	}
	return nil
}

// 2006-01-02 15:04:05.999999999 -0700 MST
var modtimeRegex = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}\s[0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{1,9} \+0000 UTC$`)

func multipartFormRequestMatcher(entry *multipartFormEntry) func(*http.Request) bool {
	return func(req *http.Request) bool {
		// Prevent parsing the body for the wrong matcher - causes all kinds of havoc.
		if entry.calls > 0 {
			return false
		}
		// Clone the request. Running ParseMultipartForm changes the behavior of the request for any additional
		// matchers by consuming the request body.
		cloneReq, err := cloneRequest(req)
		if err != nil {
			fmt.Printf("failed to clone request: %s\n", err)
			return false
		}
		contentType := cloneReq.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			return false
		}
		if err := cloneReq.ParseMultipartForm(32 << 20); err != nil {
			fmt.Printf("failed to parse multipartform: %s\n", err)
			return false
		}
		if cloneReq.Form.Get("filepath") != entry.path {
			return false
		}
		if !modtimeRegex.MatchString(cloneReq.Form.Get("filemod")) {
			return false
		}
		f, header, err := cloneReq.FormFile("file")
		if err != nil {
			fmt.Printf("failed to get form file: %s\n", err)
			return false
		}
		if header.Filename != entry.fileName {
			return false
		}
		b, err := io.ReadAll(f)
		if err != nil {
			fmt.Printf("failed to read file: %s\n", err)
			return false
		}
		if string(b) != entry.content {
			return false
		}
		entry.calls++
		return true
	}
}

type multipartFormEntry struct {
	path     string
	fileName string
	content  string
	// This prevents some weird behavior that causes the request body to get read and throw errors.
	calls int
}

type neverEnding byte

func (b neverEnding) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(context.TODO())
	var b bytes.Buffer
	if _, err := b.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(&b)
	clone.Body = io.NopCloser(bytes.NewReader(b.Bytes()))
	return clone, nil
}
