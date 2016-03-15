package ui

import (
	"reflect"
	"strconv"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/annotations"
)

func blobRenderTestdata(includeRootComponent bool) (blob, StoreData, string) {
	blob := blob{
		Contents: "abcdefghi\njklmnopqr\nst",
		Annotations: &sourcegraph.AnnotationList{
			Annotations: []*sourcegraph.Annotation{
				// Exactly coincident
				{URL: "a", StartByte: 1, EndByte: 2},
				{Class: "a", StartByte: 1, EndByte: 2},

				// Multiple URLs
				{URL: "b", StartByte: 3, EndByte: 6},
				{URL: "b2", StartByte: 3, EndByte: 6},
				{URL: "b3", StartByte: 3, EndByte: 6},
				{Class: "b", StartByte: 3, EndByte: 6},

				{URL: "c", StartByte: 6, EndByte: 12},
				{Class: "c", StartByte: 6, EndByte: 12},

				{URL: "d", StartByte: 13, EndByte: 15},
				{Class: "b", StartByte: 15, EndByte: 18},
				{Class: "c", StartByte: 18, EndByte: 19},
			},
			LineStartBytes: []uint32{0, 10, 20},
		},
		ActiveDef:              "",
		StartLine:              1,
		EndLine:                2,
		LineNumbers:            true,
		HighlightSelectedLines: true,
		VisibleLinesCount:      50,
		reactID:                1,
	}

	blob.Annotations.Annotations = annotations.Prepare(blob.Annotations.Annotations)

	var data StoreData
	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"},
			Rev:      "master",
			CommitID: "c",
		},
		Path: "myfile.txt",
	}
	data.BlobStore.AddFile(entrySpec, &sourcegraph.TreeEntry{ContentsString: blob.Contents})
	data.BlobStore.AddAnnotations(
		&sourcegraph.AnnotationsListOptions{
			Entry: entrySpec,
			Range: &sourcegraph.FileRange{},
		},
		blob.Annotations,
	)

	var want string
	if includeRootComponent {
		want = `<div class="blob-scroller" data-reactid="1" data-remove-second-reactid="yes" data-reactroot="" data-reactid="26" data-react-checksum="261928200">`
	}
	want += `<table class="line-numbered-code" data-reactid="2"><tbody data-reactid="3"><tr class="line main-byte-range" data-line="1" data-reactid="4"><td class="line-number" data-line="1" data-reactid="5"></td><td class="line-content" data-reactid="6"><!-- react-text: 7 -->a<!-- /react-text --><a class="ref" href="a" data-reactid="8"><span class="a" data-reactid="9">b</span></a><!-- react-text: 10 -->c<!-- /react-text --><a class="ref" href="b" data-reactid="11"><span class="b" data-reactid="12">def</span></a><a class="ref" href="c" data-reactid="13"><span class="c" data-reactid="14">ghi</span></a></td></tr><tr class="line main-byte-range" data-line="2" data-reactid="15"><td class="line-number" data-line="2" data-reactid="16"></td><td class="line-content" data-reactid="17"><a class="ref" href="c" data-reactid="18"><span class="c" data-reactid="19">jk</span></a><!-- react-text: 20 -->l<!-- /react-text --><a class="ref" href="d" data-reactid="21">mn</a><span class="b" data-reactid="22">opq</span><span class="c" data-reactid="23">r</span></td></tr><tr class="line" data-line="3" data-reactid="24"><td class="line-number" data-line="3" data-reactid="25"></td><td class="line-content" data-reactid="26">st</td></tr></tbody></table>`
	if includeRootComponent {
		want += "</div>"
	}

	return blob, data, want
}

func TestBlobRenderOptimized(t *testing.T) {
	blob, _, want := blobRenderTestdata(false)

	resp, err := blob.render()
	if err != nil {
		t.Fatal(err)
	}

	if want != resp {
		t.Errorf("got\n%s\n\nwant\n%s", resp, want)
	}
}

func TestAddReactIDs(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"", ""},
		{" ", " "},
		{`<span class="c"></span>`, `<span class="c" data-reactid="5"></span>`},
		{"<span>a</span> ", `<span data-reactid="5">a</span><!-- react-text: 6 --> <!-- /react-text -->`},
		{" <span>a</span>", `<!-- react-text: 5 --> <!-- /react-text --><span data-reactid="6">a</span>`},
		{
			" <span>a</span> ",
			`<!-- react-text: 5 --> <!-- /react-text --><span data-reactid="6">a</span><!-- react-text: 7 --> <!-- /react-text -->`,
		},
		{
			"<span><span>c</span></span>",
			`<span data-reactid="5"><span data-reactid="6">c</span></span>`,
		},
		{
			"<span>a<span>b</span>c</span>",
			`<span data-reactid="5"><!-- react-text: 6 -->a<!-- /react-text --><span data-reactid="7">b</span><!-- react-text: 8 -->c<!-- /react-text --></span>`,
		},
	}
	for _, test := range tests {
		reactID := 4
		nextReactID := func() string {
			reactID++
			return strconv.Itoa(reactID)
		}
		got, err := addReactIDs(nextReactID, []byte(test.src))
		if err != nil {
			t.Errorf("src %q: %s", test.src, err)
			continue
		}
		if string(got) != test.want {
			t.Errorf("src %q: got\n%s\n\nwant\n%s", test.src, got, test.want)
			continue
		}
	}
}

// NOTE: This must be kept in sync with annotationsByLine_test.js.
func TestAnnotationsByLine(t *testing.T) {
	labels := func(annsByLine [][]*sourcegraph.Annotation) [][]string {
		labels := make([][]string, len(annsByLine))
		for i, lineAnns := range annsByLine {
			labels[i] = make([]string, len(lineAnns))
			for j, ann := range lineAnns {
				labels[i][j] = ann.Class
			}
		}
		return labels
	}

	tests := map[string]struct {
		lineStartBytes []uint32
		anns           []*sourcegraph.Annotation
		lines          [][]byte
		wantLabels     [][]string
	}{
		"empty 0": {
			lineStartBytes: nil, anns: nil, lines: [][]byte{},
			wantLabels: [][]string{},
		},
		"empty 1": {
			lineStartBytes: nil, anns: nil, lines: [][]byte{[]byte(nil)},
			wantLabels: [][]string{{}},
		},
		"one line": {
			lineStartBytes: []uint32{0}, anns: nil, lines: [][]byte{[]byte("")},
			wantLabels: [][]string{{}},
		},
		"one line, one ann": {
			lineStartBytes: []uint32{0},
			anns: []*sourcegraph.Annotation{
				{StartByte: 0, EndByte: 2, Class: "a"},
			},
			lines:      [][]byte{[]byte("aaa")},
			wantLabels: [][]string{{"a"}},
		},
		"multiple lines, no cross-line span": {
			lineStartBytes: []uint32{0, 4},
			anns: []*sourcegraph.Annotation{
				{StartByte: 0, EndByte: 2, Class: "a"},
				{StartByte: 2, EndByte: 3, Class: "b"},
				{StartByte: 4, EndByte: 6, Class: "c"},
				{StartByte: 5, EndByte: 7, Class: "d"},
			},
			lines:      [][]byte{[]byte("aaa"), []byte("aaa")},
			wantLabels: [][]string{{"a", "b"}, {"c", "d"}},
		},
		"multiple lines, empty line, no cross-line span": {
			lineStartBytes: []uint32{0, 4, 8},
			anns: []*sourcegraph.Annotation{
				{StartByte: 0, EndByte: 2, Class: "a"},
				{StartByte: 2, EndByte: 3, Class: "b"},
				{StartByte: 8, EndByte: 10, Class: "c"},
				{StartByte: 9, EndByte: 11, Class: "d"},
			},
			lines:      [][]byte{[]byte("aaa"), []byte("aaa"), []byte("aaa")},
			wantLabels: [][]string{{"a", "b"}, {}, {"c", "d"}},
		},
		"cross-line span": {
			lineStartBytes: []uint32{0, 4},
			anns: []*sourcegraph.Annotation{
				{StartByte: 0, EndByte: 3, Class: "a"},
				{StartByte: 1, EndByte: 7, Class: "b"},
				{StartByte: 2, EndByte: 5, Class: "c"},
				{StartByte: 4, EndByte: 5, Class: "d"},
			},
			lines:      [][]byte{[]byte("aaa"), []byte("aaa")},
			wantLabels: [][]string{{"a", "b", "c"}, {"b", "c", "d"}},
		},
		"cross-line span complex": {
			lineStartBytes: []uint32{0, 4, 8},
			anns: []*sourcegraph.Annotation{
				{StartByte: 0, EndByte: 11, Class: "a"},
				{StartByte: 1, EndByte: 5, Class: "b"},
				{StartByte: 4, EndByte: 5, Class: "c"},
				{StartByte: 4, EndByte: 10, Class: "d"},
			},
			lines:      [][]byte{[]byte("aaa"), []byte("aaa"), []byte("aaa")},
			wantLabels: [][]string{{"a", "b"}, {"a", "b", "c", "d"}, {"a", "d"}},
		},
	}

	for name, test := range tests {
		lineAnns := annotationsByLine(&sourcegraph.AnnotationList{
			Annotations:    test.anns,
			LineStartBytes: test.lineStartBytes,
		}, test.lines)
		if labels := labels(lineAnns); !reflect.DeepEqual(labels, test.wantLabels) {
			t.Errorf("%s: got labels %v, want %v", name, labels, test.wantLabels)
			continue
		}
	}
}
