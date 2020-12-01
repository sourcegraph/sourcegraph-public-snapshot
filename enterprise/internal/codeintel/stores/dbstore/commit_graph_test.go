package dbstore

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
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
var testGraph = map[string][]string{
	"e66e8f9b": {},
	"0c5a779c": {"e66e8f9b"},
	"f635b8d1": {"e66e8f9b"},
	"4d36f88b": {"f635b8d1"},
	"026b8df9": {"f635b8d1"},
	"6c301adb": {"026b8df9"},
	"d6e54842": {"026b8df9"},
	"5340d471": {"d6e54842"},
	"cbc5cf7c": {"5340d471"},
	"95dd4b2b": {"d6e54842"},
	"5971b083": {"0c5a779c", "95dd4b2b"},
	"7cb4a974": {"95dd4b2b"},
	"0ed556d3": {"7cb4a974"},
}

func TestInternalCalculateVisibleUploads(t *testing.T) {
	commitGraphView := NewCommitGraphView()
	commitGraphView.Add(UploadMeta{UploadID: 50}, "e66e8f9b", "sub1/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 51}, "5971b083", "sub2/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 52}, "f635b8d1", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 53}, "d6e54842", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 54}, "5340d471", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 55}, "95dd4b2b", "sub3/:lsif-go")
	commitGraphView.Add(UploadMeta{UploadID: 56}, "0ed556d3", "sub3/:lsif-go")

	visibleUploads, err := calculateVisibleUploads(testGraph, commitGraphView)
	if err != nil {
		t.Fatalf("unexpected error calculating visible uploads: %s", err)
	}

	for _, visibleUploads := range visibleUploads {
		sort.Slice(visibleUploads, func(i, j int) bool {
			return visibleUploads[i].UploadID-visibleUploads[j].UploadID < 0
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
	reverseGraph := reverseGraph(testGraph)
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

func TestTopologicalSort(t *testing.T) {
	ordering, err := topologicalSort(testGraph)
	if err != nil {
		t.Fatalf("unexpected error sorting graph: %s", err)
	}

	for commit, parents := range testGraph {
		i, ok := indexOf(ordering, commit)
		if !ok {
			t.Errorf("commit %s missing from ordering", commit)
			continue
		}

		for _, parent := range parents {
			j, ok := indexOf(ordering, parent)
			if !ok {
				t.Errorf("commit %s missing from ordering", commit)
				continue
			}

			if j > i {
				t.Errorf("commit %s and %s are inverted", commit, parent)
			}
		}
	}
}

func TestTopologicalSortCycles(t *testing.T) {
	cyclicTestGraph := map[string][]string{
		"a": nil,
		"b": {"c"},
		"c": {"d"},
		"d": {"d"},
	}
	if _, err := topologicalSort(cyclicTestGraph); err != ErrCyclicCommitGraph {
		t.Fatalf("unexpected error sorting graph. want=%q have=%q", ErrCyclicCommitGraph, err)
	}
}

func BenchmarkCalculateVisibleUploads(b *testing.B) {
	graph, commitGraphView := makeBenchmarkGraph()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = calculateVisibleUploads(graph, commitGraphView)
	}
}

func makeBenchmarkGraph() (map[string][]string, *CommitGraphView) {
	var (
		numCommits = 5000
		numUploads = 2000
		numRoots   = 500
	)

	var commits []string
	for i := 0; i < numCommits; i++ {
		commits = append(commits, fmt.Sprintf("%40d", i))
	}

	graph := map[string][]string{}
	for i, commit := range commits {
		if i == 0 {
			graph[commit] = nil
		} else {
			graph[commit] = commits[i-1 : i]
		}
	}

	var roots []string
	for i := 0; i < numRoots; i++ {
		roots = append(roots, fmt.Sprintf("sub%d/", i))
	}

	commitGraphView := NewCommitGraphView()

	for i := 0; i < numUploads; i++ {
		commitGraphView.Add(
			UploadMeta{UploadID: i},
			fmt.Sprintf("%40d", int((float64(i*numUploads)/float64(numCommits)))),
			fmt.Sprintf("%s:lsif-go", roots[i%numRoots]),
		)
	}

	return graph, commitGraphView
}

func indexOf(ordering []string, commit string) (int, bool) {
	for i, v := range ordering {
		if v == commit {
			return i, true
		}
	}

	return 0, false
}
