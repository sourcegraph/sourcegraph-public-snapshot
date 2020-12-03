package dbstore

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

// testGraph has the following layout:
//
//              +--0c5a779c ------------------------------------------------------+-- [5971b083]
//              |                                                                 |
// [e66e8f9b] --+               +-- 4d36f88b                     +-- [95dd4b2b] --+-- 7cb4a974 --+-- [0ed556d3]
//              |               |                                |
//              +--[f635b8d1] --+               +-- [d6e54842] --+
//                              |               |                |
//                              +-- 026b8df9  --+                +-- [5340d471] --+-- cbc5cf7c
//                                              |
//                                              +-- 6c301adb
var logOutput = `
0ed556d3 7cb4a974
7cb4a974 95dd4b2b
5971b083 0c5a779c 95dd4b2b
95dd4b2b d6e54842
cbc5cf7c 5340d471
5340d471 d6e54842
d6e54842 026b8df9
6c301adb 026b8df9
026b8df9 f635b8d1
4d36f88b f635b8d1
f635b8d1 e66e8f9b
0c5a779c e66e8f9b
`

func TestInternalCalculateVisibleUploads(t *testing.T) {
	commitGraphView := NewCommitGraphView()
	commitGraphView.Add(UploadMeta{UploadID: 50}, "e66e8f9b", "sub1/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 51}, "5971b083", "sub2/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 52}, "f635b8d1", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 53}, "d6e54842", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 54}, "5340d471", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 55}, "95dd4b2b", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 56}, "0ed556d3", "sub3/:lsif-go")

	testGraph := gitserver.ParseCommitGraph(strings.Split(logOutput, "\n"))
	visibleUploads := calculateVisibleUploads(testGraph, commitGraphView)

	for _, uploads := range visibleUploads {
		sort.Slice(uploads, func(i, j int) bool {
			return uploads[i].UploadID-uploads[j].UploadID < 0
		})
	}

	expectedVisibleUploads := map[string][]UploadMeta{
		"e66e8f9b": {{UploadID: 50, Flags: 0}, {UploadID: 51, Flags: 2}, {UploadID: 52, Flags: 1}},
		"0c5a779c": {{UploadID: 50, Flags: 1}, {UploadID: 51, Flags: 1}},
		"f635b8d1": {{UploadID: 50, Flags: 1}, {UploadID: 51, Flags: 4}, {UploadID: 52, Flags: 0}},
		"4d36f88b": {{UploadID: 50, Flags: 2}, {UploadID: 52, Flags: 1}},
		"026b8df9": {{UploadID: 50, Flags: 2}, {UploadID: 51, Flags: 3}, {UploadID: 52, Flags: 1}},
		"6c301adb": {{UploadID: 50, Flags: 3}, {UploadID: 52, Flags: 2}},
		"d6e54842": {{UploadID: 50, Flags: 3}, {UploadID: 51, Flags: 2}, {UploadID: 53, Flags: 0}},
		"5340d471": {{UploadID: 50, Flags: 4}, {UploadID: 54, Flags: 0}},
		"cbc5cf7c": {{UploadID: 50, Flags: 5}, {UploadID: 54, Flags: 1}},
		"95dd4b2b": {{UploadID: 50, Flags: 4}, {UploadID: 51, Flags: 1}, {UploadID: 55, Flags: 0}},
		"5971b083": {{UploadID: 50, Flags: 2}, {UploadID: 51, Flags: 0}, {UploadID: 55, Flags: 1}},
		"7cb4a974": {{UploadID: 50, Flags: 5}, {UploadID: 55, Flags: 1}},
		"0ed556d3": {{UploadID: 50, Flags: 6}, {UploadID: 56, Flags: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, visibleUploads, UploadMetaComparer); diff != "" {
		t.Errorf("unexpected graph (-want +got):\n%s", diff)
	}
}

func TestReverseGraph(t *testing.T) {
	testGraph := gitserver.ParseCommitGraph(strings.Split(logOutput, "\n"))

	reverseGraph := reverseGraph(testGraph.Graph())
	for _, parents := range reverseGraph {
		sort.Strings(parents)
	}

	expectedReverseGraph := map[string][]string{
		"e66e8f9b": {"0c5a779c", "f635b8d1"},
		"0c5a779c": {"5971b083"},
		"f635b8d1": {"026b8df9", "4d36f88b"},
		"4d36f88b": nil,
		"026b8df9": {"6c301adb", "d6e54842"},
		"6c301adb": nil,
		"d6e54842": {"5340d471", "95dd4b2b"},
		"5340d471": {"cbc5cf7c"},
		"cbc5cf7c": nil,
		"95dd4b2b": {"5971b083", "7cb4a974"},
		"5971b083": nil,
		"7cb4a974": {"0ed556d3"},
		"0ed556d3": nil,
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

	_ = calculateVisibleUploads(commitGraph, commitGraphView)
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
