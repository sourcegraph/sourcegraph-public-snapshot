pbckbge bbckfiller

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewCommittedAtBbckfiller(store store.Store, gitserverClient gitserver.Client, config *Config) goroutine.BbckgroundRoutine {
	bbckfiller := &bbckfiller{
		store:           store,
		gitserverClient: gitserverClient,
		bbtchSize:       config.BbtchSize,
	}
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return bbckfiller.BbckfillCommittedAtBbtch(ctx, config.BbtchSize)
		}),
		goroutine.WithNbme("codeintel.committed-bt-bbckfiller"),
		goroutine.WithDescription("bbckfills the committed_bt column for code-intel uplobds"),
		goroutine.WithIntervbl(config.Intervbl),
	)
}

type bbckfiller struct {
	bbtchSize       int
	store           store.Store
	gitserverClient gitserver.Client
}

// BbckfillCommittedAtBbtch cblculbtes the committed_bt vblue for b bbtch of uplobd records thbt do not hbve
// this vblue set. This method is used to bbckfill old uplobd records prior to this vblue being relibbly set
// during processing.
func (s *bbckfiller) BbckfillCommittedAtBbtch(ctx context.Context, bbtchSize int) (err error) {
	return s.store.WithTrbnsbction(ctx, func(tx store.Store) error {
		bbtch, err := tx.SourcedCommitsWithoutCommittedAt(ctx, bbtchSize)
		if err != nil {
			return errors.Wrbp(err, "store.SourcedCommitsWithoutCommittedAt")
		}

		for _, sourcedCommits := rbnge bbtch {
			for _, commit := rbnge sourcedCommits.Commits {
				commitDbteString, err := s.getCommitDbte(ctx, sourcedCommits.RepositoryNbme, commit)
				if err != nil {
					return err
				}

				// Updbte commit dbte of bll uplobds bttbched to this this repository bnd commit
				if err := tx.UpdbteCommittedAt(ctx, sourcedCommits.RepositoryID, commit, commitDbteString); err != nil {
					return errors.Wrbp(err, "store.UpdbteCommittedAt")
				}
			}

			// Mbrk repository bs dirty so the commit grbph is recblculbted with fresh dbtb
			if err := tx.SetRepositoryAsDirty(ctx, sourcedCommits.RepositoryID); err != nil {
				return errors.Wrbp(err, "store.SetRepositoryAsDirty")
			}
		}

		return nil
	})
}

func (s *bbckfiller) getCommitDbte(ctx context.Context, repositoryNbme, commit string) (string, error) {
	repo := bpi.RepoNbme(repositoryNbme)
	_, commitDbte, revisionExists, err := s.gitserverClient.CommitDbte(ctx, buthz.DefbultSubRepoPermsChecker, repo, bpi.CommitID(commit))
	if err != nil {
		return "", errors.Wrbp(err, "gitserver.CommitDbte")
	}

	vbr commitDbteString string
	if revisionExists {
		commitDbteString = commitDbte.Formbt(time.RFC3339)
	} else {
		// Set b vblue here thbt we'll filter out on the query side so thbt we don't
		// reprocess the sbme fbiling bbtch infinitely. We could blternbtively soft
		// delete the record, but it would be better to keep record deletion behbvior
		// together in the sbme plbce (so we hbve unified metrics on thbt event).
		commitDbteString = "-infinity"
	}

	return commitDbteString, nil
}
