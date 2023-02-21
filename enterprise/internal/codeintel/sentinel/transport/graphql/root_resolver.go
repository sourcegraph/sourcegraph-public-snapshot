package graphql

import (
	"context"
	"strconv"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc *sentinel.Service
	operations  *operations
}

func NewRootResolver(observationCtx *observation.Context, sentinelSvc *sentinel.Service) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc: sentinelSvc,
		operations:  newOperations(observationCtx),
	}
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args resolverstubs.GetVulnerabilitiesArgs) (_ resolverstubs.VulnerabilityConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getVulnerabilities.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit := 50
	if args.First != nil {
		limit = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}

		offset = after
	}

	vulnerabilities, totalCount, err := r.sentinelSvc.GetVulnerabilities(ctx, shared.GetVulnerabilitiesArgs{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return &vulnerabilityConnectionResolver{
		vulnerabilities: vulnerabilities,
		offset:          offset,
		totalCount:      totalCount,
	}, nil
}

func (r *rootResolver) VulnerabilityMatches(ctx context.Context, args resolverstubs.GetVulnerabilityMatchesArgs) (_ resolverstubs.VulnerabilityMatchConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getMatches.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit := 50
	if args.First != nil {
		limit = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}

		offset = after
	}

	matches, totalCount, err := r.sentinelSvc.GetVulnerabilityMatches(ctx, shared.GetVulnerabilityMatchesArgs{
		Limit:  limit,
		Offset: offset,
	})

	return &vulnerabilityMatchConnectionResolver{
		matches:    matches,
		offset:     offset,
		totalCount: totalCount,
	}, nil
}

//
//

type vulnerabilityResolver struct {
	v shared.Vulnerability
}

func (r *vulnerabilityResolver) SourceID() string   { return r.v.SourceID }
func (r *vulnerabilityResolver) Summary() string    { return r.v.Summary }
func (r *vulnerabilityResolver) Details() string    { return r.v.Details }
func (r *vulnerabilityResolver) CPEs() []string     { return r.v.CPEs }
func (r *vulnerabilityResolver) CWEs() []string     { return r.v.CWEs }
func (r *vulnerabilityResolver) Aliases() []string  { return r.v.Aliases }
func (r *vulnerabilityResolver) Related() []string  { return r.v.Related }
func (r *vulnerabilityResolver) DataSource() string { return r.v.DataSource }
func (r *vulnerabilityResolver) URLs() []string     { return r.v.URLs }
func (r *vulnerabilityResolver) Severity() string   { return r.v.Severity }
func (r *vulnerabilityResolver) CVSSVector() string { return r.v.CVSSVector }
func (r *vulnerabilityResolver) CVSSScore() string  { return r.v.CVSSScore }

func (r *vulnerabilityResolver) Published() gqlutil.DateTime {
	return *gqlutil.DateTimeOrNil(&r.v.Published)
}

func (r *vulnerabilityResolver) Modified() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.Modified)
}

func (r *vulnerabilityResolver) Withdrawn() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.Withdrawn)
}

//
//

type vulnerabilityConnectionResolver struct {
	vulnerabilities []shared.Vulnerability
	offset          int
	totalCount      int
}

func (r *vulnerabilityConnectionResolver) Nodes() []resolverstubs.VulnerabilityResolver {
	var resolvers []resolverstubs.VulnerabilityResolver
	for _, v := range r.vulnerabilities {
		resolvers = append(resolvers, &vulnerabilityResolver{v: v})
	}

	return resolvers
}

func (r *vulnerabilityConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *vulnerabilityConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.offset+len(r.vulnerabilities) < r.totalCount {
		return sharedresolvers.NextPageCursor(strconv.Itoa(r.offset + len(r.vulnerabilities)))
	}

	return sharedresolvers.HasNextPage(false)
}

//
//

type vulnerabilityMatchResolver struct {
	m shared.VulnerabilityMatch
}

func (r *vulnerabilityMatchResolver) SourceID() string {
	return r.m.SourceID
}

//
//

type vulnerabilityMatchConnectionResolver struct {
	matches    []shared.VulnerabilityMatch
	offset     int
	totalCount int
}

func (r *vulnerabilityMatchConnectionResolver) Nodes() []resolverstubs.VulnerabilityMatchResolver {
	var resolvers []resolverstubs.VulnerabilityMatchResolver
	for _, m := range r.matches {
		resolvers = append(resolvers, &vulnerabilityMatchResolver{m: m})
	}

	return resolvers
}

func (r *vulnerabilityMatchConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *vulnerabilityMatchConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.offset+len(r.matches) < r.totalCount {
		return sharedresolvers.NextPageCursor(strconv.Itoa(r.offset + len(r.matches)))
	}

	return sharedresolvers.HasNextPage(false)
}
