package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
)

type fileResolver struct {
	commit commitSpec
	name   string
	path   string
}

func (r *fileResolver) Name() string {
	return r.name
}

func (r *fileResolver) Content(ctx context.Context) (string, error) {
	repoupdater.Enqueue(r.commit.RepoID, auth.ActorFromContext(ctx).UserSpec())

	file, err := backend.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{
				Repo:     r.commit.RepoID,
				CommitID: r.commit.CommitID,
			},
			Path: r.path,
		},
	})
	if err != nil {
		return "", err
	}
	return string(file.Contents), nil
}

func (r *fileResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}

	hunks, err := vcsrepo.BlameFile(ctx, r.path, &vcs.BlameOptions{
		NewestCommit: vcs.CommitID(r.commit.CommitID),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return nil, err
	}

	var hunksResolver []*hunkResolver
	for _, hunk := range hunks {
		hunksResolver = append(hunksResolver, &hunkResolver{
			hunk: hunk,
		})
	}

	return hunksResolver, nil
}

func (r *fileResolver) Definition(ctx context.Context,
	args *struct {
		Line     int32
		Column   int32
		Language string
	}) (*definitionResolver, error) {

	refs, err := backend.Defs.RefLocations(ctx, sourcegraph.RefLocationsOptions{
		RepoID:    r.commit.RepoID,
		Language:  args.Language,
		File:      r.name,
		Line:      int(args.Line),
		Character: int(args.Column),
	})
	if err != nil {
		return nil, err
	}

	var globalRefs []*globalReferencesResolver
	for _, ref := range refs.Locations {
		globalRefs = append(globalRefs, &globalReferencesResolver{
			refLocation: &refLocationResolver{
				startLineNumber: int32(ref.StartLine),
				startColumn:     int32(ref.StartChar),
				endLineNumber:   int32(ref.EndLine),
				endColumn:       int32(ref.EndChar),
			},
			uri: &uriResolver{
				host:     ref.Host,
				fragment: ref.File,
				path:     ref.Path,
				query:    ref.Version,
				scheme:   ref.Scheme,
			},
		})
	}

	return &definitionResolver{globalReferences: globalRefs}, nil
}
