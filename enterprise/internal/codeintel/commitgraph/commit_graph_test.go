package commitgraph

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

func TestCalculateVisibleUploads(t *testing.T) {
	// testGraph has the following layout:
	//
	//       +--- b -------------------------------+-- [j]
	//       |                                     |
	// [a] --+         +-- d             +-- [h] --+--- k -- [m]
	//       |         |                 |
	//       +-- [c] --+       +-- [f] --+
	//                 |       |         |
	//                 +-- e --+         +-- [i] ------ l -- [n]
	//                         |
	//                         +--- g
	//
	// NOTE: The input to ParseCommitGraph must match the order and format
	// of `git log --topo-sort`.
	testGraph := gitserver.ParseCommitGraph([]string{
		"n l",
		"m k",
		"k h",
		"j b h",
		"h f",
		"l i",
		"i f",
		"f e",
		"g e",
		"e c",
		"d c",
		"c a",
		"b a",
	})

	commitGraphView := NewCommitGraphView()
	commitGraphView.Add(UploadMeta{UploadID: 45}, "n", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 50}, "a", "sub1/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 51}, "j", "sub2/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 52}, "c", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 53}, "f", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 54}, "i", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 55}, "h", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 56}, "m", "sub3/:lsif-go")

	visibleUploads := map[string][]UploadMeta{}
	for v := range CalculateVisibleUploads(testGraph, commitGraphView) {
		visibleUploads[v.Commit] = v.Uploads
	}

	for _, uploads := range visibleUploads {
		sort.Slice(uploads, func(i, j int) bool {
			return uploads[i].UploadID-uploads[j].UploadID < 0
		})
	}

	expectedVisibleUploads := map[string][]UploadMeta{
		"a": {
			{UploadID: 50, Flags: 0 | FlagAncestorVisible},
			{UploadID: 51, Flags: 2},
			{UploadID: 52, Flags: 1},
		},
		"b": {
			{UploadID: 50, Flags: 1 | FlagAncestorVisible},
			{UploadID: 51, Flags: 1},
		},
		"c": {
			{UploadID: 50, Flags: 1 | FlagAncestorVisible},
			{UploadID: 51, Flags: 4},
			{UploadID: 52, Flags: 0 | FlagAncestorVisible},
		},
		"d": {
			{UploadID: 50, Flags: 2 | FlagAncestorVisible},
			{UploadID: 52, Flags: 1 | FlagAncestorVisible},
		},
		"e": {
			{UploadID: 50, Flags: 2 | FlagAncestorVisible},
			{UploadID: 51, Flags: 3},
			{UploadID: 52, Flags: 1 | FlagAncestorVisible},
		},
		"g": {
			{UploadID: 50, Flags: 3 | FlagAncestorVisible},
			{UploadID: 52, Flags: 2 | FlagAncestorVisible},
		},
		"f": {
			{UploadID: 50, Flags: 3 | FlagAncestorVisible},
			{UploadID: 51, Flags: 2},
			{UploadID: 53, Flags: 0 | FlagAncestorVisible},
		},
		"i": {
			{UploadID: 50, Flags: 4 | FlagAncestorVisible},
			{UploadID: 54, Flags: 0 | FlagAncestorVisible},
		},
		"l": {
			{UploadID: 45, Flags: 1},
			{UploadID: 50, Flags: 5 | FlagAncestorVisible},
			{UploadID: 54, Flags: 1 | FlagAncestorVisible | FlagOverwritten},
		},
		"h": {
			{UploadID: 50, Flags: 4 | FlagAncestorVisible},
			{UploadID: 51, Flags: 1},
			{UploadID: 55, Flags: 0 | FlagAncestorVisible},
		},
		"j": {
			{UploadID: 50, Flags: 2 | FlagAncestorVisible},
			{UploadID: 51, Flags: 0 | FlagAncestorVisible},
			{UploadID: 55, Flags: 1 | FlagAncestorVisible},
		},
		"k": {
			{UploadID: 50, Flags: 5 | FlagAncestorVisible},
			{UploadID: 55, Flags: 1 | FlagAncestorVisible},
		},
		"m": {
			{UploadID: 50, Flags: 6 | FlagAncestorVisible},
			{UploadID: 56, Flags: 0 | FlagAncestorVisible},
		},
		"n": {
			{UploadID: 45, Flags: 0 | FlagAncestorVisible},
			{UploadID: 50, Flags: 6 | FlagAncestorVisible},
		},
	}
	if diff := cmp.Diff(expectedVisibleUploads, visibleUploads); diff != "" {
		t.Errorf("unexpected graph (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsAlternateCommitGraph(t *testing.T) {
	// testGraph has the following layout:
	//
	//       [b] ------+                                          +------ n --- p
	//                 |                                          |
	//             +-- d --+                                  +-- l --+
	//             |       |                                  |       |
	// [a] -- c ---+       +-- f -- g -- h -- [i] -- j -- k --+       +-- o -- [q]
	//             |       |                                  |       |
	//             +-- e --+                                  +-- m --+
	//
	// NOTE: The input to ParseCommitGraph must match the order and format
	// of `git log --topo-sort`.
	testGraph := gitserver.ParseCommitGraph([]string{
		"q o",
		"p n",
		"o l m",
		"n l",
		"m k",
		"l k",
		"k j",
		"j i",
		"i h",
		"h g",
		"g f",
		"f d e",
		"e c",
		"d b c",
		"c a",
	})

	commitGraphView := NewCommitGraphView()
	commitGraphView.Add(UploadMeta{UploadID: 50}, "a", "sub1/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 51}, "b", "sub1/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 52}, "i", "sub2/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 53}, "q", "sub3/:lsif-go")

	visibleUploads := map[string][]UploadMeta{}
	for v := range CalculateVisibleUploads(testGraph, commitGraphView) {
		visibleUploads[v.Commit] = v.Uploads
	}

	for _, uploads := range visibleUploads {
		sort.Slice(uploads, func(i, j int) bool {
			return uploads[i].UploadID-uploads[j].UploadID < 0
		})
	}

	expectedVisibleUploads := map[string][]UploadMeta{
		"a": {
			{UploadID: 50, Flags: 0 | FlagAncestorVisible},
			{UploadID: 52, Flags: 6},
			{UploadID: 53, Flags: 11},
		},
		"b": {
			{UploadID: 51, Flags: 0 | FlagAncestorVisible},
			{UploadID: 52, Flags: 5},
			{UploadID: 53, Flags: 10},
		},
		"c": {
			{UploadID: 50, Flags: 1 | FlagAncestorVisible},
			{UploadID: 52, Flags: 5},
			{UploadID: 53, Flags: 10},
		},
		"d": {
			{UploadID: 51, Flags: 1 | FlagAncestorVisible},
			{UploadID: 52, Flags: 4},
			{UploadID: 53, Flags: 9},
		},
		"e": {
			{UploadID: 50, Flags: 2 | FlagAncestorVisible},
			{UploadID: 52, Flags: 4},
			{UploadID: 53, Flags: 9},
		},
		"f": {
			{UploadID: 51, Flags: 2 | FlagAncestorVisible},
			{UploadID: 52, Flags: 3},
			{UploadID: 53, Flags: 8},
		},
		"g": {
			{UploadID: 51, Flags: 3 | FlagAncestorVisible},
			{UploadID: 52, Flags: 2},
			{UploadID: 53, Flags: 7},
		},
		"h": {
			{UploadID: 51, Flags: 4 | FlagAncestorVisible},
			{UploadID: 52, Flags: 1},
			{UploadID: 53, Flags: 6},
		},
		"i": {
			{UploadID: 51, Flags: 5 | FlagAncestorVisible},
			{UploadID: 52, Flags: 0 | FlagAncestorVisible},
			{UploadID: 53, Flags: 5},
		},
		"j": {
			{UploadID: 51, Flags: 6 | FlagAncestorVisible},
			{UploadID: 52, Flags: 1 | FlagAncestorVisible},
			{UploadID: 53, Flags: 4},
		},
		"k": {
			{UploadID: 51, Flags: 7 | FlagAncestorVisible},
			{UploadID: 52, Flags: 2 | FlagAncestorVisible},
			{UploadID: 53, Flags: 3},
		},
		"l": {
			{UploadID: 51, Flags: 8 | FlagAncestorVisible},
			{UploadID: 52, Flags: 3 | FlagAncestorVisible},
			{UploadID: 53, Flags: 2},
		},
		"m": {
			{UploadID: 51, Flags: 8 | FlagAncestorVisible},
			{UploadID: 52, Flags: 3 | FlagAncestorVisible},
			{UploadID: 53, Flags: 2},
		},
		"n": {
			{UploadID: 51, Flags: 9 | FlagAncestorVisible},
			{UploadID: 52, Flags: 4 | FlagAncestorVisible},
		},
		"o": {
			{UploadID: 51, Flags: 9 | FlagAncestorVisible},
			{UploadID: 52, Flags: 4 | FlagAncestorVisible},
			{UploadID: 53, Flags: 1},
		},
		"p": {
			{UploadID: 51, Flags: 10 | FlagAncestorVisible},
			{UploadID: 52, Flags: 5 | FlagAncestorVisible},
		},
		"q": {
			{UploadID: 51, Flags: 10 | FlagAncestorVisible},
			{UploadID: 52, Flags: 5 | FlagAncestorVisible},
			{UploadID: 53, Flags: 0 | FlagAncestorVisible},
		},
	}
	if diff := cmp.Diff(expectedVisibleUploads, visibleUploads); diff != "" {
		t.Errorf("unexpected graph (-want +got):\n%s", diff)
	}
}

func TestReverseGraph(t *testing.T) {
	reverseGraph := reverseGraph(map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"e", "f"},
		"d": {},
		"e": {"f"},
		"f": {"g"},
	})
	for _, parents := range reverseGraph {
		sort.Strings(parents)
	}

	expectedReverseGraph := map[string][]string{
		"a": nil,
		"b": {"a"},
		"c": {"a"},
		"d": {"b"},
		"e": {"c"},
		"f": {"c", "e"},
		"g": {"f"},
	}
	if diff := cmp.Diff(expectedReverseGraph, reverseGraph); diff != "" {
		t.Errorf("unexpected graph (-want +got):\n%s", diff)
	}
}

func BenchmarkCalculateVisibleUploads(b *testing.B) {
	commitGraph, err := readBenchmarkCommitGraph()
	if err != nil {
		b.Fatalf("failed to read benchmark commit graph: %s", err)
	}
	commitGraphView, err := readBenchmarkCommitGraphView()
	if err != nil {
		b.Fatalf("failed to read benchmark commit graph view: %s", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for range CalculateVisibleUploads(commitGraph, commitGraphView) {
		}
	}
}

func readBenchmarkCommitGraph() (*gitserver.CommitGraph, error) {
	contents, err := readBenchmarkFile("./testdata/commits.txt.gz")
	if err != nil {
		return nil, err
	}

	return gitserver.ParseCommitGraph(strings.Split(string(contents), "\n")), nil
}

func readBenchmarkCommitGraphView() (*CommitGraphView, error) {
	contents, err := readBenchmarkFile("./testdata/uploads.txt.gz")
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(contents))

	commitGraphView := NewCommitGraphView()
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		commitGraphView.Add(
			UploadMeta{UploadID: id},             // meta
			record[1],                            // commit
			fmt.Sprintf("%s:lsif-go", record[2]), // token = hash({root}:{indexer})
		)
	}

	return commitGraphView, nil
}

func readBenchmarkFile(path string) ([]byte, error) {
	uploadsFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer uploadsFile.Close()

	r, err := gzip.NewReader(uploadsFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
