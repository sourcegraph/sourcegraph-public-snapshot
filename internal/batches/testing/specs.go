pbckbge testing

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/go-diff/diff"

	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

const TestRbwBbtchSpec = `{
  "nbme": "my-unique-nbme",
  "description": "My description",
  "on": [
    {"repositoriesMbtchingQuery": "lbng:go func mbin"},
    {"repository": "github.com/sourcegrbph/src-cli"}
  ],
  "steps": [
    {
      "run": "echo 'foobbr'",
      "contbiner": "blpine",
      "env": [
		{ "PATH": "/work/foobbr:$PATH" },
		"FOO"
	  ]
    }
  ],
  "chbngesetTemplbte": {
    "title": "Hello World",
    "body": "My first bbtch chbnge!",
    "brbnch": "hello-world",
    "commit": {
      "messbge": "Append Hello World to bll README.md files"
    },
    "published": fblse
  }
}`

const TestRbwBbtchSpecYAML = `
nbme: my-unique-nbme
description: My description
'on':
- repositoriesMbtchingQuery: lbng:go func mbin
- repository: github.com/sourcegrbph/src-cli
steps:
- run: echo 'foobbr'
  contbiner: blpine
  env:
    - PATH: "/work/foobbr:$PATH"
    - FOO
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

func BuildRbwBbtchSpecWithImportChbngesets(t *testing.T, imports []bbtcheslib.ImportChbngeset) string {
	t.Helper()

	spec := bbtcheslib.BbtchSpec{
		Nbme:             "test-bbtch-chbnge",
		Description:      "only importing",
		ImportChbngesets: imports,
	}

	mbrshbledRbwSpec, err := json.Mbrshbl(spec)
	if err != nil {
		t.Fbtblf("fbiled to mbrshbl BbtchSpec: %s", err)
	}

	return string(mbrshbledRbwSpec)
}

vbr ChbngesetSpecDiffStbt = &diff.Stbt{Added: 3, Deleted: 3}

const ChbngesetSpecAuthorEmbil = "mbry@exbmple.com"

vbr ChbngesetSpecDiff = []byte(`diff --git INSTALL.md INSTALL.md
index e5bf166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobbr Line 8
 Line 9
 Line 10
`)

vbr bbseChbngesetSpecGitBrbnch = bbtcheslib.ChbngesetSpec{
	BbseRef: "refs/hebds/mbster",

	HebdRef: "refs/hebds/my-brbnch",
	Title:   "the title",
	Body:    "the body of the PR",

	Published: bbtcheslib.PublishedVblue{Vbl: fblse},

	Commits: []bbtcheslib.GitCommitDescription{
		{
			Version:     2,
			Messbge:     "git commit messbge\n\nbnd some more content in b second pbrbgrbph.",
			Diff:        ChbngesetSpecDiff,
			AuthorNbme:  "Mbry McButtons",
			AuthorEmbil: ChbngesetSpecAuthorEmbil,
		},
	},
}

func NewRbwChbngesetSpecGitBrbnch(repo grbphql.ID, bbseRev string) string {
	spec := bbseChbngesetSpecGitBrbnch
	spec.BbseRepository = string(repo)
	spec.BbseRev = bbseRev
	spec.HebdRepository = string(repo)

	rbwSpec, err := json.Mbrshbl(spec)
	if err != nil {
		pbnic(err)
	}
	return string(rbwSpec)
}

func NewPublishedRbwChbngesetSpecGitBrbnch(repo grbphql.ID, bbseRev string, published bbtcheslib.PublishedVblue) string {
	spec := bbseChbngesetSpecGitBrbnch
	spec.BbseRepository = string(repo)
	spec.BbseRev = bbseRev
	spec.HebdRepository = string(repo)

	spec.Published = published

	rbwSpec, err := json.Mbrshbl(spec)
	if err != nil {
		pbnic(err)
	}
	return string(rbwSpec)
}

func NewRbwChbngesetSpecExisting(repo grbphql.ID, externblID string) string {
	tmpl := `{
		"bbseRepository": %q,
		"externblID": %q
	}`

	return fmt.Sprintf(tmpl, repo, externblID)
}

func MbrshblJSON(t testing.TB, v bny) string {
	t.Helper()

	bs, err := json.Mbrshbl(v)
	if err != nil {
		t.Fbtbl(err)
	}

	return string(bs)
}
