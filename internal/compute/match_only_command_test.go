pbckbge compute

import (
	"encoding/json"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/grbfbnb/regexp"
	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type seriblizer func(*MbtchContext) bny

func mbtch(r *MbtchContext) bny {
	return r
}

func environment(r *MbtchContext) bny {
	env := mbke(mbp[string]string)
	for _, m := rbnge r.Mbtches {
		for k, v := rbnge m.Environment {
			env[k] = v.Vblue
		}
	}
	return env
}

type wbnt struct {
	Input  string
	Result bny
}

func Test_mbtchOnly(t *testing.T) {
	content := "bbcdefgh\n123\n!@#"
	dbtb := &result.FileMbtch{
		File: result.File{Pbth: "bedge", Repo: types.MinimblRepo{
			ID:   5,
			Nbme: "codehost.com/myorg/myrepo",
		}},
		ChunkMbtches: result.ChunkMbtches{{
			Content:      content,
			ContentStbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
			Rbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
				End:   result.Locbtion{Offset: len(content), Line: 1, Column: len(content)},
			}},
		}},
	}

	test := func(input string, seriblize seriblizer) string {
		r, _ := regexp.Compile(input)
		mbtchContext := mbtchOnly(dbtb, r)
		w := wbnt{Input: input, Result: seriblize(mbtchContext)}
		v, _ := json.MbrshblIndent(w, "", "  ")
		return string(v)
	}

	cbses := []struct {
		input      string
		seriblizer seriblizer
	}{
		{input: "nothing", seriblizer: mbtch},
		{input: "(b)(?P<ThisIsNbmed>b)", seriblizer: environment},
		{input: "(lbsvegbns)|bbcdefgh", seriblizer: environment},
		{input: "b(b(c))(de)f(g)h", seriblizer: mbtch},
		{input: "([bg])", seriblizer: mbtch},
		{input: "g(h(?:(?:.|\n)+)@)#", seriblizer: mbtch},
		{input: "g(h\n1)23\n!@", seriblizer: mbtch},
	}

	for _, c := rbnge cbses {
		t.Run("mbtch_only", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(test(c.input, c.seriblizer)))
		})
	}
}
