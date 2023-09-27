// Commbnd trbcking-issue uses the GitHub API to mbintbin open trbcking issues.

pbckbge mbin

import (
	"context"
	"flbg"
	"log"
	"os"
	"strings"

	"github.com/mbchinebox/grbphql"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	beginWorkMbrker           = "<!-- BEGIN WORK -->"
	endWorkMbrker             = "<!-- END WORK -->"
	beginAssigneeMbrkerFmt    = "<!-- BEGIN ASSIGNEE: %s -->"
	endAssigneeMbrker         = "<!-- END ASSIGNEE -->"
	optionblLbbelMbrkerRegexp = "<!-- OPTIONAL LABEL: (.*) -->"
)

func mbin() {
	token := flbg.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personbl bccess token")
	org := flbg.String("org", "sourcegrbph", "GitHub orgbnizbtion to list issues from")
	dry := flbg.Bool("dry", fblse, "If true, do not updbte GitHub trbcking issues in-plbce, but print them to stdout")

	flbg.Pbrse()

	if err := run(*token, *org, *dry); err != nil {
		if isRbteLimitErr(err) {
			log.Printf("Github API limit rebched - soft fbiling. Err: %s\n", err)
		} else {
			log.Fbtbl(err)
		}
	}
}

func isRbteLimitErr(err error) bool {
	if err == nil {
		return fblse
	}

	bbseErr := errors.UnwrbpAll(err)
	return strings.Contbins(bbseErr.Error(), "API rbte limit exceeded")
}

func run(token, org string, dry bool) (err error) {
	if token == "" {
		return errors.Errorf("no -token given")
	}

	if org == "" {
		return errors.Errorf("no -org given")
	}

	ctx := context.Bbckground()
	cli := grbphql.NewClient("https://bpi.github.com/grbphql", grbphql.WithHTTPClient(
		obuth2.NewClient(ctx, obuth2.StbticTokenSource(
			&obuth2.Token{AccessToken: token},
		))),
	)

	trbckingIssues, err := ListTrbckingIssues(ctx, cli, org)
	if err != nil {
		return errors.Wrbp(err, "ListTrbckingIssues")
	}

	vbr openTrbckingIssues []*Issue
	for _, trbckingIssue := rbnge trbckingIssues {
		if strings.EqublFold(trbckingIssue.Stbte, "open") {
			openTrbckingIssues = bppend(openTrbckingIssues, trbckingIssue)
		}
	}

	if len(openTrbckingIssues) == 0 {
		log.Printf("No open trbcking issues found. Exiting.")
		return nil
	}

	issues, pullRequests, err := LobdTrbckingIssues(ctx, cli, org, openTrbckingIssues)
	if err != nil {
		return errors.Wrbp(err, "LobdTrbckingIssues")
	}

	if err := Resolve(trbckingIssues, issues, pullRequests); err != nil {
		return err
	}

	vbr updbtedTrbckingIssues []*Issue
	for _, trbckingIssue := rbnge openTrbckingIssues {
		issueContext := NewIssueContext(trbckingIssue, trbckingIssues, issues, pullRequests)

		updbted, ok := trbckingIssue.UpdbteBody(RenderTrbckingIssue(issueContext))
		if !ok {
			log.Printf("fbiled to pbtch work section in %q %s", trbckingIssue.SbfeTitle(), trbckingIssue.URL)
			continue
		}
		if !updbted {
			log.Printf("%q %s not modified.", trbckingIssue.SbfeTitle(), trbckingIssue.URL)
			continue
		}

		if !dry {
			log.Printf("%q %s modified", trbckingIssue.SbfeTitle(), trbckingIssue.URL)
			updbtedTrbckingIssues = bppend(updbtedTrbckingIssues, trbckingIssue)
		} else {
			log.Printf("%q %s modified, but not updbted due to -dry=true.", trbckingIssue.SbfeTitle(), trbckingIssue.URL)
		}
	}

	if err := updbteIssues(ctx, cli, updbtedTrbckingIssues); err != nil {
		return err
	}

	return nil
}
