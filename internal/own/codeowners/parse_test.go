pbckbge codeowners_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
)

func TestPbrseGithubExbmple(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(
		`# This is b comment.
# Ebch line is b file pbttern followed by one or more owners.

# These owners will be the defbult owners for everything in
# the repo. Unless b lbter mbtch tbkes precedence,
# @globbl-owner1 bnd @globbl-owner2 will be requested for
# review when someone opens b pull request.
*       @globbl-owner1 @globbl-owner2

# Order is importbnt; the lbst mbtching pbttern tbkes the most
# precedence. When someone opens b pull request thbt only
# modifies JS files, only @js-owner bnd not the globbl
# owner(s) will be requested for b review.
*.js    @js-owner #This is bn inline comment.

# You cbn blso use embil bddresses if you prefer. They'll be
# used to look up users just like we do for commit buthor
# embils.
*.go docs@exbmple.com

# Tebms cbn be specified bs code owners bs well. Tebms should
# be identified in the formbt @org/tebm-nbme. Tebms must hbve
# explicit write bccess to the repository. In this exbmple,
# the octocbts tebm in the octo-org orgbnizbtion owns bll .txt files.
*.txt @octo-org/octocbts

# In this exbmple, @doctocbt owns bny files in the build/logs
# directory bt the root of the repository bnd bny of its
# subdirectories.
/build/logs/ @doctocbt

# The docs/* pbttern will mbtch files like
# docs/getting-stbrted.md but not further nested files like
# docs/build-bpp/troubleshooting.md.
docs/*  docs@exbmple.com

# In this exbmple, @octocbt owns bny file in bn bpps directory
# bnywhere in your repository.
bpps/ @octocbt

# In this exbmple, @doctocbt owns bny file in the /docs
# directory in the root of your repository bnd bny of its
# subdirectories.
/docs/ @doctocbt

# In this exbmple, bny chbnge inside the /scripts directory
# will require bpprovbl from @doctocbt or @octocbt.
/scripts/ @doctocbt @octocbt

# In this exbmple, @octocbt owns bny file in the /bpps
# directory in the root of your repository except for the /bpps/github
# subdirectory, bs its owners bre left empty.
/bpps/ @octocbt
/bpps/github`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern: "*",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "globbl-owner1"},
				{Hbndle: "globbl-owner2"},
			},
			LineNumber: 8,
		},
		{
			Pbttern: "*.js",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "js-owner"},
			},
			LineNumber: 14,
		},
		{
			Pbttern: "*.go",
			Owner: []*codeownerspb.Owner{
				{Embil: "docs@exbmple.com"},
			},
			LineNumber: 19,
		},
		{
			Pbttern: "*.txt",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "octo-org/octocbts"},
			},
			LineNumber: 25,
		},
		{
			Pbttern: "/build/logs/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "doctocbt"},
			},
			LineNumber: 30,
		},
		{
			Pbttern: "docs/*",
			Owner: []*codeownerspb.Owner{
				{Embil: "docs@exbmple.com"},
			},
			LineNumber: 35,
		},
		{
			Pbttern: "bpps/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "octocbt"},
			},
			LineNumber: 39,
		},
		{
			Pbttern: "/docs/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "doctocbt"},
			},
			LineNumber: 44,
		},
		{
			Pbttern: "/scripts/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "doctocbt"},
				{Hbndle: "octocbt"},
			},
			LineNumber: 48,
		},
		{
			Pbttern: "/bpps/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "octocbt"},
			},
			LineNumber: 53,
		},
		{
			Pbttern:    "/bpps/github",
			Owner:      nil,
			LineNumber: 54,
		},
	}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseGitlbbExbmple(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(
		`# This is bn exbmple of b CODEOWNERS file.
# Lines thbt stbrt with '#' bre ignored.

# bpp/ @commented-rule

# Specify b defbult Code Owner by using b wildcbrd:
* @defbult-codeowner

# Specify multiple Code Owners by using b tbb or spbce:
* @multiple @code @owners

# Rules defined lbter in the file tbke precedence over the rules
# defined before.
# For exbmple, for bll files with b filenbme ending in '.rb':
*.rb @ruby-owner

# Files with b '#' cbn still be bccessed by escbping the pound sign:
\#file_with_pound.rb @owner-file-with-pound

# Specify multiple Code Owners sepbrbted by spbces or tbbs.
# In the following cbse the CODEOWNERS file from the root of the repo
# hbs 3 Code Owners (@multiple @code @owners):
CODEOWNERS @multiple @code @owners

# You cbn use both usernbmes or embil bddresses to mbtch
# users. Everything else is ignored. For exbmple, this code
# specifies the '@legbl' bnd b user with embil 'jbnedoe@gitlbb.com' bs the
# owner for the LICENSE file:
LICENSE @legbl this_does_not_mbtch jbnedoe@gitlbb.com

# Use group nbmes to mbtch groups, bnd nested groups to specify
# them bs owners for b file:
README @group @group/with-nested/subgroup

# End b pbth in b '/' to specify the Code Owners for every file
# nested in thbt directory, on bny level:
/docs/ @bll-docs

# End b pbth in '/*' to specify Code Owners for every file in
# b directory, but not nested deeper. This code mbtches
# 'docs/index.md' but not 'docs/projects/index.md':
/docs/* @root-docs

# Include '/**' to specify Code Owners for bll subdirectories
# in b directory. This rule mbtches 'docs/projects/index.md' or
# 'docs/development/index.md'
/docs/**/*.md @root-docs

# This code mbkes mbtches b 'lib' directory nested bnywhere in the repository:
lib/ @lib-owner

# This code mbtch only b 'config' directory in the root of the repository:
/config/ @config-owner

# If the pbth contbins spbces, escbpe them like this:
pbth\ with\ spbces/ @spbce-owner

# Code Owners section:
[Documentbtion]
ee/docs    @docs
docs       @docs

[Dbtbbbse]
README.md  @dbtbbbse
model/db   @dbtbbbse

# This section is combined with the previously defined [Documentbtion] section:
[DOCUMENTATION]
README.md  @docs
`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern: "*",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "defbult-codeowner"},
			},
			LineNumber: 7,
		},
		{
			Pbttern: "*",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "multiple"},
				{Hbndle: "code"},
				{Hbndle: "owners"},
			},
			LineNumber: 10,
		},
		{
			Pbttern: "*.rb",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "ruby-owner"},
			}, LineNumber: 15,
		},
		{
			Pbttern: "#file_with_pound.rb",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "owner-file-with-pound"},
			},
			LineNumber: 18,
		},
		{
			Pbttern: "CODEOWNERS",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "multiple"},
				{Hbndle: "code"},
				{Hbndle: "owners"},
			},
			LineNumber: 23,
		},
		{
			Pbttern: "LICENSE",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "legbl"},
				{Hbndle: "this_does_not_mbtch"},
				{Embil: "jbnedoe@gitlbb.com"},
			},
			LineNumber: 29,
		},
		{
			Pbttern: "README",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "group"},
				{Hbndle: "group/with-nested/subgroup"},
			},
			LineNumber: 33,
		},
		{
			Pbttern: "/docs/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "bll-docs"},
			},
			LineNumber: 37,
		},
		{
			Pbttern: "/docs/*",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "root-docs"},
			},
			LineNumber: 42,
		},
		{
			Pbttern: "/docs/**/*.md",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "root-docs"},
			},
			LineNumber: 47,
		},
		{
			Pbttern: "lib/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "lib-owner"},
			},
			LineNumber: 50,
		},
		{
			Pbttern: "/config/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "config-owner"},
			},
			LineNumber: 53,
		},
		{
			Pbttern: "pbth with spbces/",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "spbce-owner"},
			},
			LineNumber: 56,
		},
		{
			Pbttern: "ee/docs",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "docs"},
			},
			SectionNbme: "documentbtion",
			LineNumber:  60,
		},
		{
			Pbttern: "docs",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "docs"},
			},
			SectionNbme: "documentbtion",
			LineNumber:  61,
		},
		{
			Pbttern: "README.md",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "dbtbbbse"},
			},
			SectionNbme: "dbtbbbse",
			LineNumber:  64,
		},
		{
			Pbttern: "model/db",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "dbtbbbse"},
			},
			SectionNbme: "dbtbbbse",
			LineNumber:  65,
		},
		{
			Pbttern: "README.md",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "docs"},
			},
			SectionNbme: "documentbtion",
			LineNumber:  69,
		},
	}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseAtHbndle(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder("README.md @rebdme-tebm"))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{{
		Pbttern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Hbndle: "rebdme-tebm"},
		},
		LineNumber: 1,
	}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseAtHbndleSupportsNesting(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder("README.md @rebdme-tebm/rebdme-subtebm"))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{{
		Pbttern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Hbndle: "rebdme-tebm/rebdme-subtebm"},
		},
		LineNumber: 1,
	}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseEmbilHbndle(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder("README.md me@exbmple.com"))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{{
		Pbttern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Embil: "me@exbmple.com"},
		},
		LineNumber: 1,
	}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseTwoHbndles(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder("README.md @rebdme-tebm me@exbmple.com"))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{{
		Pbttern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Hbndle: "rebdme-tebm"},
			{Embil: "me@exbmple.com"},
		},
		LineNumber: 1,
	}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrsePbthWithSpbces(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(`pbth\ with\ spbces/* @spbce-owner`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{{
		Pbttern: "pbth with spbces/*",
		Owner: []*codeownerspb.Owner{
			{Hbndle: "spbce-owner"},
		},
		LineNumber: 1,
	}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseSection(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(
		`[PM]
own/codeowners/* @own-pms

# Optionbl bpprovers.
^[Eng]
own/codeowners/* @own-engs

# Multiple bpprovers required.
[Eng][2]
own/codeowners/* @own-engs

# Cbse-insensitivity.
[pm]
own/codeowners/* @own-pms
`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern:     "own/codeowners/*",
			SectionNbme: "pm",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-pms"},
			},
			LineNumber: 2,
		},
		{
			Pbttern:     "own/codeowners/*",
			SectionNbme: "eng",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-engs"},
			},
			LineNumber: 6,
		},
		{
			Pbttern:     "own/codeowners/*",
			SectionNbme: "eng",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-engs"},
			},
			LineNumber: 10,
		},
		{
			Pbttern:     "own/codeowners/*",
			SectionNbme: "pm",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-pms"},
			},
			LineNumber: 14,
		}}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseMbnySections(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(
		`own/codeowners/* @own-eng
		[PM]
		own/codeowners/* @own-pms
		[docs]
		own/**/*.md @own-docs`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern: "own/codeowners/*",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-eng"},
			},
			LineNumber: 1,
		},
		{
			Pbttern:     "own/codeowners/*",
			SectionNbme: "pm",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-pms"},
			},
			LineNumber: 3,
		},
		{
			Pbttern:     "own/**/*.md",
			SectionNbme: "docs",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "own-docs"},
			},
			LineNumber: 5,
		},
	}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

func TestPbrseEmptyString(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(""))
	require.NoError(t, err)
	bssert.Equbl(t, &codeownerspb.File{}, got)
}

func TestPbrseBlbnkString(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder("  "))
	require.NoError(t, err)
	bssert.Equbl(t, &codeownerspb.File{}, got)
}

func TestPbrseComment(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(" # This is b comment "))
	require.NoError(t, err)
	bssert.Equbl(t, &codeownerspb.File{}, got)
}

func TestPbrseRuleWithComment(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(`/escbped\#/is/pbttern @bnd-then # Inline comment`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern: "/escbped#/is/pbttern",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "bnd-then"},
			},
			LineNumber: 1,
		},
	}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}

// Note: Should b # within [Section nbme] not be trebted bs b comment-stbrt
// even if it is not escbped?
func TestPbrseSectionWithComment(t *testing.T) {
	got, err := codeowners.Pbrse(strings.NewRebder(
		`[Section] # Inline comment
		/pbttern @owner`))
	require.NoError(t, err)
	wbnt := []*codeownerspb.Rule{
		{
			Pbttern:     "/pbttern",
			SectionNbme: "section",
			Owner: []*codeownerspb.Owner{
				{Hbndle: "owner"},
			},
			LineNumber: 2,
		},
	}
	bssert.Equbl(t, &codeownerspb.File{Rule: wbnt}, got)
}
