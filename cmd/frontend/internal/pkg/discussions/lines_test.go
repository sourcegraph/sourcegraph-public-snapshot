package discussions

import (
	"reflect"
	"testing"
)

func TestLinesForSelection(t *testing.T) {
	tests := []struct {
		name                                       string
		fileContent                                string
		selection                                  LineRange
		wantLinesBefore, wantLines, wantLinesAfter []string
		wantTotalCapturedLines                     int
	}{
		{
			name:      "basic_single_line",
			selection: LineRange{StartLine: 3, EndLine: 4},
			fileContent: `before1
before2
before3
1
after1
after2
after3
`,
			wantLinesBefore:        []string{"before1", "before2", "before3"},
			wantLines:              []string{"1"},
			wantLinesAfter:         []string{"after1", "after2", "after3"},
			wantTotalCapturedLines: 8,
		},
		{
			name:      "basic_two_lines",
			selection: LineRange{StartLine: 3, EndLine: 5},
			fileContent: `before1
before2
before3
1
2
after1
after2
after3
`,
			wantLinesBefore:        []string{"before1", "before2", "before3"},
			wantLines:              []string{"1", "2"},
			wantLinesAfter:         []string{"after1", "after2", "after3"},
			wantTotalCapturedLines: 9,
		},
		{
			name:      "no_before",
			selection: LineRange{StartLine: 0, EndLine: 2},
			fileContent: `zero
one
two
three
four
five
six
seven
eight
`,
			wantLinesBefore:        []string{},
			wantLines:              []string{"zero", "one"},
			wantLinesAfter:         []string{"two", "three", "four"},
			wantTotalCapturedLines: 6,
		},
		{
			name:      "no_after",
			selection: LineRange{StartLine: 7, EndLine: 9},
			fileContent: `zero
one
two
three
four
five
six
seven
eight
`,
			wantLinesBefore:        []string{"four", "five", "six"},
			wantLines:              []string{"seven", "eight"},
			wantLinesAfter:         []string{""},
			wantTotalCapturedLines: 6,
		},
		{
			name:                   "one_line",
			selection:              LineRange{StartLine: 0, EndLine: 1},
			fileContent:            `1`,
			wantLinesBefore:        []string{},
			wantLines:              []string{"1"},
			wantLinesAfter:         []string{},
			wantTotalCapturedLines: 1,
		},
		{
			name:                   "two_lines_top",
			selection:              LineRange{StartLine: 0, EndLine: 1},
			fileContent:            "1\n2\n",
			wantLinesBefore:        []string{},
			wantLines:              []string{"1"},
			wantLinesAfter:         []string{"2", ""},
			wantTotalCapturedLines: 3,
		},
		{
			name:                   "two_lines_bottom",
			selection:              LineRange{StartLine: 1, EndLine: 2},
			fileContent:            "1\n2",
			wantLinesBefore:        []string{"1"},
			wantLines:              []string{"2"},
			wantLinesAfter:         []string{},
			wantTotalCapturedLines: 2,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			gotBefore, got, gotAfter := LinesForSelection(tst.fileContent, tst.selection)
			if !reflect.DeepEqual(gotBefore, tst.wantLinesBefore) || !reflect.DeepEqual(got, tst.wantLines) || !reflect.DeepEqual(gotAfter, tst.wantLinesAfter) {
				t.Logf("got  before: %q", gotBefore)
				t.Logf("want before: %q", tst.wantLinesBefore)
				t.Log("")
				t.Logf("got  lines: %q", got)
				t.Logf("want lines: %q", tst.wantLines)
				t.Log("")
				t.Logf("got  after: %q", gotAfter)
				t.Logf("want after: %q", tst.wantLinesAfter)
				t.Fail()
			}

			// Ensure that reconstructing the lines together to form a snippet is easy and logical.
			snippet := tst.wantLinesBefore
			snippet = append(snippet, tst.wantLines...)
			snippet = append(snippet, tst.wantLinesAfter...)
			gotNumLines := len(snippet)
			if gotNumLines != tst.wantTotalCapturedLines {
				t.Logf("%q", snippet)
				t.Logf("got %d lines want %d", gotNumLines, tst.wantTotalCapturedLines)
			}

			// Uncomment this to see how hard it is to e.g. turn the lines into
			// plaintext email notifications.
			/*
				withTrailingNewline := func(s string) string {
					if s == "" {
						return ""
					}
					if !strings.HasSuffix(s, "\n") {
						return s + "\n"
					}
					return s
				}
				var b bytes.Buffer
				fmt.Fprint(&b, "\n", tst.name, "\n")
				fmt.Fprint(&b, "***********************************\n")
				fmt.Fprint(&b, withTrailingNewline(tst.wantLinesBefore))
				fmt.Fprint(&b, "-----------------------------------\n")
				fmt.Fprint(&b, withTrailingNewline(tst.wantLines))
				fmt.Fprint(&b, "-----------------------------------\n")
				fmt.Fprint(&b, withTrailingNewline(tst.wantLinesAfter))
				fmt.Fprint(&b, "***********************************")
				t.Log(b.String())
			*/
		})
	}
}
