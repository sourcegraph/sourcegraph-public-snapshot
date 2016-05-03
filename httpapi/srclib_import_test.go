package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srclibstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sqs/pbtypes"
)

func TestSrclibImport(t *testing.T) {
	c, mock := newTest()

	const (
		wantRepo     = "r"
		wantCommitID = "c"
	)

	// Sample srclib data to import.
	files := map[string]interface{}{
		"a/b.unit.json":    &unit.SourceUnit{Name: "a/b", Type: "t", Dir: ".", Files: []string{"f"}},
		"a/b/t.graph.json": graph.Output{Defs: []*graph.Def{{DefKey: graph.DefKey{Path: "p"}, Name: "n", File: "f"}}},
	}

	calledReposGet := mock.Repos.MockGet(t, wantRepo)
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, wantCommitID)
	mock.Search.RefreshIndex_ = func(ctx context.Context, in *sourcegraph.SearchRefreshIndexOp) (*pbtypes.Void, error) {
		return nil, nil
	}

	// Mock the srclib store interface (and replace the old
	// newSrclibStoreClient value when done).
	var calledSrclibStoreImport, calledSrclibStoreIndex, calledSrclibStoreCreateVersion bool
	orig := newSrclibStoreClient
	newSrclibStoreClient = func(context.Context, pb.MultiRepoImporterClient) pb.MultiRepoImporterIndexer {
		return srclibstore.MockMultiRepoStore{
			Import_: func(repo, commitID string, unit *unit.SourceUnit, data graph.Output) error {
				calledSrclibStoreImport = true
				if repo != wantRepo {
					t.Errorf("got repo %q, want %q", repo, wantRepo)
				}
				if commitID != wantCommitID {
					t.Errorf("got commitID %q, want %q", commitID, wantCommitID)
				}
				if want := files["a/b.unit.json"]; !reflect.DeepEqual(unit, want) {
					t.Errorf("got unit %+v, want %+v", unit, want)
				}
				if want := files["a/b/t.graph.json"]; !reflect.DeepEqual(data, want) {
					t.Errorf("got graph data %+v, want %+v", data, want)
				}
				return nil
			},
			Index_: func(repo, commitID string) error {
				calledSrclibStoreIndex = true
				if repo != wantRepo {
					t.Errorf("got repo %q, want %q", repo, wantRepo)
				}
				if commitID != wantCommitID {
					t.Errorf("got commitID %q, want %q", commitID, wantCommitID)
				}
				return nil
			},
			CreateVersion_: func(repo, commitID string) error {
				calledSrclibStoreCreateVersion = true
				if repo != wantRepo {
					t.Errorf("got repo %q, want %q", repo, wantRepo)
				}
				if commitID != wantCommitID {
					t.Errorf("got commitID %q, want %q", commitID, wantCommitID)
				}
				return nil
			},
		}
	}
	defer func() { newSrclibStoreClient = orig }()

	// Create a dummy srclib archive.
	var zipData bytes.Buffer
	zipW := zip.NewWriter(&zipData)
	for name, v := range files {
		f, err := zipW.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewEncoder(f).Encode(v); err != nil {
			t.Fatal(err)
		}
	}
	if err := zipW.Close(); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/repos/r@v/-/srclib-import", &zipData)
	req.Header.Set("content-type", "application/x-zip-compressed")
	req.Header.Set("content-transfer-encoding", "binary")
	if _, err := c.DoOK(req); err != nil {
		t.Fatal(err)
	}
	if !calledSrclibStoreImport {
		t.Error("!calledSrclibStoreImport")
	}
	if !calledSrclibStoreIndex {
		t.Error("!calledSrclibStoreIndex")
	}
	if !calledSrclibStoreCreateVersion {
		t.Error("!calledSrclibStoreCreateVersion")
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}

func TestSrclibImport_empty(t *testing.T) {
	c, mock := newTest()

	calledReposGet := mock.Repos.MockGet(t, "r")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")

	// POST an empty zip archive.
	req, _ := http.NewRequest("PUT", "/repos/r@v/-/srclib-import", nil)
	req.Header.Set("content-type", "application/x-zip-compressed")
	req.Header.Set("content-transfer-encoding", "binary")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if want := http.StatusBadRequest; resp.StatusCode != want {
		t.Errorf("got HTTP response status %d, want %d", resp.StatusCode, want)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
