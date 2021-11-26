package resolvers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *componentResolver) WhoKnows(ctx context.Context, args *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	slocs, err := r.sourceLocationSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.WhoKnows(ctx, args)
}

func (r *rootResolver) GitTreeEntryWhoKnows(ctx context.Context, treeEntry *gql.GitTreeEntryResolver, args *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	return sourceLocationSetResolverFromTreeEntry(treeEntry, r.db).WhoKnows(ctx, args)
}

func (r *sourceLocationSetResolver) WhoKnows(ctx context.Context, args *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	contributorConnection, err := r.Contributors(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	var contributors []gql.ComponentAuthorEdgeResolver
	if contributorConnection != nil {
		contributors = contributorConnection.Edges()
	}

	codeOwnerConnection, err := r.CodeOwners(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	var codeOwners []gql.ComponentCodeOwnerEdgeResolver
	if codeOwnerConnection != nil {
		codeOwners = codeOwnerConnection.Edges()
	}

	usage, err := r.Usage(ctx)
	if err != nil {
		return nil, err
	}
	var callers []gql.ComponentUsedByPersonEdgeResolver
	if usage != nil {
		callers, err = usage.People(ctx)
		if err != nil {
			return nil, err
		}
	}

	byEmail := map[string]*whoKnowsEdgeResolver{}

	for _, contributor := range contributors {
		email := contributor.Person().Email()
		e := byEmail[email]
		if e == nil {
			e = &whoKnowsEdgeResolver{person: contributor.Person(), db: r.db}
			byEmail[email] = e
		}
		e.contributorEdge = contributor.(*componentAuthorEdgeResolver)
	}

	for _, codeOwner := range codeOwners {
		email := codeOwner.Node().Email()
		e := byEmail[email]
		if e == nil {
			e = &whoKnowsEdgeResolver{person: codeOwner.Node(), db: r.db}
			byEmail[email] = e
		}
		e.codeOwnerEdge = codeOwner.(*componentCodeOwnerEdgeResolver)
	}

	maxCalls := 0
	for _, caller := range callers {
		email := caller.Node().Email()
		e := byEmail[email]
		if e == nil {
			e = &whoKnowsEdgeResolver{person: caller.Node(), db: r.db}
			byEmail[email] = e
		}
		e.callerEdge = caller.(*componentUsedByPersonEdgeResolver)

		if e.callerEdge.data.LineCount > maxCalls {
			maxCalls = e.callerEdge.data.LineCount
		}
	}

	// Delete hi@sourcegraph.com dummy owner.
	// TODO(sqs): Come up with a general way to omit these dummy entries.
	delete(byEmail, "hi@sourcegraph.com")
	delete(byEmail, "bot@renovateapp.com")
	delete(byEmail, "29139614+renovate[bot]@users.noreply.github.com")
	delete(byEmail, "renovate[bot]@users.noreply.github.com")

	// Score
	for _, edge := range byEmail {
		type reasonScore struct {
			reason string
			score  float64
		}
		var reasonScores []reasonScore

		if edge.contributorEdge != nil {
			lastActivity := edge.contributorEdge.data.LastCommitDate
			if daysAgo := 1 + time.Since(lastActivity)/(24*time.Hour); daysAgo < 45 {
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Contributed %s", humanize.Time(lastActivity)),
					score:  7 / float64(daysAgo),
				})
			}

			authorCount := edge.contributorEdge.data.LineCount
			authorProportion := edge.contributorEdge.AuthoredLineProportion()
			if authorCount > 50 || authorProportion > 0.05 {
				const maxAuthorCountScale = 2500
				scaledAuthorCount := float64(authorCount) / maxAuthorCountScale
				if scaledAuthorCount > 1 {
					scaledAuthorCount = 1
				}
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Major code contributor (%d lines, %.0f%%)", authorCount, authorProportion*100),
					score:  5*scaledAuthorCount + 5*authorProportion,
				})
			}

		}

		if edge.codeOwnerEdge != nil {
			if codeOwnerProportion := edge.codeOwnerEdge.FileProportion(); codeOwnerProportion > 0.03 {
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Owns %.0f%% of the code", codeOwnerProportion*100),
					score:  5 * codeOwnerProportion,
				})
			}

		}

		if edge.callerEdge != nil {
			lastActivity := edge.callerEdge.data.LastCommitDate
			daysAgo := 1 + time.Since(lastActivity)/(24*time.Hour)
			recentlyCalled := daysAgo < 45
			if recentlyCalled {
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Called API %s", humanize.Time(lastActivity)),
					score:  0.5 / float64(daysAgo),
				})
			}

			if callCount := edge.callerEdge.data.LineCount; callCount > 5 {
				scaledCallCount := float64(callCount) / float64(maxCalls)
				if recentlyCalled {
					scaledCallCount *= 3
				}
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Frequently calls the API (%d times)", callCount),
					score:  scaledCallCount,
				})
			}

		}

		sort.Slice(reasonScores, func(i, j int) bool { return reasonScores[i].score > reasonScores[j].score })
		for _, v := range reasonScores {
			edge.score += v.score
			if v.score > 0.1 {
				edge.reasons = append(edge.reasons, v.reason)
			}
		}
	}

	// Scale to (0,1]
	var max float64
	for _, edge := range byEmail {
		if edge.score > max {
			max = edge.score
		}
	}
	if max != 0 {
		for _, edge := range byEmail {
			edge.score = edge.score / max
		}
	}

	edges := make([]gql.WhoKnowsEdgeResolver, 0, len(byEmail))
	for _, edge := range byEmail {
		edges = append(edges, edge)
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].Score() > edges[j].Score() })

	// Only take top.
	const (
		minScore = 0.1
		maxEdges = 10
	)
	for i, edge := range edges {
		if edge.Score() < minScore || i > maxEdges {
			edges = edges[:i]
			break
		}
	}

	return edges, nil
}

type whoKnowsEdgeResolver struct {
	person  *gql.PersonResolver
	reasons []string
	score   float64

	contributorEdge *componentAuthorEdgeResolver
	codeOwnerEdge   *componentCodeOwnerEdgeResolver
	callerEdge      *componentUsedByPersonEdgeResolver

	db database.DB
}

func (r *whoKnowsEdgeResolver) Node() *gql.PersonResolver { return r.person }
func (r *whoKnowsEdgeResolver) Reasons() []string         { return r.reasons }
func (r *whoKnowsEdgeResolver) Score() float64            { return r.score }
