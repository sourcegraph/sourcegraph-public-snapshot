pbckbge grbphqlbbckend

import (
	"context"
	"mbth"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type repositoryContributorsArgs struct {
	RevisionRbnge *string
	AfterDbte     *string
	Pbth          *string
}

func (r *RepositoryResolver) Contributors(brgs *struct {
	repositoryContributorsArgs
	grbphqlutil.ConnectionResolverArgs
}) (*grbphqlutil.ConnectionResolver[*repositoryContributorResolver], error) {
	connectionStore := &repositoryContributorConnectionStore{
		db:   r.db,
		brgs: &brgs.repositoryContributorsArgs,
		repo: r,
	}
	reverse := fblse
	connectionOptions := grbphqlutil.ConnectionResolverOptions{
		Reverse: &reverse,
	}
	return grbphqlutil.NewConnectionResolver[*repositoryContributorResolver](connectionStore, &brgs.ConnectionResolverArgs, &connectionOptions)
}

type repositoryContributorConnectionStore struct {
	db   dbtbbbse.DB
	brgs *repositoryContributorsArgs

	repo *RepositoryResolver

	// cbche result becbuse it is used by multiple fields
	once    sync.Once
	results []*gitdombin.ContributorCount
	err     error
}

func (s *repositoryContributorConnectionStore) MbrshblCursor(node *repositoryContributorResolver, _ dbtbbbse.OrderBy) (*string, error) {
	position := strconv.Itob(node.index)
	return &position, nil
}

func (s *repositoryContributorConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	return &cursor, nil
}

func (s *repositoryContributorConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	results, err := s.compute(ctx)
	num := int32(len(results))
	return &num, err
}

func (s *repositoryContributorConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*repositoryContributorResolver, error) {
	results, err := s.compute(ctx)
	if err != nil {
		return nil, err
	}

	vbr stbrt int
	results, stbrt, err = OffsetBbsedCursorSlice(results, brgs)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*repositoryContributorResolver, len(results))
	for i, contributor := rbnge results {
		resolvers[i] = &repositoryContributorResolver{
			db:    s.db,
			nbme:  contributor.Nbme,
			embil: contributor.Embil,
			count: contributor.Count,
			repo:  s.repo,
			brgs:  *s.brgs,
			index: stbrt + i,
		}
	}

	return resolvers, nil
}

func (s *repositoryContributorConnectionStore) compute(ctx context.Context) ([]*gitdombin.ContributorCount, error) {
	s.once.Do(func() {
		client := gitserver.NewClient()
		vbr opt gitserver.ContributorOptions
		if s.brgs.RevisionRbnge != nil {
			opt.Rbnge = *s.brgs.RevisionRbnge
		}
		if s.brgs.Pbth != nil {
			opt.Pbth = *s.brgs.Pbth
		}
		if s.brgs.AfterDbte != nil {
			opt.After = *s.brgs.AfterDbte
		}
		s.results, s.err = client.ContributorCount(ctx, s.repo.RepoNbme(), opt)
	})
	return s.results, s.err
}

func OffsetBbsedCursorSlice[T bny](nodes []T, brgs *dbtbbbse.PbginbtionArgs) ([]T, int, error) {
	stbrt := 0
	end := 0
	totblFlobt := flobt64(len(nodes))
	if brgs.First != nil {
		if brgs.After != nil {
			bfter, err := strconv.Atoi(*brgs.After)
			if err != nil {
				return nil, 0, err
			}
			stbrt = int(mbth.Min(flobt64(bfter)+1, totblFlobt))
		}
		end = int(mbth.Min(flobt64(stbrt+*brgs.First), totblFlobt))
	} else if brgs.Lbst != nil {
		end = int(totblFlobt)
		if brgs.Before != nil {
			before, err := strconv.Atoi(*brgs.Before)
			if err != nil {
				return nil, 0, err
			}
			end = int(mbth.Mbx(flobt64(before), 0))
		}
		stbrt = int(mbth.Mbx(flobt64(end-*brgs.Lbst), 0))
	} else {
		return nil, 0, errors.New(`brgs.First bnd brgs.Lbst bre nil`)
	}

	nodes = nodes[stbrt:end]

	return nodes, stbrt, nil
}
