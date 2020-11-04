package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/upload_store/mocks"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func TestGetUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	var fullContents []byte
	for i := 0; i < 1000; i++ {
		fullContents = append(fullContents, []byte(fmt.Sprintf("payload %d\n", i))...)
	}

	uploadStore := mocks.NewMockStore()
	uploadStore.GetFunc.SetDefaultReturn(ioutil.NopCloser(bytes.NewReader(compress(fullContents))), nil)

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL, uploadStore: uploadStore, ioCopy: io.Copy}
	r, err := client.GetUpload(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	}

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}

	if diff := cmp.Diff(fullContents, contents); diff != "" {
		t.Errorf("unexpected payload (-want +got):\n%s", diff)
	}
}

func TestGetUploadTransientErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	var fullContents []byte
	for i := 0; i < 1000; i++ {
		fullContents = append(fullContents, []byte(fmt.Sprintf("payload %d\n", i))...)
	}

	// mockCopy is like io.Copy but it will read 50 bytes and return an error
	// that appears to be a transient connection error.
	mockCopy := func(w io.Writer, r io.Reader) (int64, error) {
		var buf bytes.Buffer
		_, readErr := io.CopyN(&buf, r, 50)
		if readErr != nil && readErr != io.EOF {
			return 0, readErr
		}

		n, writeErr := io.Copy(w, bytes.NewReader(buf.Bytes()))
		if writeErr != nil {
			return 0, writeErr
		}

		if readErr == io.EOF {
			readErr = nil
		} else {
			readErr = errors.New("read: connection reset by peer")
		}
		return n, readErr
	}

	uploadStore := mocks.NewMockStore()
	uploadStore.GetFunc.SetDefaultHook(func(ctx context.Context, key string, seek int64) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(compress(fullContents)[seek:])), nil
	})

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL, uploadStore: uploadStore, ioCopy: mockCopy}
	r, err := client.GetUpload(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	}

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}

	if diff := cmp.Diff(fullContents, contents); diff != "" {
		t.Errorf("unexpected payload (-want +got):\n%s", diff)
	}
}

func TestGetUploadReadNothingLoop(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	var fullContents []byte
	for i := 0; i < 1000; i++ {
		fullContents = append(fullContents, []byte(fmt.Sprintf("payload %d\n", i))...)
	}

	uploadStore := mocks.NewMockStore()
	uploadStore.GetFunc.SetDefaultHook(func(ctx context.Context, key string, seek int64) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(compress(fullContents)[seek:])), nil
	})

	// Ensure that no progress transient errors do not cause an infinite loop
	mockCopy := func(w io.Writer, r io.Reader) (int64, error) {
		return 0, errors.New("read: connection reset by peer")
	}

	client := &bundleManagerClientImpl{bundleManagerURL: ts.URL, uploadStore: uploadStore, ioCopy: mockCopy}
	if _, err := client.GetUpload(context.Background(), 42); err != ErrNoDownloadProgress {
		t.Fatalf("unexpected error getting upload. want=%q have=%q", ErrNoDownloadProgress, err)
	}
}

func compress(payload []byte) []byte {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, _ = io.Copy(gzipWriter, bytes.NewReader(payload))
	_ = gzipWriter.Close()
	return buf.Bytes()
}
