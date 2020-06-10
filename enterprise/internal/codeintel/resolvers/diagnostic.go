package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
)

type diagnosticResolver struct {
	repo               *types.Repo
	commit             api.CommitID
	diagnostic         codeintelapi.ResolvedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

var _ graphqlbackend.DiagnosticResolver = &diagnosticResolver{}

func (r *diagnosticResolver) Location(ctx context.Context) (graphqlbackend.LocationResolver, error) {
	clientRange := client.Range{
		Start: client.Position{Line: r.diagnostic.Diagnostic.StartLine, Character: r.diagnostic.Diagnostic.EndLine},
		End:   client.Position{Line: r.diagnostic.Diagnostic.StartCharacter, Character: r.diagnostic.Diagnostic.EndCharacter},
	}

	adjustedCommit, adjustedRange, err := adjustLocation(ctx, r.diagnostic.Dump.RepositoryID, r.diagnostic.Dump.Commit, r.diagnostic.Diagnostic.Path, clientRange, r.repo, r.commit)
	if err != nil {
		return nil, err
	}

	treeResolver, err := r.collectionResolver.resolve(ctx, api.RepoID(r.repo.ID), string(adjustedCommit), r.diagnostic.Diagnostic.Path)
	if err != nil {
		return nil, err
	}

	if treeResolver == nil {
		return nil, nil
	}

	return graphqlbackend.NewLocationResolver(treeResolver, &adjustedRange), nil
}

var severities = map[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func (r *diagnosticResolver) Severity(ctx context.Context) (*string, error) {
	if severity, ok := severities[r.diagnostic.Diagnostic.Severity]; ok {
		return &severity, nil
	}

	return nil, fmt.Errorf("unknown diagnostic severity %d", r.diagnostic.Diagnostic.Severity)
}

func (r *diagnosticResolver) Code(ctx context.Context) (*string, error) {
	if r.diagnostic.Diagnostic.Code == "" {
		return nil, nil
	}

	return &r.diagnostic.Diagnostic.Code, nil
}

func (r *diagnosticResolver) Source(ctx context.Context) (*string, error) {
	if r.diagnostic.Diagnostic.Source == "" {
		return nil, nil
	}

	return &r.diagnostic.Diagnostic.Source, nil
}

func (r *diagnosticResolver) Message(ctx context.Context) (*string, error) {
	if r.diagnostic.Diagnostic.Message == "" {
		return nil, nil
	}

	return &r.diagnostic.Diagnostic.Message, nil
}
