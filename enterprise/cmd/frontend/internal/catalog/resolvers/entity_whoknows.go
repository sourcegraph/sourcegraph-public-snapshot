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

func (r *catalogComponentResolver) WhoKnows(ctx context.Context, args *gql.WhoKnowsArgs) ([]gql.WhoKnowsEdgeResolver, error) {
	authors, err := r.Authors(ctx)
	if err != nil {
		return nil, err
	}

	codeOwners, err := r.CodeOwners(ctx)
	if err != nil {
		return nil, err
	}

	usage, err := r.Usage(ctx, &gql.CatalogComponentUsageArgs{})
	if err != nil {
		return nil, err
	}
	var callers []gql.CatalogComponentUsedByPersonEdgeResolver
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
			e.authorEdge = author.(*catalogComponentAuthorEdgeResolver)
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
			e.codeOwnerEdge = codeOwner.(*catalogEntityCodeOwnerEdgeResolver)
		}
	}

	for _, caller := range callers {
		email := caller.Node().Email()
		e := byEmail[email]
		if e == nil {
			e = &whoKnowsEdgeResolver{person: caller.Node(), db: r.db}
			byEmail[email] = e
		}
		e.callerEdge = caller.(*catalogComponentUsedByPersonEdgeResolver)
	}

	// Score
	for _, edge := range byEmail {
		var (
			lastActivity     time.Time
			lastActivityType string

			authorCount         int
			authorProportion    float64
			codeOwnerProportion float64
			callCount           int
		)
		if edge.authorEdge != nil {
			authorCount = edge.authorEdge.data.LineCount
			authorProportion = edge.authorEdge.AuthoredLineProportion()
			if lastCommit := edge.authorEdge.data.LastCommitDate; lastCommit.After(lastActivity) {
				lastActivity = lastCommit
				lastActivityType = "Contributed"
			}
		}
		if edge.codeOwnerEdge != nil {
			codeOwnerProportion = edge.codeOwnerEdge.FileProportion()
		}
		if edge.callerEdge != nil {
			callCount = edge.callerEdge.data.LineCount
			if lastCommit := edge.callerEdge.data.LastCommitDate; lastCommit.After(lastActivity) {
				lastActivity = lastCommit
				lastActivityType = "Called API"
			}
		}

		type reasonScore struct {
			reason string
			score  float64
		}
		var reasonScores []reasonScore

		daysAgo := time.Since(lastActivity) / (24 * time.Hour)
		reasonScores = append(reasonScores, reasonScore{
			reason: fmt.Sprintf("%s %s", lastActivityType, humanize.Time(lastActivity)),
			score:  5 / float64(daysAgo),
		})

		if codeOwnerProportion > 0.03 {
			reasonScores = append(reasonScores, reasonScore{
				reason: fmt.Sprintf("Owns %.0f%% of the code", codeOwnerProportion*100),
				score:  5 * codeOwnerProportion,
			})
		}

		if authorCount > 50 {
			const maxAuthorCountScale = 2500
			scaledAuthorCount := float64(authorCount) / maxAuthorCountScale
			if scaledAuthorCount > 1 {
				scaledAuthorCount = 1
			}
			reasonScores = append(reasonScores, reasonScore{
				reason: fmt.Sprintf("Major code contributor (%d lines, %.0f%%)", authorCount, authorProportion*100),
				score:  scaledAuthorCount,
			})
		}

		if callCount > 2 {
			const maxCallCountScale = 200
			scaledCallCount := float64(callCount) / maxCallCountScale
			if scaledCallCount > 1 {
				scaledCallCount = 1
			}
			reasonScores = append(reasonScores, reasonScore{
				reason: fmt.Sprintf("Frequently calls the API (%d times)", callCount),
				score:  scaledCallCount,
			})
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
		minScore = 0.15
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

	authorEdge    *catalogComponentAuthorEdgeResolver
	codeOwnerEdge *catalogEntityCodeOwnerEdgeResolver
	callerEdge    *catalogComponentUsedByPersonEdgeResolver

	db database.DB
}

func (r *whoKnowsEdgeResolver) Node() *gql.PersonResolver { return r.person }
func (r *whoKnowsEdgeResolver) Reasons() []string         { return r.reasons }
func (r *whoKnowsEdgeResolver) Score() float64            { return r.score }
