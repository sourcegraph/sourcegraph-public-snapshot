package graphql

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadResolver struct {
	db               *dbstore.Store
	gitserver        policies.GitserverClient
	resolver         resolvers.Resolver
	upload           dbstore.Upload
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadResolver(db *dbstore.Store, gitserver policies.GitserverClient, resolver resolvers.Resolver, upload dbstore.Upload, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, traceErrs *observation.ErrCollector) gql.LSIFUploadResolver {
	if upload.AssociatedIndexID != nil {
		// Request the next batch of index fetches to contain the record's associated
		// index id, if one exists it exists. This allows the prefetcher.GetIndexByID
		// invocation in the AssociatedIndex method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkIndex(*upload.AssociatedIndexID)
	}

	return &UploadResolver{
		db:               db,
		gitserver:        gitserver,
		resolver:         resolver,
		upload:           upload,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        traceErrs,
	}
}

func (r *UploadResolver) ID() graphql.ID            { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *UploadResolver) InputCommit() string       { return r.upload.Commit }
func (r *UploadResolver) InputRoot() string         { return r.upload.Root }
func (r *UploadResolver) IsLatestForRepo() bool     { return r.upload.VisibleAtTip }
func (r *UploadResolver) UploadedAt() gql.DateTime  { return gql.DateTime{Time: r.upload.UploadedAt} }
func (r *UploadResolver) Failure() *string          { return r.upload.FailureMessage }
func (r *UploadResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.upload.StartedAt) }
func (r *UploadResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.upload.FinishedAt) }
func (r *UploadResolver) InputIndexer() string      { return r.upload.Indexer }
func (r *UploadResolver) PlaceInQueue() *int32      { return toInt32(r.upload.Rank) }

func (r *UploadResolver) State() string {
	state := strings.ToUpper(r.upload.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *UploadResolver) AssociatedIndex(ctx context.Context) (_ gql.LSIFIndexResolver, err error) {
	// TODO - why are a bunch of them zero?
	if r.upload.AssociatedIndexID == nil || *r.upload.AssociatedIndexID == 0 {
		return nil, nil
	}

	defer r.traceErrs.Collect(&err,
		log.String("uploadResolver.field", "associatedIndex"),
		log.Int("associatedIndex", *r.upload.AssociatedIndexID),
	)

	index, exists, err := r.prefetcher.GetIndexByID(ctx, *r.upload.AssociatedIndexID)
	if err != nil || !exists {
		return nil, err
	}

	return NewIndexResolver(r.db, r.gitserver, r.resolver, index, r.prefetcher, r.locationResolver, r.traceErrs), nil
}

func (r *UploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

func (r *UploadResolver) RetentionPolicyOverview(ctx context.Context, args *gql.LSIFUploadRetentionPolicyMatchesArgs) (_ gql.CodeIntelligenceRetentionPolicyMatchesConnectionResolver, err error) {
	var afterID int64
	if args.After != nil {
		afterID, err = unmarshalConfigurationPolicyGQLID(graphql.ID(*args.After))
		if err != nil {
			return nil, err
		}
	}

	pageSize := DefaultRetentionPolicyMatchesPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	policyMatcher := policies.NewMatcher(r.gitserver, policies.RetentionExtractor, false, false)

	var term string
	if args.Query != nil {
		term = *args.Query
	}

	policies, _, err := r.resolver.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
		RepositoryID:     r.upload.RepositoryID,
		Term:             term,
		ForDataRetention: true,
		Limit:            pageSize,
		Offset:           int(afterID),
	})
	if err != nil {
		return nil, err
	}

	visibileCommits, err := r.commitsVisibleToUpload(ctx)
	if err != nil {
		return nil, err
	}

	matchingPolicies, err := policyMatcher.CommitsDescribedByPolicy(ctx, r.upload.RepositoryID, policies, time.Now(), visibileCommits...)
	if err != nil {
		return nil, err
	}

	var (
		now                    = time.Now()
		potentialMatchIndexSet map[int]int // map of polciy ID to array index
		potentialMatches       []retentionPolicyMatchCandidate
	)

	if args.MatchesOnly {
		potentialMatches, _ = r.populateMatchingCommits(visibileCommits, matchingPolicies, policies, now)
	} else {
		potentialMatches, potentialMatchIndexSet = r.populateMatchingCommits(visibileCommits, matchingPolicies, policies, now)

		// populate with remaining unmatched policies
		for _, policy := range policies {
			if _, ok := potentialMatchIndexSet[policy.ID]; !ok {
				potentialMatches = append(potentialMatches, retentionPolicyMatchCandidate{
					ConfigurationPolicy: policy,
					matched:             false,
				})
			}
		}
	}

	sort.Slice(potentialMatches, func(i, j int) bool {
		return potentialMatches[i].ID < potentialMatches[j].ID
	})

	return NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(database.NewDBWith(r.db), r.resolver, potentialMatches, len(potentialMatches), r.traceErrs), nil
}

func (r *UploadResolver) commitsVisibleToUpload(ctx context.Context) (commits []string, err error) {
	var token *string
	for first := true; first || token != nil; first = false {
		cs, nextToken, err := r.db.CommitsVisibleToUpload(ctx, r.upload.ID, 50, token)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.CommitsVisibleToUpload")
		}
		token = nextToken

		commits = append(commits, cs...)
	}

	return
}

func (r *UploadResolver) populateMatchingCommits(visibileCommits []string, matchingPolicies map[string][]policies.PolicyMatch, policies []dbstore.ConfigurationPolicy, now time.Time) ([]retentionPolicyMatchCandidate, map[int]int) {
	var (
		potentialMatches       = make([]retentionPolicyMatchCandidate, 0, len(policies))
		potentialMatchIndexSet = make(map[int]int, len(policies))
	)

	// First add all matches for the commit of this upload. We do this to ensure that if a policy matches both the upload's commit
	// and a visible commit, we ensure an entry for that policy is only added for the upload's commit. This makes the logic in checking
	// the visible commits a bit simpler, as we don't have to check if policy X has already been added for a visible commit in the case
	// that the upload's commit is not first in the list.
	if policyMatches, ok := matchingPolicies[r.upload.Commit]; ok {
		for _, policyMatch := range policyMatches {
			potentialMatches = append(potentialMatches, retentionPolicyMatchCandidate{
				ConfigurationPolicy: *policyByID(policies, *policyMatch.PolicyID),
				matched:             true,
			})
			potentialMatchIndexSet[*policyMatch.PolicyID] = len(potentialMatches) - 1
		}
	}

	for _, commit := range visibileCommits {
		if commit == r.upload.Commit {
			continue
		}
		if policyMatches, ok := matchingPolicies[commit]; ok {
			for _, policyMatch := range policyMatches {
				if policyMatch.PolicyDuration == nil || now.Sub(r.upload.UploadedAt) < *policyMatch.PolicyDuration {
					if index, ok := potentialMatchIndexSet[*policyMatch.PolicyID]; ok && potentialMatches[index].protectingCommits != nil {
						potentialMatches[index].protectingCommits = append(potentialMatches[index].protectingCommits, commit)
					} else {
						potentialMatches = append(potentialMatches, retentionPolicyMatchCandidate{
							ConfigurationPolicy: *policyByID(policies, *policyMatch.PolicyID),
							matched:             true,
							protectingCommits:   []string{commit},
						})
						potentialMatchIndexSet[*policyMatch.PolicyID] = len(potentialMatches) - 1
					}
				}
			}
		}
	}

	return potentialMatches, potentialMatchIndexSet
}

func policyByID(policies []dbstore.ConfigurationPolicy, id int) *dbstore.ConfigurationPolicy {
	for _, policy := range policies {
		if policy.ID == id {
			return &policy
		}
	}
	return nil
}
