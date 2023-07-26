package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

// computeCodeowners evaluates the codeowners file (if any) against given file (blob)
// and returns resolvers for identified owners.
func (r *ownResolver) computeCodeowners(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver) ([]reasonAndReference, error) {
	repo := blob.Repository()
	repoID, repoName := repo.IDInt32(), repo.RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	// Find ruleset which represents CODEOWNERS file at given revision.
	ruleset, err := r.ownService().RulesetForRepo(ctx, repoName, repoID, commitID)
	if err != nil {
		return nil, err
	}
	var rule *codeownerspb.Rule
	if ruleset != nil {
		rule = ruleset.Match(blob.Path())
	}
	// Compute repo context if possible to allow better unification of references.
	var repoContext *own.RepoContext
	if len(rule.GetOwner()) > 0 {
		spec, err := repo.ExternalRepo(ctx)
		// Best effort resolution. We still want to serve the reason if external service cannot be resolved here.
		if err == nil {
			repoContext = &own.RepoContext{
				Name:         repoName,
				CodeHostKind: spec.ServiceType,
			}
		}
	}
	// Return references
	var rrs []reasonAndReference
	for _, o := range rule.GetOwner() {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				codeownersRule:   rule,
				codeownersSource: ruleset.GetSource(),
			},
			reference: own.Reference{
				RepoContext: repoContext,
				Handle:      o.Handle,
				Email:       o.Email,
			},
		})
	}
	return rrs, nil
}

type codeownersFileEntryResolver struct {
	db              database.DB
	source          codeowners.RulesetSource
	matchLineNumber int32
	repo            *graphqlbackend.RepositoryResolver
	gitserverClient gitserver.Client
}

func (r *codeownersFileEntryResolver) Title() (string, error) {
	return "codeowners", nil
}

func (r *codeownersFileEntryResolver) Description() (string, error) {
	return "Owner is associated with a rule in a CODEOWNERS file.", nil
}

func (r *codeownersFileEntryResolver) CodeownersFile(ctx context.Context) (graphqlbackend.FileResolver, error) {
	switch src := r.source.(type) {
	case codeowners.IngestedRulesetSource:
		// For ingested, create a virtual file resolver that loads the raw contents
		// on demand.
		stat := graphqlbackend.CreateFileInfo("CODEOWNERS", false)
		return graphqlbackend.NewVirtualFileResolver(stat, func(ctx context.Context) (string, error) {
			f, err := r.db.Codeowners().GetCodeownersForRepo(ctx, api.RepoID(src.ID))
			if err != nil {
				return "", err
			}
			return f.Contents, nil
		}, graphqlbackend.VirtualFileResolverOptions{
			URL: fmt.Sprintf("%s/-/own/edit", r.repo.URL()),
		}), nil
	case codeowners.GitRulesetSource:
		// For committed, we can return a GitTreeEntry, as it implements File2.
		c := graphqlbackend.NewGitCommitResolver(r.db, r.gitserverClient, r.repo, src.Commit, nil)
		return c.File(ctx, &struct{ Path string }{Path: src.Path})
	default:
		return nil, errors.New("unknown ownership file source")
	}
}

func (r *codeownersFileEntryResolver) RuleLineMatch(_ context.Context) (int32, error) {
	return r.matchLineNumber, nil
}
