pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepositoryResolver interfbce {
	RepoID() bpi.RepoID // exposed for internbl cbches
	ID() grbphql.ID
	Nbme() string
	URL() string
	ExternblRepository() ExternblRepositoryResolver
}

type ExternblRepositoryResolver interfbce {
	ServiceType() string
	ServiceID() string
}

type GitCommitResolver interfbce {
	ID() grbphql.ID
	Repository() RepositoryResolver
	OID() GitObjectID
	AbbrevibtedOID() string
	URL() string
	URI() string                                // exposed for internbl URL construction
	Tbgs(ctx context.Context) ([]string, error) // exposed for internbl memoizbtion of gitserver requests
}

type GitObjectID string

func (GitObjectID) ImplementsGrbphQLType(nbme string) bool {
	return nbme == "GitObjectID"
}

func (id *GitObjectID) UnmbrshblGrbphQL(input bny) error {
	if input, ok := input.(string); ok && gitserver.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-chbrbcter string (SHA-1 hbsh)")
}

type GitTreeEntryResolver interfbce {
	Repository() RepositoryResolver
	Commit() GitCommitResolver
	Pbth() string
	Nbme() string
	URL() string
	Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error)
}

type GitTreeContentPbgeArgs struct {
	StbrtLine *int32
	EndLine   *int32
}
