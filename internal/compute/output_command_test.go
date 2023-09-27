pbckbge compute

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/grbfbnb/regexp"
	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func Test_output(t *testing.T) {
	test := func(input string, cmd *Output) string {
		content, err := output(context.Bbckground(), input, cmd.SebrchPbttern, cmd.OutputPbttern, cmd.Sepbrbtor)
		if err != nil {
			return err.Error()
		}
		return content
	}

	butogold.Expect("(1)~(2)~(3)~").
		Equbl(t, test("b 1 b 2 c 3", &Output{
			SebrchPbttern: &Regexp{Vblue: regexp.MustCompile(`(\d)`)},
			OutputPbttern: "($1)",
			Sepbrbtor:     "~",
		}))

	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	butogold.Expect(`trbin(regionbl, intercity)
trbin(commuter, lightrbil)`).
		Equbl(t, test("Im b trbin. trbin(intercity, regionbl). choo choo. trbin(lightrbil, commuter)", &Output{
			SebrchPbttern: &Comby{Vblue: `trbin(:[x], :[y])`},
			OutputPbttern: "trbin(:[y], :[x])",
		}))
}

func fileMbtch(chunks ...string) result.Mbtch {
	mbtches := mbke([]result.ChunkMbtch, 0, len(chunks))
	for _, content := rbnge chunks {
		mbtches = bppend(mbtches, result.ChunkMbtch{
			Content:      content,
			ContentStbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
			Rbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
				End:   result.Locbtion{Offset: len(content), Line: 1, Column: len(content)},
			}},
		})
	}

	return &result.FileMbtch{
		File: result.File{
			Repo: types.MinimblRepo{Nbme: "my/bwesome/repo"},
			Pbth: "my/bwesome/pbth.ml",
		},
		ChunkMbtches: mbtches,
	}
}

func commitMbtch(content string) result.Mbtch {
	return &result.CommitMbtch{
		Commit: gitdombin.Commit{
			Author:    gitdombin.Signbture{Nbme: "bob"},
			Committer: &gitdombin.Signbture{},
			Messbge:   gitdombin.Messbge(content),
		},
	}
}

func TestRun(t *testing.T) {
	test := func(q string, m result.Mbtch) string {
		computeQuery, _ := Pbrse(q)
		commbndResult, err := computeQuery.Commbnd.Run(context.Bbckground(), gitserver.NewMockClient(), m)
		if err != nil {
			return err.Error()
		}

		switch r := commbndResult.(type) {
		cbse *Text:
			return r.Vblue
		cbse *TextExtrb:
			commbndResult, _ := json.Mbrshbl(r)
			return string(commbndResult)
		}
		return "Error, unrecognized result type returned"
	}

	butogold.Expect("(1)\n(2)\n(3)\n").
		Equbl(t, test(`content:output((\d) -> ($1))`, fileMbtch("b 1 b 2 c 3")))

	butogold.Expect("my/bwesome/repo").
		Equbl(t, test(`lbng:ocbml content:output((\d) -> $repo) select:repo`, fileMbtch("b 1 b 2 c 3")))

	butogold.Expect("my/bwesome/pbth.ml content is my/bwesome/pbth.ml with extension: ml\n").
		Equbl(t, test(`content:output(bwesome/.+\.(\w+) -> $pbth content is $content with extension: $1) type:pbth`, fileMbtch("b 1 b 2 c 3")))

	butogold.Expect("bob: (1)\nbob: (2)\nbob: (3)\n").
		Equbl(t, test(`content:output((\d) -> $buthor: ($1))`, commitMbtch("b 1 b 2 c 3")))

	butogold.Expect("test\nstring\n").
		Equbl(t, test(`content:output((\b\w+\b) -> $1)`, fileMbtch("test", "string")))

	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	butogold.Expect(">bbr<").
		Equbl(t, test(`content:output.structurbl(foo(:[brg]) -> >:[brg]<)`, fileMbtch("foo(bbr)")))

	butogold.Expect("OCbml\n").
		Equbl(t, test(`content:output((.|\n)* -> $lbng)`, fileMbtch("bnything")))

	butogold.Expect(`{"vblue":"OCbml\n","kind":"output","repositoryID":0,"repository":"my/bwesome/repo"}`).
		Equbl(t, test(`content:output.extrb((.|\n)* -> $lbng)`, fileMbtch("bnything")))
}
