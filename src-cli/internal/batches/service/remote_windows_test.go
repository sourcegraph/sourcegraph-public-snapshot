package service_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	mockclient "github.com/sourcegraph/src-cli/internal/api/mock"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestService_UploadBatchSpecWorkspaceFiles_Windows_Path(t *testing.T) {
	// TODO: use TempDir when https://github.com/golang/go/issues/51442 is cherry-picked into 1.18 or upgrade to 1.19+
	//tempDir := t.TempDir()
	workingDir, err := os.MkdirTemp("", "windows_path")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(workingDir)
	})

	// Set up test files and directories
	dir := filepath.Join(workingDir, "scripts")
	err = os.Mkdir(dir, os.ModePerm)
	require.NoError(t, err)

	dir = filepath.Join(dir, "another-dir")
	err = os.Mkdir(dir, os.ModePerm)
	require.NoError(t, err)

	err = writeTempFile(dir, "hello.txt", "hello world!")
	require.NoError(t, err)

	client := new(mockclient.Client)
	svc := service.New(&service.Opts{Client: client})

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
		path:     "scripts/another-dir",
		fileName: "hello.txt",
		content:  "hello world!",
	}
	requestMatcher := multipartFormRequestMatcher(entry)
	client.On("Do", mock.MatchedBy(requestMatcher)).
		Return(resp, nil).
		Once()

	steps := []batches.Step{{
		Mount: []batches.Mount{{
			Path: ".\\scripts\\another-dir\\hello.txt",
		}},
	}}
	err = svc.UploadBatchSpecWorkspaceFiles(context.Background(), workingDir, "123", steps)
	assert.NoError(t, err)

	client.AssertExpectations(t)
}
