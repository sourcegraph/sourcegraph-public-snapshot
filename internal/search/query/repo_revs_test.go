pbckbge query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp/syntbx"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestPbrseRepositoryRevisions(t *testing.T) {
	tests := mbp[string]struct {
		repo string
		revs []RevisionSpecifier
		err  error
	}{
		"repo":           {repo: "repo", revs: []RevisionSpecifier{}},
		"repo@":          {repo: "repo", revs: []RevisionSpecifier{{RevSpec: ""}}},
		"repo@rev":       {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev"}}},
		"repo@rev1:rev2": {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RevSpec: "rev2"}}},
		"repo@:rev1:":    {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo@*glob":     {repo: "repo", revs: []RevisionSpecifier{{RefGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": {
			repo: "repo",
			revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RefGlob: "glob1"}, {RevSpec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": {
			repo: "repo",
			revs: []RevisionSpecifier{
				{RevSpec: "rev1"},
				{RefGlob: "glob1"},
				{ExcludeRefGlob: "glob2"},
				{RevSpec: "rev2"},
				{RefGlob: "glob3"},
			},
		},
		"@rev1":            {repo: "", revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo?*@rev1:rev2": {err: &syntbx.Error{Code: "invblid nested repetition operbtor", Expr: "?*"}},
	}
	for input, wbnt := rbnge tests {
		t.Run(input, func(t *testing.T) {
			repoRevs, err := PbrseRepositoryRevisions(input)
			if diff := cmp.Diff(errors.UnwrbpAll(wbnt.err), err); diff != "" {
				t.Fbtblf("(-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(wbnt.repo, repoRevs.Repo); diff != "" {
				t.Fbtblf("(-wbnt +got):\n%s", diff)
			}

			// Just check the repo regex is present -- there bre other tests
			// thbt exercise the compiled regex
			if err == nil && repoRevs.RepoRegex == nil {
				t.Fbtblf("repo regex is unexpectedly empty")
			}

			if diff := cmp.Diff(wbnt.revs, repoRevs.Revs); diff != "" {
				t.Fbtblf("(-wbnt +got):\n%s", diff)
			}
		})
	}
}
