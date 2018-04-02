package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

func TestClone(t *testing.T) {
	fileList := []batchOpenResult{
		batchOpenResult{
			path:    "/a.py",
			content: "This is file A.",
		},
		batchOpenResult{
			path:    "/b.py",
			content: "This is file B.",
		},
		batchOpenResult{
			path:    "/dir/c.py",
			content: "This is file C.",
		},
	}

	files := make(map[string]string)

	for _, aFile := range fileList {
		files[aFile.path] = aFile.content
	}

	baseDir, err := ioutil.TempDir("", uuid.New().String()+"testClone")
	if err != nil {
		t.Fatalf(errors.Wrap(err, "when creating temp directory for clone test").Error())
	}

	defer os.Remove(baseDir)

	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		err := fs.Clone(ctx, baseDir)

		if err != nil {
			t.Error(errors.Wrapf(err, "when calling clone(baseDir=%s)", baseDir))
		}

		discoveredFiles := make(map[string]string)

		err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return errors.Wrapf(err, "when walking walkFunc for path %s", path)
			}

			if info.IsDir() {
				return nil
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "when calling readFile for path %s", path)
			}

			if pathHasPrefix(path, baseDir) {
				path = filepath.Join("/", pathTrimPrefix(path, baseDir))
			}

			discoveredFiles[path] = string(content)

			return nil
		})

		if err != nil {
			t.Error(errors.Wrapf(err, "when calling Walk for baseDir %s", baseDir))
		}

		if !reflect.DeepEqual(files, discoveredFiles) {
			t.Errorf("for clone(baseDir=%s) expected %v, actual %v", baseDir, files, discoveredFiles)
		}
	})
}

func TestBatchOpen(t *testing.T) {
	fileList := []batchOpenResult{
		batchOpenResult{
			path:    "/a.py",
			content: "This is file A.",
		},
		batchOpenResult{
			path:    "/b.py",
			content: "This is file B.",
		},
		batchOpenResult{
			path:    "/dir/c.py",
			content: "This is file C.",
		},
	}

	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].path < fileList[j].path
	})

	files := make(map[string]string)

	for _, aFile := range fileList {
		files[aFile.path] = aFile.content
	}

	// open single file
	for _, aFile := range fileList {
		runTest(t, files, func(ctx context.Context, fs *remoteFS) {
			results, err := fs.BatchOpen(ctx, []string{aFile.path})

			if err != nil {
				t.Error(errors.Wrapf(err, "when calling batchOpen on path: %s", aFile.path))
			}

			if !reflect.DeepEqual(results, []batchOpenResult{aFile}) {
				t.Errorf("for batchOpen(paths=%v) expected %v, actual %v", []string{aFile.path}, []batchOpenResult{aFile}, results)
			}
		})
	}

	// open multiple files
	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		var allPaths []string

		for _, aFile := range fileList {
			allPaths = append(allPaths, aFile.path)
		}

		results, err := fs.BatchOpen(ctx, allPaths)

		if err != nil {
			t.Error(errors.Wrapf(err, "when calling batchOpen on paths: %v", allPaths))
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].path < results[j].path
		})

		if !reflect.DeepEqual(results, fileList) {
			t.Errorf("for batchOpen(paths=%v) expected %v, actual %v", allPaths, fileList, results)
		}
	})

	// open single invalid file
	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		_, err := fs.BatchOpen(ctx, []string{"/non/existent/file.py"})

		if err == nil {
			t.Error("expected error when trying to batchOpen non-existent file '/non/existent/file.py'")
		}
	})

	// open multiple valid files and one invalid file
	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		allPaths := []string{"non/existent/file.py"}

		for _, aFile := range fileList {
			allPaths = append(allPaths, aFile.path)
		}

		_, err := fs.BatchOpen(ctx, allPaths)

		if err == nil {
			t.Errorf("expected error when trying to batchOpen(paths=%v) which includes non-existent file '/non/existent/file.py'", allPaths)
		}
	})

	// open zero files
	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		results, err := fs.BatchOpen(ctx, []string{})

		if err != nil {
			t.Error(errors.Wrapf(err, "when calling batchOpen on zero paths"))
		}

		if len(results) > 0 {
			t.Error("expected zero results when trying to batchOpen zero paths")
		}
	})
}

func TestOpen(t *testing.T) {
	fileList := []batchOpenResult{
		batchOpenResult{
			path:    "/a.py",
			content: "This is file A.",
		},
		batchOpenResult{
			path:    "/b.py",
			content: "This is file B.",
		},
		batchOpenResult{
			path:    "/dir/c.py",
			content: "This is file C.",
		},
	}

	files := make(map[string]string)

	for _, aFile := range fileList {
		files[aFile.path] = aFile.content
	}

	for _, aFile := range fileList {
		runTest(t, files, func(ctx context.Context, fs *remoteFS) {
			actualFileContent, err := fs.Open(ctx, aFile.path)

			if err != nil {
				t.Error(errors.Wrapf(err, "when calling open on path: %s", aFile.path))
			}

			if actualFileContent != aFile.content {
				t.Errorf("for open(path=%s) expected %v, actual %v", aFile.path, aFile.content, actualFileContent)
			}
		})
	}

	runTest(t, files, func(ctx context.Context, fs *remoteFS) {
		_, err := fs.Open(ctx, "/c.py")
		if err == nil {
			t.Errorf("expected error when trying to open non-existent file '/c.py'")
		}
	})
}

func TestWalk(t *testing.T) {
	type testCase struct {
		fileNames         []string
		base              string
		expectedFileNames []string
	}

	tests := []testCase{
		testCase{
			fileNames:         []string{"/a.py", "/b.py", "/dir/c.py"},
			base:              "/",
			expectedFileNames: []string{"/a.py", "/b.py", "/dir/c.py"},
		},
		testCase{
			fileNames:         []string{"/a.py", "/b.py", "/dir/c.py"},
			base:              "/dir",
			expectedFileNames: []string{"/dir/c.py"},
		},
		testCase{
			fileNames:         []string{"/a.py", "/b.py", "/dir/c.py"},
			base:              "/di",
			expectedFileNames: []string{},
		},
		testCase{
			fileNames:         []string{"/a.py", "/b.py", "/dir/c.py"},
			base:              "/notadir",
			expectedFileNames: []string{},
		},
	}

	for _, test := range tests {
		files := make(map[string]string)

		for _, fileName := range test.fileNames {
			files[fileName] = ""
		}

		runTest(t, files, func(ctx context.Context, fs *remoteFS) {
			actualFileNames, err := fs.Walk(ctx, test.base)
			if err != nil {
				t.Error(errors.Wrapf(err, "when calling walk on base: %s", test.base))
			}

			sort.Strings(actualFileNames)
			sort.Strings(test.expectedFileNames)

			if len(actualFileNames) == 0 && len(test.expectedFileNames) == 0 {
				// special case empty slice versus nil comparsion below?
				return
			}

			if !reflect.DeepEqual(actualFileNames, test.expectedFileNames) {
				t.Errorf("for walk(base=%s) expected %v, actual %v", test.base, test.expectedFileNames, actualFileNames)
			}
		})
	}
}

func runTest(t *testing.T, files map[string]string, checkFunc func(ctx context.Context, fs *remoteFS)) {
	ctx := context.Background()

	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	clientConn := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(a, jsonrpc2.VSCodeObjectCodec{}), &testFS{
		t:     t,
		files: files,
	})

	serverConn := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(b, jsonrpc2.VSCodeObjectCodec{}), &noopHandler{})
	defer clientConn.Close()
	defer serverConn.Close()

	fs := &remoteFS{
		conn: serverConn,
	}

	checkFunc(ctx, fs)
}

type testFS struct {
	t     *testing.T
	files map[string]string // map of file names to content
}

func (client *testFS) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Notif {
		return
	}

	switch req.Method {
	case "textDocument/xcontent":
		var contentParams lspext.ContentParams
		if err := json.Unmarshal(*req.Params, &contentParams); err != nil {
			client.t.Fatalf(errors.Wrapf(err, "unable to unmarshal params %v for textdocument/xcontent", req.Params).Error())
		}

		filePathRawURI := string(contentParams.TextDocument.URI)
		filePathURI, err := url.Parse(filePathRawURI)
		if err != nil {
			client.t.Fatalf(errors.Wrapf(err, "unable to parse URI %vfor textdocument/xcontent", filePathRawURI).Error())
		}

		content, present := client.files[filePathURI.Path]

		if !present {
			err := &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: fmt.Sprintf("requested file path %s does not exist", filePathURI),
				Data:    nil,
			}
			if replyErr := conn.ReplyWithError(ctx, req.ID, err); replyErr != nil {
				client.t.Fatalf(errors.Wrapf(replyErr, "error when sending back error reply for document %s", filePathURI).Error())
			}
			return
		}

		document := lsp.TextDocumentItem{
			URI:  contentParams.TextDocument.URI,
			Text: content,
		}

		if replyErr := conn.Reply(ctx, req.ID, document); replyErr != nil {
			client.t.Fatalf(errors.Wrapf(replyErr, "error when sending back content reply for document %v", document).Error())
		}

	case "workspace/xfiles":
		var filesParams lspext.FilesParams
		if err := json.Unmarshal(*req.Params, &filesParams); err != nil {
			client.t.Fatalf(errors.Wrapf(err, "unable to unmarshal params %v for workspace/xfiles", req.Params).Error())
		}

		var results []lsp.TextDocumentIdentifier
		for filePath := range client.files {
			if pathHasPrefix(filePath, filesParams.Base) {
				results = append(results, lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI(filePath),
				})
			}
		}

		if replyErr := conn.Reply(ctx, req.ID, results); replyErr != nil {
			client.t.Fatalf(errors.Wrapf(replyErr, "error when sending back files reply for base %s", filesParams.Base).Error())
		}

	default:
		err := &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: fmt.Sprintf("method %s is invalid - only textdocument/xcontent and workspace/xfiles are supported", req.Method),
			Data:    nil,
		}

		if replyErr := conn.ReplyWithError(ctx, req.ID, err); replyErr != nil {
			client.t.Fatalf(errors.Wrapf(replyErr, "error when sending back error reply for invalid method %s", req.Method).Error())
		}
	}
}

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}
