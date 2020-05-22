package enqueuer

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
)

func TestHandleEnqueueSinglePayload(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.InsertUploadFunc.SetDefaultReturn(42, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"commit":       []string{"deadbeef"},
		"root":         []string{"proj/"},
		"repositoryId": []string{"50"},
		"indexerName":  []string{"lsif-go"},
	}).Encode()

	var expectedContents []byte
	for i := 0; i < 20000; i++ {
		expectedContents = append(expectedContents, byte(i))
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), bytes.NewReader(expectedContents))
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusAccepted {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusAccepted, w.Code)
	}
	if diff := cmp.Diff([]byte(`{"id":"42"}`), w.Body.Bytes()); diff != "" {
		t.Errorf("unexpected response payload (-want +got):\n%s", diff)
	}

	if len(mockDB.InsertUploadFunc.History()) != 1 {
		t.Errorf("unexpected number of InsertUploadFunc calls. want=%d have=%d", 1, len(mockDB.InsertUploadFunc.History()))
	} else {
		call := mockDB.InsertUploadFunc.History()[0]
		if call.Arg1.Commit != "deadbeef" {
			t.Errorf("unexpected commit. want=%q have=%q", "deadbeef", call.Arg1.Commit)
		}
		if call.Arg1.Root != "proj/" {
			t.Errorf("unexpected root. want=%q have=%q", "proj/", call.Arg1.Root)
		}
		if call.Arg1.RepositoryID != 50 {
			t.Errorf("unexpected repository id. want=%d have=%d", 50, call.Arg1.RepositoryID)
		}
		if call.Arg1.Indexer != "lsif-go" {
			t.Errorf("unexpected indexer name. want=%q have=%q", "lsif-go", call.Arg1.Indexer)
		}
	}

	if len(mockBundleManagerClient.SendUploadFunc.History()) != 1 {
		t.Errorf("unexpected number of SendUploadFunc calls. want=%d have=%d", 1, len(mockBundleManagerClient.SendUploadFunc.History()))
	} else {
		call := mockBundleManagerClient.SendUploadFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected bundle id. want=%d have=%d", 42, call.Arg1)
		}

		contents, err := ioutil.ReadAll(call.Arg2)
		if err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		}

		if diff := cmp.Diff(expectedContents, contents); diff != "" {
			t.Errorf("unexpected file contents (-want +got):\n%s", diff)
		}
	}
}

func TestHandleEnqueueSinglePayloadNoIndexerName(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.InsertUploadFunc.SetDefaultReturn(42, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"commit":       []string{"deadbeef"},
		"root":         []string{"proj/"},
		"repositoryId": []string{"50"},
	}).Encode()

	var lines []string
	lines = append(lines, `{"label": "metaData", "toolInfo": {"name": "lsif-go"}}`)
	for i := 0; i < 20000; i++ {
		lines = append(lines, `{"id": "a", "type": "edge", "label": "textDocument/references", "outV": "b", "inV": "c"}`)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, _ = io.Copy(gzipWriter, bytes.NewReader([]byte(strings.Join(lines, "\n"))))
	gzipWriter.Close()
	expectedContents := buf.Bytes()

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), bytes.NewReader(expectedContents))
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusAccepted {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusAccepted, w.Code)
	}

	if len(mockBundleManagerClient.SendUploadFunc.History()) != 1 {
		t.Errorf("unexpected number of SendUploadFunc calls. want=%d have=%d", 1, len(mockBundleManagerClient.SendUploadFunc.History()))
	} else {
		call := mockBundleManagerClient.SendUploadFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected bundle id. want=%d have=%d", 42, call.Arg1)
		}

		contents, err := ioutil.ReadAll(call.Arg2)
		if err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		}

		if diff := cmp.Diff(expectedContents, contents); diff != "" {
			t.Errorf("unexpected file contents (-want +got):\n%s", diff)
		}
	}
}

func TestHandleEnqueueMultipartSetup(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.InsertUploadFunc.SetDefaultReturn(42, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"commit":       []string{"deadbeef"},
		"root":         []string{"proj/"},
		"repositoryId": []string{"50"},
		"indexerName":  []string{"lsif-go"},
		"multiPart":    []string{"true"},
		"numParts":     []string{"3"},
	}).Encode()

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), nil)
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusAccepted {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusAccepted, w.Code)
	}
	if diff := cmp.Diff([]byte(`{"id":"42"}`), w.Body.Bytes()); diff != "" {
		t.Errorf("unexpected response payload (-want +got):\n%s", diff)
	}

	if len(mockDB.InsertUploadFunc.History()) != 1 {
		t.Errorf("unexpected number of InsertUploadFunc calls. want=%d have=%d", 1, len(mockDB.InsertUploadFunc.History()))
	} else {
		call := mockDB.InsertUploadFunc.History()[0]
		if call.Arg1.Commit != "deadbeef" {
			t.Errorf("unexpected commit. want=%q have=%q", "deadbeef", call.Arg1.Commit)
		}
		if call.Arg1.Root != "proj/" {
			t.Errorf("unexpected root. want=%q have=%q", "proj/", call.Arg1.Root)
		}
		if call.Arg1.RepositoryID != 50 {
			t.Errorf("unexpected repository id. want=%d have=%d", 50, call.Arg1.RepositoryID)
		}
		if call.Arg1.Indexer != "lsif-go" {
			t.Errorf("unexpected indexer name. want=%q have=%q", "lsif-go", call.Arg1.Indexer)
		}
	}
}

func TestHandleEnqueueMultipartUpload(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	upload := db.Upload{
		ID:            42,
		NumParts:      5,
		UploadedParts: []int{0, 1, 2, 3, 4},
	}

	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.GetUploadByIDFunc.SetDefaultReturn(upload, true, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"uploadId": []string{"42"},
		"index":    []string{"3"},
	}).Encode()

	var expectedContents []byte
	for i := 0; i < 20000; i++ {
		expectedContents = append(expectedContents, byte(i))
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), bytes.NewReader(expectedContents))
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusNoContent, w.Code)
	}

	if len(mockDB.AddUploadPartFunc.History()) != 1 {
		t.Errorf("unexpected number of AddUploadPartFunc calls. want=%d have=%d", 1, len(mockDB.AddUploadPartFunc.History()))
	} else {
		call := mockDB.AddUploadPartFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected commit. want=%q have=%q", 42, call.Arg1)
		}
		if call.Arg2 != 3 {
			t.Errorf("unexpected root. want=%q have=%q", 3, call.Arg2)
		}
	}

	if len(mockBundleManagerClient.SendUploadPartFunc.History()) != 1 {
		t.Errorf("unexpected number of SendUploadPartFunc calls. want=%d have=%d", 1, len(mockBundleManagerClient.SendUploadPartFunc.History()))
	} else {
		call := mockBundleManagerClient.SendUploadPartFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected bundle id. want=%d have=%d", 42, call.Arg1)
		}
		if call.Arg2 != 3 {
			t.Errorf("unexpected part index. want=%d have=%d", 3, call.Arg1)
		}

		contents, err := ioutil.ReadAll(call.Arg3)
		if err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		}

		if diff := cmp.Diff(expectedContents, contents); diff != "" {
			t.Errorf("unexpected file contents (-want +got):\n%s", diff)
		}
	}
}

func TestHandleEnqueueMultipartFinalize(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	upload := db.Upload{
		ID:            42,
		NumParts:      5,
		UploadedParts: []int{0, 1, 2, 3, 4},
	}
	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.GetUploadByIDFunc.SetDefaultReturn(upload, true, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"uploadId": []string{"42"},
		"done":     []string{"true"},
	}).Encode()

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), nil)
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusNoContent, w.Code)
	}

	if len(mockDB.MarkQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of MarkQueuedFunc calls. want=%d have=%d", 1, len(mockDB.MarkQueuedFunc.History()))
	} else {
		if call := mockDB.MarkQueuedFunc.History()[0]; call.Arg1 != 42 {
			t.Errorf("unexpected upload id. want=%d have=%d", 42, call.Arg1)
		}
	}

	if len(mockBundleManagerClient.StitchPartsFunc.History()) != 1 {
		t.Errorf("unexpected number of StitchPartsFunc calls. want=%d have=%d", 1, len(mockBundleManagerClient.StitchPartsFunc.History()))
	} else {
		if call := mockBundleManagerClient.StitchPartsFunc.History()[0]; call.Arg1 != 42 {
			t.Errorf("unexpected bundle id. want=%d have=%d", 42, call.Arg1)
		}
	}
}

func TestHandleEnqueueMultipartFinalizeIncompleteUpload(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	upload := db.Upload{
		ID:            42,
		NumParts:      5,
		UploadedParts: []int{0, 1, 3, 4},
	}
	mockDB.GetUploadByIDFunc.SetDefaultReturn(upload, true, nil)

	testURL, err := url.Parse("http://test.com/upload")
	if err != nil {
		t.Fatalf("unexpected error constructing url: %s", err)
	}
	testURL.RawQuery = (url.Values{
		"uploadId": []string{"42"},
		"done":     []string{"true"},
	}).Encode()

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", testURL.String(), nil)
	if err != nil {
		t.Fatalf("unexpected error constructing request: %s", err)
	}

	s := &Enqueuer{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
	}
	s.HandleEnqueue(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusBadRequest, w.Code)
	}
}
