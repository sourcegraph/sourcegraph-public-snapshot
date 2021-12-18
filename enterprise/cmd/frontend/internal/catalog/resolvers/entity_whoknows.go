package resolvers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *componentResolver) WhoKnows(ctx context.Context, args *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	authors, err := r.Authors(ctx)
	if err != nil {
		return nil, err
	}

	codeOwners, err := r.CodeOwners(ctx)
	if err != nil {
		return nil, err
	}

	usage, err := r.Usage(ctx, &gql.ComponentUsageArgs{})
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

	if authors != nil {
		for _, author := range *authors {
			email := author.Person().Email()
			e := byEmail[email]
			if e == nil {
				e = &whoKnowsEdgeResolver{person: author.Person(), db: r.db}
				byEmail[email] = e
			}
			e.authorEdge = author.(*componentAuthorEdgeResolver)
		}
	}

	if codeOwners != nil {
		for _, codeOwner := range *codeOwners {
			email := codeOwner.Node().Email()
			e := byEmail[email]
			if e == nil {
				e = &whoKnowsEdgeResolver{person: codeOwner.Node(), db: r.db}
				byEmail[email] = e
			}
			e.codeOwnerEdge = codeOwner.(*componentCodeOwnerEdgeResolver)
		}
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

	// Score
	for _, edge := range byEmail {
		type reasonScore struct {
			reason string
			score  float64
		}
		var reasonScores []reasonScore

		if edge.authorEdge != nil {
			lastActivity := edge.authorEdge.data.LastCommitDate
			if daysAgo := time.Since(lastActivity) / (24 * time.Hour); daysAgo < 45 {
				reasonScores = append(reasonScores, reasonScore{
					reason: fmt.Sprintf("Contributed %s", humanize.Time(lastActivity)),
					score:  7 / float64(daysAgo),
				})
			}

			authorCount := edge.authorEdge.data.LineCount
			authorProportion := edge.authorEdge.AuthoredLineProportion()
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
			daysAgo := time.Since(lastActivity) / (24 * time.Hour)
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
	for _, edge := range byEmail {
		edge.score = edge.score / max
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

	authorEdge    *componentAuthorEdgeResolver
	codeOwnerEdge *componentCodeOwnerEdgeResolver
	callerEdge    *componentUsedByPersonEdgeResolver

	db database.DB
}

func (r *whoKnowsEdgeResolver) Node() *gql.PersonResolver { return r.person }
func (r *whoKnowsEdgeResolver) Reasons() []string         { return r.reasons }
func (r *whoKnowsEdgeResolver) Score() float64            { return r.score }
