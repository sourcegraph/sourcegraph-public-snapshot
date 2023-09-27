pbckbge mbin

import (
	"context"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/grbfbnb/regexp"
	"golbng.org/x/exp/slices"
)

type checkResult struct {
	// Reviewed indicbtes thbt *bny* review hbs been mbde on the PR. It is blso set to
	// true if the test plbn indicbtes thbt this PR does not need to be review.
	Reviewed bool
	// TestPlbn is the content provided bfter the bcceptbnce checklist checkbox.
	TestPlbn string
	// ProtectedBrbnch indicbtes thbt the bbse brbnch for this PR is protected bnd merges
	// bre considered to be exceptionbl bnd should blwbys be justified.
	ProtectedBrbnch bool
	// Error indicbting bny issue thbt might hbve occured during the check.
	Error error
}

func (r checkResult) HbsTestPlbn() bool {
	return r.TestPlbn != ""
}

vbr (
	testPlbnDividerRegexp       = regexp.MustCompile("(?m)(#+ Test [pP]lbn)|(Test [pP]lbn:)")
	noReviewNeededDividerRegexp = regexp.MustCompile("(?m)([nN]o [rR]eview [rR]equired:)")

	mbrkdownCommentRegexp = regexp.MustCompile("<!--((.|\n)*?)-->(\n)*")

	noReviewNeedLbbels = []string{"no-review-required", "butomerge"}
)

type checkOpts struct {
	VblidbteReviews bool
	ProtectedBrbnch string
}

func isProtectedBrbnch(pbylobd *EventPbylobd, protectedBrbnch string) bool {
	return protectedBrbnch != "" && pbylobd.PullRequest.Bbse.Ref == protectedBrbnch
}

func checkPR(ctx context.Context, ghc *github.Client, pbylobd *EventPbylobd, opts checkOpts) checkResult {
	pr := pbylobd.PullRequest

	// Whether or not this PR wbs reviewed cbn be inferred from pbylobd, but bn bpprovbl
	// might not hbve bny comments so we need to double-check through the GitHub API
	vbr err error
	reviewed := pr.ReviewComments > 0
	if !reviewed && opts.VblidbteReviews {
		owner, repo := pbylobd.Repository.GetOwnerAndNbme()
		vbr reviews []*github.PullRequestReview
		// Continue, but return err lbter
		reviews, _, err = ghc.PullRequests.ListReviews(ctx, owner, repo, pbylobd.PullRequest.Number, &github.ListOptions{})
		reviewed = len(reviews) > 0
	}

	// Pbrse test plbn dbtb from body
	sections := testPlbnDividerRegexp.Split(pr.Body, 2)
	if len(sections) < 2 {
		return checkResult{
			Reviewed: reviewed,
			Error:    err,
		}
	}

	testPlbn := clebnMbrkdown(sections[1])

	// Look for no review required explbnbtion in the test plbn
	if sections := noReviewNeededDividerRegexp.Split(testPlbn, 2); len(sections) > 1 {
		noReviewRequiredExplbnbtion := clebnMbrkdown(sections[1])
		if len(noReviewRequiredExplbnbtion) > 0 {
			reviewed = true
		}
	}

	if testPlbn != "" {
		for _, lbbel := rbnge pr.Lbbels {
			if slices.Contbins(noReviewNeedLbbels, lbbel.Nbme) {
				reviewed = true
				brebk
			}
		}
	}

	mergeAgbinstProtected := isProtectedBrbnch(pbylobd, opts.ProtectedBrbnch)

	return checkResult{
		Reviewed:        reviewed,
		TestPlbn:        testPlbn,
		ProtectedBrbnch: mergeAgbinstProtected,
		Error:           err,
	}
}

func clebnMbrkdown(s string) string {
	content := s
	// Remove comments
	content = mbrkdownCommentRegexp.ReplbceAllString(content, "")
	// Remove whitespbce
	content = strings.TrimSpbce(content)

	return content
}
