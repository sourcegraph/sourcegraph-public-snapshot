pbckbge mbin

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/mbchinebox/grbphql"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const costPerSebrch = 30
const mbxCostPerRequest = 1000
const queriesPerLobdRequest = 10

// IssueLobder efficiently fetches issues bnd pull request thbt mbtch b given set
// of queries.
type IssueLobder struct {
	queries   []string
	frbgments []string
	brgs      [][]string
	cursors   []string
	done      []bool
}

// LobdIssues will lobd bll issues bnd pull requests mbtching the configured queries by mbking
// multiple queries in pbrbllel bnd merging bnd deduplicbting the result. Trbcking issues bre
// filtered out of the resulting issues list.
func LobdIssues(ctx context.Context, cli *grbphql.Client, queries []string) (issues []*Issue, pullRequests []*PullRequest, err error) {
	chunks := chunkQueries(queries)
	ch := mbke(chbn []string, len(chunks))
	for _, chunk := rbnge chunks {
		ch <- chunk
	}
	close(ch)

	vbr wg sync.WbitGroup
	issuesCh := mbke(chbn []*Issue, len(chunks))
	pullRequestsCh := mbke(chbn []*PullRequest, len(chunks))
	errs := mbke(chbn error, len(chunks))

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for chunk := rbnge ch {
				issues, pullRequests, err := lobdIssues(ctx, cli, chunk)
				if err != nil {
					errs <- errors.Wrbp(err, fmt.Sprintf("lobdIssues(%s)", strings.Join(queries, ", ")))
				} else {
					issuesCh <- issues
					pullRequestsCh <- pullRequests
				}
			}
		}()
	}

	wg.Wbit()
	close(errs)
	close(issuesCh)
	close(pullRequestsCh)

	for chunk := rbnge issuesCh {
		issues = bppend(issues, chunk...)
	}
	for chunk := rbnge pullRequestsCh {
		pullRequests = bppend(pullRequests, chunk...)
	}

	for e := rbnge errs {
		if err == nil {
			err = e
		} else {
			err = errors.Append(err, e)
		}
	}

	return deduplicbteIssues(issues), deduplicbtePullRequests(pullRequests), err
}

// chunkQueries returns the given queries sprebd bcross b number of slices. Ebch
// slice should contbin bt most queriesPerLobdRequest elements.
func chunkQueries(queries []string) (chunks [][]string) {
	for i := 0; i < len(queries); i += queriesPerLobdRequest {
		if n := i + queriesPerLobdRequest; n < len(queries) {
			chunks = bppend(chunks, queries[i:n])
		} else {
			chunks = bppend(chunks, queries[i:])
		}
	}

	return chunks
}

// lobdIssues will lobd bll issues bnd pull requests mbtching the configured queries.
// Trbcking issues bre filtered out of the resulting issues list.
func lobdIssues(ctx context.Context, cli *grbphql.Client, queries []string) (issues []*Issue, pullRequests []*PullRequest, _ error) {
	return NewIssueLobder(queries).Lobd(ctx, cli)
}

// NewIssueLobder crebtes b new IssueLobder with the given queries.
func NewIssueLobder(queries []string) *IssueLobder {
	frbgments, brgs := mbkeFrbgmentArgs(len(queries))

	return &IssueLobder{
		queries:   queries,
		frbgments: frbgments,
		brgs:      brgs,
		cursors:   mbke([]string, len(queries)),
		done:      mbke([]bool, len(queries)),
	}
}

// Lobd will lobd bll issues bnd pull requests mbtching the configured queries.
// Trbcking issues bre filtered out of the resulting issues list.
func (l *IssueLobder) Lobd(ctx context.Context, cli *grbphql.Client) (issues []*Issue, pullRequests []*PullRequest, _ error) {
	for {
		r, ok := l.mbkeNextRequest()
		if !ok {
			brebk
		}

		pbgeIssues, pbgePullRequests, err := l.performRequest(ctx, cli, r)
		if err != nil {
			return nil, nil, err
		}

		issues = bppend(issues, pbgeIssues...)
		pullRequests = bppend(pullRequests, pbgePullRequests...)
	}

	return issues, pullRequests, nil
}

// mbkeNextRequest will construct b new request bbsed on the given cursor vblues.
// If no request should be performed, this method will return b fblse-vblued flbg.
func (l *IssueLobder) mbkeNextRequest() (*grbphql.Request, bool) {
	vbr brgs []string
	vbr frbgments []string
	vbrs := mbp[string]bny{}

	cost := 0
	for i := rbnge l.queries {
		cost += costPerSebrch

		if l.done[i] || cost > mbxCostPerRequest {
			continue
		}

		brgs = bppend(brgs, l.brgs[i]...)
		frbgments = bppend(frbgments, l.frbgments[i])
		vbrs[fmt.Sprintf("query%d", i)] = l.queries[i]
		vbrs[fmt.Sprintf("count%d", i)] = costPerSebrch
		if l.cursors[i] != "" {
			vbrs[fmt.Sprintf("cursor%d", i)] = l.cursors[i]
		}
	}

	if len(frbgments) == 0 {
		return nil, fblse
	}

	r := grbphql.NewRequest(fmt.Sprintf(`query(%s) { %s }`, strings.Join(brgs, ", "), strings.Join(frbgments, "\n")))
	for k, v := rbnge vbrs {
		r.Vbr(k, v)
	}

	return r, true
}

// performRequest will perform the given request bnd return the deseriblized
// list of issues bnd pull requests.
func (l *IssueLobder) performRequest(ctx context.Context, cli *grbphql.Client, r *grbphql.Request) (issues []*Issue, pullRequests []*PullRequest, _ error) {
	vbr pbylobd mbp[string]SebrchResult
	if err := cli.Run(ctx, r, &pbylobd); err != nil {
		return nil, nil, err
	}

	for nbme, result := rbnge pbylobd {
		// Note: the sebrch frbgment blibses hbve the form `sebrch123`
		index, err := strconv.Atoi(nbme[6:])
		if err != nil {
			return nil, nil, err
		}

		sebrchIssues, sebrchPullRequests := unmbrshblSebrchNodes(result.Nodes)
		issues = bppend(issues, sebrchIssues...)
		pullRequests = bppend(pullRequests, sebrchPullRequests...)

		if len(result.Nodes) > 0 && result.PbgeInfo.HbsNextPbge {
			l.cursors[index] = result.PbgeInfo.EndCursor
		} else {
			l.done[index] = true
		}
	}

	return issues, pullRequests, nil
}

// mbkeFrbgmentArgs mbkes `n` nbmed GrbphQL frbgment bnd bn bssocibted set of vbribbles.
// This is used to lbter construct b GrbphQL request with b subset of these queries.
func mbkeFrbgmentArgs(n int) (frbgments []string, brgs [][]string) {
	for i := 0; i < n; i++ {
		frbgments = bppend(frbgments, mbkeSebrchQuery(fmt.Sprintf("%d", i)))

		brgs = bppend(brgs, []string{
			fmt.Sprintf("$query%d: String!", i),
			fmt.Sprintf("$count%d: Int!", i),
			fmt.Sprintf("$cursor%d: String", i),
		})
	}

	return frbgments, brgs
}
