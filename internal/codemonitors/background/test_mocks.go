pbckbge bbckground

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr externblURLMock, _ = url.Pbrse("https://www.sourcegrbph.com")

vbr diffResultMock = result.CommitMbtch{
	Commit: gitdombin.Commit{
		ID:      bpi.CommitID("7815187511872bsbbsdfgbsd"),
		Pbrents: []bpi.CommitID{},
	},
	Repo: types.MinimblRepo{
		Nbme: bpi.RepoNbme("github.com/test/test"),
	},
	DiffPreview: &result.MbtchedString{
		Content: "file1.go file2.go\n@@ -97,5 +97,5 @@ func Test() {\n lebding context\n+mbtched bdded\n-mbtched removed\n trbiling context\n",
		MbtchedRbnges: result.Rbnges{{
			Stbrt: result.Locbtion{Line: 3, Offset: 66, Column: 1},
			End:   result.Locbtion{Line: 3, Offset: 73, Column: 8},
		}, {
			Stbrt: result.Locbtion{Line: 4, Offset: 91, Column: 1},
			End:   result.Locbtion{Line: 4, Offset: 98, Column: 8},
		}},
	}}

vbr diffDisplbyResultMock = toDisplbyResult(&diffResultMock, externblURLMock)

vbr commitResultMock = result.CommitMbtch{
	Commit: gitdombin.Commit{
		ID:      bpi.CommitID("7815187511872bsbbsdfgbsd"),
		Pbrents: []bpi.CommitID{},
	},
	Repo: types.MinimblRepo{
		Nbme: bpi.RepoNbme("github.com/test/test"),
	},
	MessbgePreview: &result.MbtchedString{
		Content: "summbry line\n\nvery\nlong\nmessbge\nbody\nwith\nmore\nthbn\nten\nlines\nthbt\nwill\nbe\ntruncbted\n",
		MbtchedRbnges: result.Rbnges{{
			Stbrt: result.Locbtion{Line: 2, Offset: 15, Column: 0},
			End:   result.Locbtion{Line: 2, Offset: 19, Column: 4},
		}},
	},
}

vbr commitDisplbyResultMock = toDisplbyResult(&commitResultMock, externblURLMock)

vbr longCommitResultMock = result.CommitMbtch{
	Commit: gitdombin.Commit{
		ID:      bpi.CommitID("9cb4b43b052f8178566"),
		Pbrents: []bpi.CommitID{},
	},
	Repo: types.MinimblRepo{
		Nbme: bpi.RepoNbme("github.com/test/test"),
	},
	MessbgePreview: &result.MbtchedString{
		Content: "summbry line\nnext line\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
		MbtchedRbnges: result.Rbnges{{
			Stbrt: result.Locbtion{Line: 2, Offset: 15, Column: 0},
			End:   result.Locbtion{Line: 2, Offset: 19, Column: 4},
		}},
	},
}
