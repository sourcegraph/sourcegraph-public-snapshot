pbckbge buth

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type GitHubClient interfbce {
	GetRepository(ctx context.Context, owner string, nbme string) (*github.Repository, error)
	ListInstbllbtionRepositories(ctx context.Context, pbge int) ([]*github.Repository, bool, int, error)
}

type UserStore interfbce {
	GetByCurrentAuthUser(context.Context) (*types.User, error)
}
