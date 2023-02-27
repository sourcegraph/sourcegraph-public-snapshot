package highlight

import (
	"context"
	"encoding/base64"
	"html/template"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestIdentifyError(t *testing.T) {
	errs := []error{gosyntect.ErrPanic, gosyntect.ErrHSSWorkerTimeout, gosyntect.ErrRequestTooLarge}
	for _, err := range errs {
		wrappedErr := errors.Wrap(err, "some other information")
		known, problem := identifyError(wrappedErr)
		require.True(t, known)
		require.NotEqual(t, "", problem)
	}
}

func TestDeserialize(t *testing.T) {
	original := new(scip.Document)
	original.Occurrences = append(original.Occurrences, &scip.Occurrence{
		SyntaxKind: scip.SyntaxKind_IdentifierAttribute,
	})

	marshaled, _ := proto.Marshal(original)
	data, _ := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(marshaled))

	roundtrip := new(scip.Document)
	err := proto.Unmarshal(data, roundtrip)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(original.String(), roundtrip.String()); diff != "" {
		t.Fatalf("Round trip encode and decode should return the same data: %s", diff)
	}
}

func TestGeneratePlainTable(t *testing.T) {
	input := `line 1
line 2

`
	want := template.HTML(`<table><tr><td class="line" data-line="1"></td><td class="code"><span>line 1</span></td></tr><tr><td class="line" data-line="2"></td><td class="code"><span>line 2</span></td></tr><tr><td class="line" data-line="3"></td><td class="code"><span>
</span></td></tr><tr><td class="line" data-line="4"></td><td class="code"><span>
</span></td></tr></table>`)
	response, err := generatePlainTable(input)
	if err != nil {
		t.Fatal(err)
	}

	got, _ := response.HTML()
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestGeneratePlainTableSecurity(t *testing.T) {
	input := `<strong>line 1</strong>
<script>alert("line 2")</script>

`
	want := template.HTML(`<table><tr><td class="line" data-line="1"></td><td class="code"><span>&lt;strong&gt;line 1&lt;/strong&gt;</span></td></tr><tr><td class="line" data-line="2"></td><td class="code"><span>&lt;script&gt;alert(&#34;line 2&#34;)&lt;/script&gt;</span></td></tr><tr><td class="line" data-line="3"></td><td class="code"><span>
</span></td></tr><tr><td class="line" data-line="4"></td><td class="code"><span>
</span></td></tr></table>`)
	response, err := generatePlainTable(input)
	if err != nil {
		t.Fatal(err)
	}

	got, _ := response.HTML()
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestSplitHighlightedLines(t *testing.T) {
	input := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div></div></td></tr></table>`

	want := []template.HTML{
		`<div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div>`,
		`<div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div>`,
		`<div><span style="color:#323232;">)
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div></div>`}

	response := &HighlightedCode{html: template.HTML(input)}
	have, err := response.SplitHighlightedLines(false)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestCodeAsLines(t *testing.T) {
	fileContent := `line1
line2
line3`
	highlightedCode := `<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#657b83;">line 1
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#657b83;">line 2
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#657b83;">line 3</span></div></td></tr></tbody></table>`
	Mocks.Code = func(p Params) (response *HighlightedCode, aborted bool, err error) {
		return &HighlightedCode{
			html: template.HTML(highlightedCode),
		}, false, nil
	}
	t.Cleanup(ResetMocks)

	highlightedLines, aborted, err := CodeAsLines(context.Background(), Params{
		Content:  []byte(fileContent),
		Filepath: "test/file.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if aborted {
		t.Fatalf("highlighting aborted")
	}

	wantLines := []template.HTML{
		"<div><span style=\"color:#657b83;\">line 1\n</span></div>",
		"<div><span style=\"color:#657b83;\">line 2\n</span></div>",
		"<div><span style=\"color:#657b83;\">line 3</span></div>",
	}
	if diff := cmp.Diff(wantLines, highlightedLines); diff != "" {
		t.Fatalf("wrong highlighted lines: %s", diff)
	}
}

func Test_normalizeFilepath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normalize_path",
			input: "a/b/c/FOO.TXT",
			want:  "a/b/c/FOO.txt",
		},
		{
			name:  "normalize_partial_path",
			input: "FOO.Sh",
			want:  "FOO.sh",
		},
		{
			name:  "unmodified_path",
			input: "a/b/c/FOO.txt",
			want:  "a/b/c/FOO.txt",
		},
		{
			name:  "unmodified_path_no_extension",
			input: "a/b/c/Makefile",
			want:  "a/b/c/Makefile",
		},
		{
			name:  "unmodified_partial_path_no_extension",
			input: "Makefile",
			want:  "Makefile",
		},
		{
			name:  "unmodified_partial_path_extension",
			input: "Makefile.am",
			want:  "Makefile.am",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := normalizeFilepath(tst.input)
			if diff := cmp.Diff(got, tst.want); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSplitLineRanges(t *testing.T) {
	html := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div></div></td></tr></table>`

	tests := []struct {
		name  string
		input []LineRange
		want  [][]string
	}{
		{
			name: "clamped_negative",
			input: []LineRange{
				{StartLine: -10, EndLine: 1},
			},
			want: [][]string{
				{
					"<tr><td class=\"line\" data-line=\"1\"></td><td class=\"code\"><div><span style=\"font-weight:bold;color:#a71d5d;\">package</span><span style=\"color:#323232;\"> spans on short lines like this are kept\n</span></div></td></tr>",
				},
			},
		},
		{
			name: "clamped_positive",
			input: []LineRange{
				{StartLine: 0, EndLine: 10000},
			},
			want: [][]string{
				{
					"<tr><td class=\"line\" data-line=\"1\"></td><td class=\"code\"><div><span style=\"font-weight:bold;color:#a71d5d;\">package</span><span style=\"color:#323232;\"> spans on short lines like this are kept\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"2\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"3\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;net/http&#34;\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"4\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"5\"></td><td class=\"code\"><div><span style=\"color:#323232;\">)\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"6\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"7\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"8\"></td><td class=\"code\"><div></div></td></tr>",
				},
			},
		},
		{
			name: "1_range",
			input: []LineRange{
				{StartLine: 3, EndLine: 6},
			},
			want: [][]string{
				{
					"<tr><td class=\"line\" data-line=\"4\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"5\"></td><td class=\"code\"><div><span style=\"color:#323232;\">)\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"6\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
				},
			},
		},
		{
			name: "2_ranges",
			input: []LineRange{
				{StartLine: 1, EndLine: 3},
				{StartLine: 4, EndLine: 6},
			},
			want: [][]string{
				{
					"<tr><td class=\"line\" data-line=\"2\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"3\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;net/http&#34;\n</span></div></td></tr>",
				},
				{
					"<tr><td class=\"line\" data-line=\"5\"></td><td class=\"code\"><div><span style=\"color:#323232;\">)\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"6\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
				},
			},
		},
		{
			name: "3_ranges_unordered",
			input: []LineRange{
				{StartLine: 5, EndLine: 6},
				{StartLine: 7, EndLine: 8},
				{StartLine: 2, EndLine: 4},
			},
			want: [][]string{
				{
					"<tr><td class=\"line\" data-line=\"6\"></td><td class=\"code\"><div><span style=\"color:#323232;\">\n</span></div></td></tr>",
				},
				{
					"<tr><td class=\"line\" data-line=\"8\"></td><td class=\"code\"><div></div></td></tr>",
				},
				{
					"<tr><td class=\"line\" data-line=\"3\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;net/http&#34;\n</span></div></td></tr>",
					"<tr><td class=\"line\" data-line=\"4\"></td><td class=\"code\"><div><span style=\"color:#323232;\">	</span><span style=\"color:#183691;\">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;\n</span></div></td></tr>",
				},
			},
		},
		{
			name: "bad_range",
			input: []LineRange{
				{StartLine: 6, EndLine: 3},
			},
			want: [][]string{
				{},
			},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := SplitLineRanges(template.HTML(html), tst.input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tst.want, got); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
