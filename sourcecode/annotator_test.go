package sourcecode

import (
	"testing"
)

// Ensures that annotator properly splits code to lines
func TestCodeSplit(t *testing.T) {

	for index, tt := range []struct {
		code    string
		offsets [][]int32
	}{
		{
			code: "a\nb\nc\nd",
			offsets: [][]int32{
				{0, 1},
				{2, 3},
				{4, 5},
				{6, 6},
			},
		},
		{
			code: "\n\n\n",
			offsets: [][]int32{
				{0, 0},
				{1, 1},
				{2, 2},
			},
		},
		{
			code: "\n\na",
			offsets: [][]int32{
				{0, 0},
				{1, 1},
				{2, 2},
			},
		},
		{
			code: "\n",
			offsets: [][]int32{
				{0, 0},
			},
		},
		{
			code: "a",
			offsets: [][]int32{
				{0, 0},
			},
		},
	} {
		a := NewNilAnnotator(newFileWithRange([]byte(tt.code)))
		if len(a.Code.Lines) != len(tt.offsets) {
			t.Fatalf("Expected: %d lines, Got: %d at %d\n", len(tt.offsets), len(a.Code.Lines), index)
		}
		for i := 0; i < len(a.Code.Lines); i++ {
			if a.Code.Lines[i].StartByte != tt.offsets[i][0] || a.Code.Lines[i].EndByte != tt.offsets[i][1] {
				t.Fatalf("Expected: [%d-%d], Got: [%d-%d] at %d/%d\n", tt.offsets[i][0], tt.offsets[i][1], a.Code.Lines[i].StartByte, a.Code.Lines[i].EndByte, index, i)
			}
		}
	}
}
