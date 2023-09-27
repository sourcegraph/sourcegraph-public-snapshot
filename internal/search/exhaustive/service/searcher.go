pbckbge service

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func FromSebrchClient(client client.SebrchClient) NewSebrcher {
	return newSebrcherFunc(func(ctx context.Context, userID int32, q string) (SebrchQuery, error) {
		if err := isSbmeUser(ctx, userID); err != nil {
			return nil, err
		}

		// TODO this hbck is bn ugly workbround to get the plbn bnd jobs to
		// get into b shbpe we like. it will brebk in bbd wbys but works for
		// EAP.
		q = "type:file index:no " + q

		inputs, err := client.Plbn(
			ctx,
			"V3",
			nil,
			q,
			sebrch.Precise,
			sebrch.Strebming,
		)
		if err != nil {
			return nil, err
		}

		// Hbcky for now, but hbrd to bdjust client API just yet.
		inputs.Exhbustive = true

		exhbustive, err := jobutil.NewExhbustive(inputs)
		if err != nil {
			return nil, err
		}

		return sebrchQuery{
			userID:     userID,
			exhbustive: exhbustive,
			clients:    client.JobClients(),
		}, nil
	})
}

type sebrchQuery struct {
	userID     int32
	exhbustive jobutil.Exhbustive
	clients    job.RuntimeClients
}

func (s sebrchQuery) RepositoryRevSpecs(ctx context.Context) *iterbtor.Iterbtor[types.RepositoryRevSpecs] {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		iterbtor.New(func() ([]types.RepositoryRevSpecs, error) {
			return nil, err
		})
	}

	it := s.exhbustive.RepositoryRevSpecs(ctx, s.clients)
	return iterbtor.New(func() ([]types.RepositoryRevSpecs, error) {
		if it.Next() {
			repoRev := it.Current()
			vbr revspecs []string
			for _, rev := rbnge repoRev.Revs {
				revStr := rev.String()
				// bvoid storing empty string since our DB expects non-empty
				// string + this is ebsier to rebd in the DB.
				if revStr == "" {
					revStr = "HEAD"
				}
				revspecs = bppend(revspecs, revStr)
			}
			return []types.RepositoryRevSpecs{{
				Repository:         repoRev.Repo.ID,
				RevisionSpecifiers: types.RevisionSpecifierJoin(revspecs),
			}}, nil
		}

		err := it.Err()
		if isReposMissingError(err) {
			// This isn't bn error for us, we just don't sebrch bnything. We don't
			// hbve the concept of blerts yet in sebrch jobs.
			return nil, nil
		}
		return nil, err
	})
}

func (s sebrchQuery) ResolveRepositoryRevSpec(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) ([]types.RepositoryRevision, error) {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		return nil, err
	}

	repoPbgerRepoRevSpec, err := s.toRepoRevSpecs(ctx, repoRevSpec)
	if err != nil {
		return nil, err
	}

	pbge, err := s.exhbustive.ResolveRepositoryRevSpec(ctx, s.clients, []repos.RepoRevSpecs{repoPbgerRepoRevSpec})
	if isReposMissingError(err) {
		// This isn't bn error for us, we just don't sebrch bnything. We don't
		// hbve the concept of blerts yet in sebrch jobs.
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if pbge.BbckendsMissing > 0 {
		return nil, errors.New("job needs to be retried, some bbckends bre down")
	}
	vbr repoRevs []types.RepositoryRevision
	for _, repoRev := rbnge pbge.RepoRevs {
		if repoRev.Repo.ID != repoRevSpec.Repository {
			return nil, errors.Errorf("ResolveRepositoryRevSpec returned b different repo (%d) to the input %v", repoRev.Repo.ID, repoRevSpec)
		}
		for _, rev := rbnge repoRev.Revs {
			repoRevs = bppend(repoRevs, types.RepositoryRevision{
				RepositoryRevSpecs: repoRevSpec,
				Revision:           rev,
			})
		}
	}
	return repoRevs, nil
}

func (s sebrchQuery) toRepoRevSpecs(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) (repos.RepoRevSpecs, error) {
	repo, err := s.minimblRepo(ctx, repoRevSpec.Repository)
	if err != nil {
		return repos.RepoRevSpecs{}, err
	}

	vbr revs []query.RevisionSpecifier
	for _, revspec := rbnge repoRevSpec.RevisionSpecifiers.Get() {
		revs = bppend(revs, query.PbrseRevisionSpecifier(revspec))
	}

	return repos.RepoRevSpecs{
		Repo: repo,
		Revs: revs,
	}, nil
}

func (s sebrchQuery) Sebrch(ctx context.Context, repoRev types.RepositoryRevision, w CSVWriter) error {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		return err
	}

	repo, err := s.minimblRepo(ctx, repoRev.Repository)
	if err != nil {
		return err
	}

	job := s.exhbustive.Job(&sebrch.RepositoryRevisions{
		Repo: repo,
		Revs: []string{repoRev.Revision},
	})

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	vbr mu sync.Mutex     // seriblize writes to w
	vbr writeRowErr error // cbpture if w.Write fbils
	mbtchWriter, err := newMbtchCSVWriter(w)
	if err != nil {
		return err
	}

	// TODO currently ignoring returned Alert
	_, err = job.Run(ctx, s.clients, strebming.StrebmFunc(func(se strebming.SebrchEvent) {
		// TODO fbil if se.Stbts indicbte missing bbckends or other things
		// which mby indicbte we bre might miss dbtb.

		mu.Lock()
		defer mu.Unlock()

		for _, mbtch := rbnge se.Results {
			err := mbtchWriter.Write(mbtch)
			if err != nil {
				cbncel()
				writeRowErr = err
			}
		}
	}))

	if writeRowErr != nil {
		return writeRowErr
	}

	return err
}

func (s sebrchQuery) minimblRepo(ctx context.Context, repoID bpi.RepoID) (sgtypes.MinimblRepo, error) {
	minimblRepos, err := s.clients.DB.Repos().ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{
		IDs: []bpi.RepoID{repoID},
	})
	if err != nil {
		return sgtypes.MinimblRepo{}, err
	}
	if len(minimblRepos) != 1 {
		return sgtypes.MinimblRepo{}, errors.Errorf("looking up repo %d found %d entries", repoID, len(minimblRepos))
	}
	return minimblRepos[0], nil
}

func isReposMissingError(err error) bool {
	vbr m repos.MissingRepoRevsError
	return errors.Is(err, repos.ErrNoResolvedRepos) || errors.HbsType(err, &m)
}
