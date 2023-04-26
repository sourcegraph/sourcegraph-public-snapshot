package upload

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestUploadIndex(t *testing.T) {
	var expectedPayload []byte
	for i := 0; i < 500; i++ {
		expectedPayload = append(expectedPayload, byte(i))
	}
	const (
		lsifContentType        = "application/x-ndjson+lsif"
		scipShardedContentType = "application/x-protobuf+scip-sharded"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading request body: %s", err)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != lsifContentType && contentType != scipShardedContentType {
			t.Fatalf("Content-Type header expected to be '%s' or '%s', got '%s'",
				lsifContentType, scipShardedContentType, r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Authorization") != "token hunter2" {
			t.Fatalf("Authorization header expected to be '%s', got '%s'", "token hunter2", r.Header.Get("Authorization"))
		}

		gzipReader, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("unexpected error creating gzip.Reader: %s", err)
		}
		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			t.Fatalf("unexpected error reading from gzip.Reader: %s", err)
		}

		if contentType == lsifContentType {
			if diff := cmp.Diff(expectedPayload, decompressed); diff != "" {
				t.Errorf("unexpected request payload (-want +got):\n%s", diff)
			}
		} else {
			require.Equal(t, contentType, scipShardedContentType)
			tarReader := tar.NewReader(bytes.NewReader(decompressed))
			numFiles := 0
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					break
				}
				if header.Typeflag != tar.TypeReg {
					continue
				}
				numFiles++
				require.NoError(t, err, "malformed tar file")
				buf := make([]byte, header.Size)
				_, err = io.ReadFull(tarReader, buf)
				require.NoError(t, err, "failed to read from tar file")
				if diff := cmp.Diff(expectedPayload, buf); diff != "" {
					t.Errorf("unexpected request payload (-want +got):\n%s", diff)
				}
			}
			require.NotEqual(t, numFiles, 0)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"42"}`))
	}))
	defer ts.Close()

	type testCase struct {
		numShards int
	}
	testCases := []testCase{{0}, {3}}

	makePayloadFile := func(dir string, pattern string) string {
		f, err := os.CreateTemp(dir, pattern)
		require.NoError(t, err, "failed to create temp file")
		_, err = io.Copy(f, bytes.NewReader(expectedPayload))
		require.NoError(t, err, "failed to write to temp file")
		_ = f.Close()
		return f.Name()
	}

	for _, testCase := range testCases {
		var indexPath string
		var contentType string
		if testCase.numShards == 0 {
			tmpFileName := makePayloadFile("", "")
			contentType = lsifContentType
			indexPath = tmpFileName
		} else {
			dir, err := os.MkdirTemp("", "")
			require.NoError(t, err, "failed to create temp dir")
			for i := 0; i < testCase.numShards; i++ {
				_ = makePayloadFile(dir, "*.shard.scip")
			}
			contentType = scipShardedContentType
			indexPath = dir
		}
		defer os.RemoveAll(indexPath)

		id, err := UploadIndex(context.Background(), indexPath, http.DefaultClient, UploadOptions{
			UploadRecordOptions: UploadRecordOptions{
				Repo:    "foo/bar",
				Commit:  "deadbeef",
				Root:    "proj/",
				Indexer: "lsif-go",
			},
			SourcegraphInstanceOptions: SourcegraphInstanceOptions{
				SourcegraphURL:      ts.URL,
				AccessToken:         "hunter2",
				GitHubToken:         "ght",
				MaxPayloadSizeBytes: 1000,
				AdditionalHeaders:   map[string]string{"Content-Type": contentType},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error uploading index: %s", err)
		}

		if id != 42 {
			t.Errorf("unexpected id. want=%d have=%d", 42, id)
		}
	}
}

func TestUploadIndexMultipart(t *testing.T) {
	var expectedPayload []byte
	for i := 0; i < 20000; i++ {
		expectedPayload = append(expectedPayload, byte(i))
	}

	var m sync.Mutex
	payloads := map[int][]byte{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("multiPart") != "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"42"}`)) // graphql id is TFNJRlVwbG9hZDoiNDIi
			return
		}

		if r.URL.Query().Get("index") != "" {
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("unexpected error reading request body: %s", err)
			}

			index, _ := strconv.Atoi(r.URL.Query().Get("index"))
			m.Lock()
			payloads[index] = payload
			m.Unlock()
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp file: %s", err)
	}
	defer func() { os.Remove(f.Name()) }()
	_, _ = io.Copy(f, bytes.NewReader(expectedPayload))
	_ = f.Close()

	id, err := UploadIndex(context.Background(), f.Name(), http.DefaultClient, UploadOptions{
		UploadRecordOptions: UploadRecordOptions{
			Repo:    "foo/bar",
			Commit:  "deadbeef",
			Root:    "proj/",
			Indexer: "lsif-go",
		},
		SourcegraphInstanceOptions: SourcegraphInstanceOptions{
			SourcegraphURL:      ts.URL,
			AccessToken:         "hunter2",
			GitHubToken:         "ght",
			MaxPayloadSizeBytes: 100,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error uploading index: %s", err)
	}

	if id != 42 {
		t.Errorf("unexpected id. want=%d have=%d", 42, id)
	}

	if len(payloads) != 5 {
		t.Errorf("unexpected payloads size. want=%d have=%d", 5, len(payloads))
	}

	var allPayloads []byte
	for i := 0; i < 5; i++ {
		allPayloads = append(allPayloads, payloads[i]...)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(allPayloads))
	if err != nil {
		t.Fatalf("unexpected error creating gzip.Reader: %s", err)
	}
	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("unexpected error reading from gzip.Reader: %s", err)
	}
	if diff := cmp.Diff(expectedPayload, decompressed); diff != "" {
		t.Errorf("unexpected gzipped contents (-want +got):\n%s", diff)
	}
}
