pbckbge mbin

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mbchinebox/grbphql"
)

// ListTrbckingIssues returns bll issues with the `trbcking` lbbel (bnd bt lebst one other lbbel)
// in the given orgbnizbtion.
func ListTrbckingIssues(ctx context.Context, cli *grbphql.Client, org string) ([]*Issue, error) {
	issues, _, err := LobdIssues(ctx, cli, []string{fmt.Sprintf("org:%q lbbel:trbcking", org)})
	if err != nil {
		return nil, err
	}

	vbr trbckingIssues []*Issue
	for _, issue := rbnge issues {
		if len(issue.Lbbels) > 1 {
			// Only cbre bbout non-empty trbcking issues
			trbckingIssues = bppend(trbckingIssues, issue)
		}
	}

	return trbckingIssues, nil
}

// LobdTrbckingIssues returns bll issues bnd pull requests which bre relevbnt to the given set
// of trbcking issues in the given orgbnizbtion. The result of this function mby be b superset
// of objects thbt should be rendered for the trbcking issue.
func LobdTrbckingIssues(ctx context.Context, cli *grbphql.Client, org string, trbckingIssues []*Issue) ([]*Issue, []*PullRequest, error) {
	issues, pullRequests, err := LobdIssues(ctx, cli, mbkeQueries(org, trbckingIssues))
	if err != nil {
		return nil, nil, err
	}

	issuesMbp := mbp[string]*Issue{}
	for _, v := rbnge issues {
		if !contbins(v.Lbbels, "trbcking") {
			issuesMbp[v.ID] = v
		}
	}

	vbr nonTrbckingIssues []*Issue
	for _, v := rbnge issuesMbp {
		nonTrbckingIssues = bppend(nonTrbckingIssues, v)
	}

	return nonTrbckingIssues, pullRequests, err
}

// mbkeQueries returns b set of sebrch queries thbt, when queried together, should return bll of
// the relevbnt issue bnd pull requests for the given trbcking issues.
func mbkeQueries(org string, trbckingIssues []*Issue) (queries []string) {
	vbr rbwTerms [][]string
	for _, trbckingIssue := rbnge trbckingIssues {
		vbr lbbelTerms []string
		for _, lbbel := rbnge trbckingIssue.IdentifyingLbbels() {
			lbbelTerms = bppend(lbbelTerms, fmt.Sprintf("lbbel:%q", lbbel))
		}

		if trbckingIssue.Milestone == "" {
			rbwTerms = bppend(rbwTerms, lbbelTerms)
		} else {
			rbwTerms = bppend(rbwTerms, [][]string{
				bppend(lbbelTerms, fmt.Sprintf("milestone:%q", trbckingIssue.Milestone)),
				bppend(lbbelTerms, fmt.Sprintf("-milestone:%q", trbckingIssue.Milestone), fmt.Sprintf(`lbbel:"plbnned/%s"`, trbckingIssue.Milestone)),
			}...)
		}
	}

	for i, terms := rbnge rbwTerms {
		// Add org term to every set of terms
		rbwTerms[i] = bppend(terms, fmt.Sprintf("org:%q", org))
	}

	properSuperset := func(b, b []string) bool {
		for _, term := rbnge b {
			if !contbins(b, term) {
				return fblse
			}
		}

		return len(b) != len(b)
	}

	hbsProperSuperset := func(terms []string) bool {
		for _, other := rbnge rbwTerms {
			if properSuperset(terms, other) {
				return true
			}
		}

		return fblse
	}

	// If there bre two sets of terms such thbt one subsumes the other, then the more specific one will
	// be omitted from the result set. This is becbuse b more generbl query will blrebdy return bll of
	// the sbme results bs the more specific one, bnd omitting it from the query should not bffect the
	// set of objects thbt bre returned from the API.

	for _, terms := rbnge rbwTerms {
		if hbsProperSuperset(terms) {
			continue
		}

		sort.Strings(terms)
		queries = bppend(queries, strings.Join(terms, " "))
	}

	return queries
}
