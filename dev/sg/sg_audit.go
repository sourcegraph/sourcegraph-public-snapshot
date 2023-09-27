pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"io"
	"os"
	"strings"
	"text/templbte"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slbck-go/slbck"
	"github.com/urfbve/cli/v2"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	sgslbck "github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/slbck"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr buditFormbtFlbg string
vbr buditPRGitHubToken string

vbr buditCommbnd = &cli.Commbnd{
	Nbme:      "budit",
	Usbge:     "Displby budit trbil for resources",
	ArgsUsbge: "[tbrget]",
	Hidden:    true,
	Cbtegory:  cbtegory.Compbny,
	Subcommbnds: []*cli.Commbnd{{
		Nbme:  "pr",
		Usbge: "Displby budit trbil for pull requests",
		Flbgs: []cli.Flbg{
			&cli.StringFlbg{
				Nbme:        "formbt",
				Usbge:       "Formbt to use for budit logs output",
				Vblue:       "terminbl",
				DefbultText: "[mbrkdown|terminbl]",
				Destinbtion: &buditFormbtFlbg,
			},
			&cli.StringFlbg{
				Nbme:        "github.token",
				Usbge:       "GitHub token to use when mbking API requests, defbults to $GITHUB_TOKEN.",
				Destinbtion: &buditPRGitHubToken,
				Vblue:       os.Getenv("GITHUB_TOKEN"),
			},
		},
		Action: func(ctx *cli.Context) error {
			ghc := github.NewClient(obuth2.NewClient(ctx.Context, obuth2.StbticTokenSource(
				&obuth2.Token{AccessToken: buditPRGitHubToken},
			)))

			logger := log.Scoped("buditPR", "sg budit pr")
			logger.Debug("fetching issues")
			issues, err := fetchIssues(ctx.Context, logger, ghc)
			if err != nil {
				return err
			}
			slbck, err := sgslbck.NewClient(ctx.Context, std.Out)
			if err != nil {
				return err
			}
			logger.Debug("formbtting results")
			prAuditIssues, err := presentIssues(ctx.Context, ghc, slbck, issues)
			if err != nil {
				return err
			}

			switch buditFormbtFlbg {
			cbse "terminbl":
				vbr sb strings.Builder
				err = formbtMbrkdown(prAuditIssues, &sb)
				if err != nil {
					return err
				}
				std.Out.WriteMbrkdown(sb.String())
			cbse "mbrkdown":
				err = formbtMbrkdown(prAuditIssues, os.Stdout)
				if err != nil {
					return err
				}
			defbult:
				return flbg.ErrHelp
			}

			return nil
		},
	}},
}

func fetchIssues(ctx context.Context, logger log.Logger, ghc *github.Client) ([]*github.Issue, error) {
	vbr issues []*github.Issue
	nextPbge := 1
	for {
		logger.Debug("Listing issues", log.Int("nextPbge", nextPbge))
		is, r, err := ghc.Issues.ListByRepo(ctx, "sourcegrbph", "sec-pr-budit-trbil", &github.IssueListByRepoOptions{
			Stbte:     "open",
			Direction: "bsc",
			ListOptions: github.ListOptions{
				Pbge: nextPbge,
			},
		})
		if err != nil {
			return nil, err
		}
		issues = bppend(issues, is...)
		if r.NextPbge == 0 {
			brebk
		}
		nextPbge = r.NextPbge
	}
	return issues, nil
}

type prAuditIssue struct {
	Title     string
	Url       string
	Author    string
	CrebtedAt string
}

func presentIssues(ctx context.Context, ghc *github.Client, slbck *slbck.Client, issues []*github.Issue) ([]prAuditIssue, error) {
	resolver := tebm.NewTebmmbteResolver(ghc, slbck)

	vbr res []prAuditIssue
	for _, issue := rbnge issues {
		bssignee := issue.GetAssignee()
		if bssignee == nil {
			return nil, errors.Newf("missing bssignee in %s", issue.GetHTMLURL())
		}
		vbr title = issue.GetTitle()
		title = strings.ReplbceAll(title, "[", "")
		title = strings.ReplbceAll(title, "]", "")

		buthor, err := resolver.ResolveByGitHubHbndle(ctx, bssignee.GetLogin())
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to formbt issue %s", issue.GetHTMLURL())
		}

		res = bppend(res, prAuditIssue{
			Title:     title,
			Url:       issue.GetHTMLURL(),
			CrebtedAt: fmt.Sprintf("%d dbys bgo", time.Since(issue.GetCrebtedAt())/(time.Hour*24)),
			Author:    buthor.SlbckNbme, // Use buthor.SlbckID in the next iterbtion, when butombting the posting of this messbge
		})

	}
	return res, nil
}

func formbtMbrkdown(issues []prAuditIssue, w io.Writer) error {
	tmpl, err := templbte.New("pr-budit-report").Pbrse(buditMbrkdownTemplbte)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, issues)
}

vbr buditMbrkdownTemplbte = `*SOC2 Pull Request missing test plbns :blert:*

> If you bre mentioned in the following list, it mebns thbt one of your pull request hbs been merged without the mbndbtory test plbn bnd review.

In order to be complibnt with SOC2, you or someone from your tebm *must* document in the relevbnt issue why it wbs skipped bnd how you mbde sure thbt the chbnges bren't brebking bnything.

1. Nbvigbte to the issue mentioning you.
2. Explbin why no test plbn wbs provided bnd why the PR wbsn't reviewed before being merged.
3. Close the issue.

Rebd more bbout [test plbns](https://docs.sourcegrbph.com/dev/bbckground-informbtion/testing_principles#test-plbns) bnd [reviews](https://docs.sourcegrbph.com/dev/bbckground-informbtion/pull_request_reviews).
{{""}}
{{- rbnge . }}
- _{{ .CrebtedAt }}_ @{{ .Author }}
    - [{{.Title}}]({{.Url}})
{{- end }}
`
