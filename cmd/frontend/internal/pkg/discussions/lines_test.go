package discussions

import (
	"strings"
	"testing"
)

func TestLinesForSelection(t *testing.T) {
	tests := []struct {
		name                                       string
		fileContent                                string
		selection                                  LineRange
		wantLinesBefore, wantLines, wantLinesAfter string
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
			wantLinesBefore:        "before1\nbefore2\nbefore3\n",
			wantLines:              "1\n",
			wantLinesAfter:         "after1\nafter2\nafter3\n",
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
			wantLinesBefore:        "before1\nbefore2\nbefore3\n",
			wantLines:              "1\n2\n",
			wantLinesAfter:         "after1\nafter2\nafter3\n",
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
			wantLinesBefore:        "",
			wantLines:              "zero\none\n",
			wantLinesAfter:         "two\nthree\nfour\n",
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
			wantLinesBefore:        "four\nfive\nsix\n",
			wantLines:              "seven\neight\n",
			wantLinesAfter:         "",
			wantTotalCapturedLines: 6,
		},
		{
			name:                   "one_line",
			selection:              LineRange{StartLine: 0, EndLine: 1},
			fileContent:            `1`,
			wantLinesBefore:        "",
			wantLines:              "1",
			wantLinesAfter:         "",
			wantTotalCapturedLines: 1,
		},
		{
			name:                   "two_lines_top",
			selection:              LineRange{StartLine: 0, EndLine: 1},
			fileContent:            "1\n2\n",
			wantLinesBefore:        "",
			wantLines:              "1\n",
			wantLinesAfter:         "2\n",
			wantTotalCapturedLines: 3,
		},
		{
			name:                   "two_lines_bottom",
			selection:              LineRange{StartLine: 1, EndLine: 2},
			fileContent:            "1\n2",
			wantLinesBefore:        "1\n",
			wantLines:              "2",
			wantLinesAfter:         "",
			wantTotalCapturedLines: 2,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			gotBefore, got, gotAfter := LinesForSelection(tst.fileContent, tst.selection)
			if gotBefore != tst.wantLinesBefore || got != tst.wantLines || gotAfter != tst.wantLinesAfter {
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
			snippet := tst.wantLinesBefore + tst.wantLines + tst.wantLinesAfter
			gotNumLines := len(strings.Split(snippet, "\n"))
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
