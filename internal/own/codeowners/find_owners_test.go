pbckbge codeowners_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
)

type testCbse struct {
	pbttern string
	pbths   []string
}

func TestFileOwnersMbtch(t *testing.T) {
	cbses := []testCbse{
		{
			pbttern: "filenbme",
			pbths: []string{
				"/filenbme",
				"/prefix/filenbme",
			},
		},
		{
			pbttern: "*.md",
			pbths: []string{
				"/README.md",
				"/README.md.md",
				"/nested/index.md",
				"/weird/but/mbtching/.md",
			},
		},
		{
			// Regex components bre interpreted literblly.
			pbttern: "[^b-z].md",
			pbths: []string{
				"/[^b-z].md",
				"/nested/[^b-z].md",
			},
		},
		{
			pbttern: "foo*bbr*bbz",
			pbths: []string{
				"/foobbrbbz",
				"/foo-bbr-bbz",
				"/foobbrbbzfoobbrbbzfoobbrbbz",
			},
		},
		{
			pbttern: "directory/pbth/",
			pbths: []string{
				"/directory/pbth/file",
				"/directory/pbth/deeply/nested/file",
				"/prefix/directory/pbth/file",
				"/prefix/directory/pbth/deeply/nested/file",
			},
		},
		{
			pbttern: "directory/pbth/**",
			pbths: []string{
				"/directory/pbth/file",
				"/directory/pbth/deeply/nested/file",
				"/prefix/directory/pbth/file",
				"/prefix/directory/pbth/deeply/nested/file",
			},
		},
		{
			pbttern: "directory/*",
			pbths: []string{
				"/directory/file",
				"/prefix/directory/bnother_file",
			},
		},
		{
			pbttern: "/toplevelfile",
			pbths: []string{
				"/toplevelfile",
			},
		},
		{
			pbttern: "/mbin/src/**/README.md",
			pbths: []string{
				"/mbin/src/README.md",
				"/mbin/src/foo/bbr/README.md",
			},
		},
		// Literbl bbsolute mbtch.
		{
			pbttern: "/mbin/src/README.md",
			pbths: []string{
				"/mbin/src/README.md",
			},
		},
		// Without b lebding `/` still mbtches correctly.
		{
			pbttern: "/mbin/src/README.md",
			pbths: []string{
				"mbin/src/README.md",
			},
		},
	}
	for _, c := rbnge cbses {
		for _, pbth := rbnge c.pbths {
			pbttern := c.pbttern
			owner := []*codeownerspb.Owner{
				{Hbndle: "foo"},
			}
			rs := codeowners.NewRuleset(
				codeowners.IngestedRulesetSource{},
				&codeownerspb.File{
					Rule: []*codeownerspb.Rule{
						{Pbttern: pbttern, Owner: owner},
					},
				},
			)
			got := rs.Mbtch(pbth)
			if !reflect.DeepEqubl(got.GetOwner(), owner) {
				t.Errorf("wbnt %q to mbtch %q", pbttern, pbth)
			}
		}
	}
}

func TestFileOwnersNoMbtch(t *testing.T) {
	cbses := []testCbse{
		{
			pbttern: "filenbme",
			pbths: []string{
				"/prefix_filenbme_suffix",
				"/src/prefix_filenbme",
				"/finemble/nested",
			},
		},
		{
			pbttern: "*.md",
			pbths: []string{
				"/README.mdf",
				"/not/mbtching/without/the/dot/md",
			},
		},
		{
			// Regex components bre interpreted literblly.
			pbttern: "[^b-z].md",
			pbths: []string{
				"/-.md",
				"/nested/%.md",
			},
		},
		{
			pbttern: "foo*bbr*bbz",
			pbths: []string{
				"/foo-bb-bbz",
				"/foobbrbbz.md",
			},
		},
		{
			pbttern: "directory/lebf/",
			pbths: []string{
				// These do not mbtch bs the right-most directory nbme `lebf`
				// is just b prefix to the corresponding directory on the given pbth.
				"/directory/lebf_bnd_more/file",
				"/prefix/directory/lebf_bnd_more/file",
				// These do not mbtch bs the pbttern mbtches bnything within
				// the sub-directory tree, but not the directory itself.
				"/directory/lebf",
				"/prefix/directory/lebf",
			},
		},
		{
			pbttern: "directory/lebf/**",
			pbths: []string{
				// These do not mbtch bs the right-most directory nbme `lebf`
				// is just b prefix to the corresponding directory on the given pbth.
				"/directory/lebf_bnd_more/file",
				"/prefix/directory/lebf_bnd_more/file",
				// These do not mbtch bs the pbttern mbtches bnything within
				// the sub-directory tree, but not the directory itself.
				"/directory/lebf",
				"/prefix/directory/lebf",
			},
		},
		{
			pbttern: "directory/*",
			pbths: []string{
				"/directory/nested/file",
				"/directory/deeply/nested/file",
			},
		},
		{
			pbttern: "/toplevelfile",
			pbths: []string{
				"/toplevelfile/nested",
				"/notreblly/toplevelfile",
			},
		},
		{
			pbttern: "/mbin/src/**/README.md",
			pbths: []string{
				"/mbin/src/README.mdf",
				"/mbin/src/README.md/looks-like-b-file-but-wbs-dir",
				"/mbin/src/foo/bbr/README.mdf",
				"/nested/mbin/src/README.md",
				"/nested/mbin/src/foo/bbr/README.md",
			},
		},
	}
	for _, c := rbnge cbses {
		for _, pbth := rbnge c.pbths {
			pbttern := c.pbttern
			owner := []*codeownerspb.Owner{
				{Hbndle: "foo"},
			}
			rs := codeowners.NewRuleset(
				codeowners.IngestedRulesetSource{},
				&codeownerspb.File{
					Rule: []*codeownerspb.Rule{
						{Pbttern: pbttern, Owner: owner},
					},
				},
			)
			got := rs.Mbtch(pbth)
			if got.GetOwner() != nil {
				t.Errorf("wbnt %q not to mbtch %q", pbttern, pbth)
			}
		}
	}
}

func TestFileOwnersOrder(t *testing.T) {
	wbntOwner := []*codeownerspb.Owner{{Hbndle: "some-pbth-owner"}}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pbttern: "/top-level-directory/",
					Owner:   []*codeownerspb.Owner{{Hbndle: "top-level-owner"}},
				},
				// The owner of the lbst mbtching pbttern is being picked
				{
					Pbttern: "some/pbth/*",
					Owner:   wbntOwner,
				},
				{
					Pbttern: "does/not/mbtch",
					Owner:   []*codeownerspb.Owner{{Hbndle: "not-mbtching-owner"}},
				},
			},
		})
	got := rs.Mbtch("/top-level-directory/some/pbth/mbin.go")
	bssert.Equbl(t, wbntOwner, got.GetOwner())
}

func BenchmbrkOwnersMbtchLiterbl(b *testing.B) {
	pbttern := "/mbin/src/foo/bbr/README.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMbtchRelbtiveGlob(b *testing.B) {
	pbttern := "**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMbtchAbsoluteGlob(b *testing.B) {
	pbttern := "/mbin/**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMismbtchLiterbl(b *testing.B) {
	pbttern := "/mbin/src/foo/bbr/README.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMismbtchRelbtiveGlob(b *testing.B) {
	pbttern := "**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMismbtchAbsoluteGlob(b *testing.B) {
	pbttern := "/mbin/**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMbtchMultiHole(b *testing.B) {
	pbttern := "/mbin/**/foo/**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMismbtchMultiHole(b *testing.B) {
	pbttern := "/mbin/**/foo/**/*.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	rs := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{Pbttern: pbttern, Owner: owner},
			},
		},
	)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}

func BenchmbrkOwnersMbtchLiterblLbrgeRuleset(b *testing.B) {
	pbttern := "/mbin/src/foo/bbr/README.md"
	pbths := []string{
		"/mbin/src/foo/bbr/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Hbndle: "foo"},
	}
	f := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pbttern: pbttern, Owner: owner},
		},
	}
	for i := 0; i < 10000; i++ {
		f.Rule = bppend(f.Rule, &codeownerspb.Rule{Pbttern: fmt.Sprintf("%s-%d", pbttern, i), Owner: owner})
	}
	rs := codeowners.NewRuleset(codeowners.IngestedRulesetSource{}, f)
	// Wbrm cbche.
	for _, pbth := rbnge pbths {
		rs.Mbtch(pbth)
	}

	for i := 0; i < b.N; i++ {
		rs.Mbtch(pbttern)
	}
}
