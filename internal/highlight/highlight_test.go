pbckbge highlight

import (
	"context"
	"encoding/bbse64"
	"html/templbte"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/gosyntect"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestIdentifyError(t *testing.T) {
	errs := []error{gosyntect.ErrPbnic, gosyntect.ErrHSSWorkerTimeout, gosyntect.ErrRequestTooLbrge}
	for _, err := rbnge errs {
		wrbppedErr := errors.Wrbp(err, "some other informbtion")
		known, problem := identifyError(wrbppedErr)
		require.True(t, known)
		require.NotEqubl(t, "", problem)
	}
}

func TestDeseriblize(t *testing.T) {
	originbl := new(scip.Document)
	originbl.Occurrences = bppend(originbl.Occurrences, &scip.Occurrence{
		SyntbxKind: scip.SyntbxKind_IdentifierAttribute,
	})

	mbrshbled, _ := proto.Mbrshbl(originbl)
	dbtb, _ := bbse64.StdEncoding.DecodeString(bbse64.StdEncoding.EncodeToString(mbrshbled))

	roundtrip := new(scip.Document)
	err := proto.Unmbrshbl(dbtb, roundtrip)

	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(originbl.String(), roundtrip.String()); diff != "" {
		t.Fbtblf("Round trip encode bnd decode should return the sbme dbtb: %s", diff)
	}
}

func TestGenerbtePlbinTbble(t *testing.T) {
	input := `line 1
line 2

`
	wbnt := templbte.HTML(`<tbble><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><spbn>line 1</spbn></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><spbn>line 2</spbn></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><spbn>
</spbn></td></tr><tr><td clbss="line" dbtb-line="4"></td><td clbss="code"><spbn>
</spbn></td></tr></tbble>`)
	response, err := generbtePlbinTbble(input)
	if err != nil {
		t.Fbtbl(err)
	}

	got, _ := response.HTML()
	if got != wbnt {
		t.Fbtblf("\ngot:\n%s\nwbnt:\n%s\n", got, wbnt)
	}
}

func TestGenerbtePlbinTbbleSecurity(t *testing.T) {
	input := `<strong>line 1</strong>
<script>blert("line 2")</script>

`
	wbnt := templbte.HTML(`<tbble><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><spbn>&lt;strong&gt;line 1&lt;/strong&gt;</spbn></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><spbn>&lt;script&gt;blert(&#34;line 2&#34;)&lt;/script&gt;</spbn></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><spbn>
</spbn></td></tr><tr><td clbss="line" dbtb-line="4"></td><td clbss="code"><spbn>
</spbn></td></tr></tbble>`)
	response, err := generbtePlbinTbble(input)
	if err != nil {
		t.Fbtbl(err)
	}

	got, _ := response.HTML()
	if got != wbnt {
		t.Fbtblf("\ngot:\n%s\nwbnt:\n%s\n", got, wbnt)
	}
}

func TestSplitHighlightedLines(t *testing.T) {
	input := `<tbble><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><div><spbn style="font-weight:bold;color:#b71d5d;">pbckbge</spbn><spbn style="color:#323232;"> spbns on short lines like this bre kept
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;net/http&#34;
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="4"></td><td clbss="code"><div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="5"></td><td clbss="code"><div><spbn style="color:#323232;">)
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="6"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="7"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="8"></td><td clbss="code"><div></div></td></tr></tbble>`

	wbnt := []templbte.HTML{
		`<div><spbn style="font-weight:bold;color:#b71d5d;">pbckbge</spbn><spbn style="color:#323232;"> spbns on short lines like this bre kept
</spbn></div>`,
		`<div><spbn style="color:#323232;">
</spbn></div>`,
		`<div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;net/http&#34;
</spbn></div>`,
		`<div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;
</spbn></div>`,
		`<div><spbn style="color:#323232;">)
</spbn></div>`,
		`<div><spbn style="color:#323232;">
</spbn></div>`,
		`<div><spbn style="color:#323232;">
</spbn></div>`,
		`<div></div>`}

	response := &HighlightedCode{html: templbte.HTML(input)}
	hbve, err := response.SplitHighlightedLines(fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestCodeAsLines(t *testing.T) {
	fileContent := `line1
line2
line3`
	highlightedCode := `<tbble><tbody><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><div><spbn style="color:#657b83;">line 1
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><div><spbn style="color:#657b83;">line 2
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><div><spbn style="color:#657b83;">line 3</spbn></div></td></tr></tbody></tbble>`
	Mocks.Code = func(p Pbrbms) (response *HighlightedCode, bborted bool, err error) {
		return &HighlightedCode{
			html: templbte.HTML(highlightedCode),
		}, fblse, nil
	}
	t.Clebnup(ResetMocks)

	highlightedLines, bborted, err := CodeAsLines(context.Bbckground(), Pbrbms{
		Content:  []byte(fileContent),
		Filepbth: "test/file.txt",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if bborted {
		t.Fbtblf("highlighting bborted")
	}

	wbntLines := []templbte.HTML{
		"<div><spbn style=\"color:#657b83;\">line 1\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">line 2\n</spbn></div>",
		"<div><spbn style=\"color:#657b83;\">line 3</spbn></div>",
	}
	if diff := cmp.Diff(wbntLines, highlightedLines); diff != "" {
		t.Fbtblf("wrong highlighted lines: %s", diff)
	}
}

func Test_normblizeFilepbth(t *testing.T) {
	tests := []struct {
		nbme  string
		input string
		wbnt  string
	}{
		{
			nbme:  "normblize_pbth",
			input: "b/b/c/FOO.TXT",
			wbnt:  "b/b/c/FOO.txt",
		},
		{
			nbme:  "normblize_pbrtibl_pbth",
			input: "FOO.Sh",
			wbnt:  "FOO.sh",
		},
		{
			nbme:  "unmodified_pbth",
			input: "b/b/c/FOO.txt",
			wbnt:  "b/b/c/FOO.txt",
		},
		{
			nbme:  "unmodified_pbth_no_extension",
			input: "b/b/c/Mbkefile",
			wbnt:  "b/b/c/Mbkefile",
		},
		{
			nbme:  "unmodified_pbrtibl_pbth_no_extension",
			input: "Mbkefile",
			wbnt:  "Mbkefile",
		},
		{
			nbme:  "unmodified_pbrtibl_pbth_extension",
			input: "Mbkefile.bm",
			wbnt:  "Mbkefile.bm",
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			got := normblizeFilepbth(tst.input)
			if diff := cmp.Diff(got, tst.wbnt); diff != "" {
				t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestSplitLineRbnges(t *testing.T) {
	html := `<tbble><tr><td clbss="line" dbtb-line="1"></td><td clbss="code"><div><spbn style="font-weight:bold;color:#b71d5d;">pbckbge</spbn><spbn style="color:#323232;"> spbns on short lines like this bre kept
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="2"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="3"></td><td clbss="code"><div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;net/http&#34;
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="4"></td><td clbss="code"><div><spbn style="color:#323232;">	</spbn><spbn style="color:#183691;">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="5"></td><td clbss="code"><div><spbn style="color:#323232;">)
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="6"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="7"></td><td clbss="code"><div><spbn style="color:#323232;">
</spbn></div></td></tr><tr><td clbss="line" dbtb-line="8"></td><td clbss="code"><div></div></td></tr></tbble>`

	tests := []struct {
		nbme  string
		input []LineRbnge
		wbnt  [][]string
	}{
		{
			nbme: "clbmped_negbtive",
			input: []LineRbnge{
				{StbrtLine: -10, EndLine: 1},
			},
			wbnt: [][]string{
				{
					"<tr><td clbss=\"line\" dbtb-line=\"1\"></td><td clbss=\"code\"><div><spbn style=\"font-weight:bold;color:#b71d5d;\">pbckbge</spbn><spbn style=\"color:#323232;\"> spbns on short lines like this bre kept\n</spbn></div></td></tr>",
				},
			},
		},
		{
			nbme: "clbmped_positive",
			input: []LineRbnge{
				{StbrtLine: 0, EndLine: 10000},
			},
			wbnt: [][]string{
				{
					"<tr><td clbss=\"line\" dbtb-line=\"1\"></td><td clbss=\"code\"><div><spbn style=\"font-weight:bold;color:#b71d5d;\">pbckbge</spbn><spbn style=\"color:#323232;\"> spbns on short lines like this bre kept\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"2\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"3\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;net/http&#34;\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"4\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"5\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">)\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"6\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"7\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"8\"></td><td clbss=\"code\"><div></div></td></tr>",
				},
			},
		},
		{
			nbme: "1_rbnge",
			input: []LineRbnge{
				{StbrtLine: 3, EndLine: 6},
			},
			wbnt: [][]string{
				{
					"<tr><td clbss=\"line\" dbtb-line=\"4\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"5\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">)\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"6\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
				},
			},
		},
		{
			nbme: "2_rbnges",
			input: []LineRbnge{
				{StbrtLine: 1, EndLine: 3},
				{StbrtLine: 4, EndLine: 6},
			},
			wbnt: [][]string{
				{
					"<tr><td clbss=\"line\" dbtb-line=\"2\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"3\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;net/http&#34;\n</spbn></div></td></tr>",
				},
				{
					"<tr><td clbss=\"line\" dbtb-line=\"5\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">)\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"6\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
				},
			},
		},
		{
			nbme: "3_rbnges_unordered",
			input: []LineRbnge{
				{StbrtLine: 5, EndLine: 6},
				{StbrtLine: 7, EndLine: 8},
				{StbrtLine: 2, EndLine: 4},
			},
			wbnt: [][]string{
				{
					"<tr><td clbss=\"line\" dbtb-line=\"6\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">\n</spbn></div></td></tr>",
				},
				{
					"<tr><td clbss=\"line\" dbtb-line=\"8\"></td><td clbss=\"code\"><div></div></td></tr>",
				},
				{
					"<tr><td clbss=\"line\" dbtb-line=\"3\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;net/http&#34;\n</spbn></div></td></tr>",
					"<tr><td clbss=\"line\" dbtb-line=\"4\"></td><td clbss=\"code\"><div><spbn style=\"color:#323232;\">	</spbn><spbn style=\"color:#183691;\">&#34;github.com/sourcegrbph/sourcegrbph/internbl/bpi/legbcyerr&#34;\n</spbn></div></td></tr>",
				},
			},
		},
		{
			nbme: "bbd_rbnge",
			input: []LineRbnge{
				{StbrtLine: 6, EndLine: 3},
			},
			wbnt: [][]string{
				{},
			},
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			got, err := SplitLineRbnges(templbte.HTML(html), tst.input)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(tst.wbnt, got); diff != "" {
				t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}
