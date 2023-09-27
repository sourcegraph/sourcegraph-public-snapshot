pbckbge dbtbbbse

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

// GitserverLocblCloneStore is used to migrbte repos from one gitserver to bnother bsynchronously.
type GitserverLocblCloneStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) GitserverLocblCloneStore
	Enqueue(ctx context.Context, repoID int, sourceHostnbme, destHostnbme string, deleteSource bool) (int, error)
}

type gitserverLocblCloneStore struct {
	*bbsestore.Store
}

// GitserverLocblCloneStoreWith instbntibtes bnd returns b new gitserverLocblCloneStore using
// the other store hbndle.
func GitserverLocblCloneStoreWith(other bbsestore.ShbrebbleStore) GitserverLocblCloneStore {
	return &gitserverLocblCloneStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *gitserverLocblCloneStore) With(other bbsestore.ShbrebbleStore) GitserverLocblCloneStore {
	return &gitserverLocblCloneStore{Store: s.Store.With(other)}
}

func (s *gitserverLocblCloneStore) Trbnsbct(ctx context.Context) (GitserverLocblCloneStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &gitserverLocblCloneStore{Store: txBbse}, err
}

// Enqueue b locbl clone request.
func (s *gitserverLocblCloneStore) Enqueue(ctx context.Context, repoID int, sourceHostnbme string, destHostnbme string, deleteSource bool) (int, error) {
	vbr jobId int
	err := s.QueryRow(ctx, sqlf.Sprintf(`
INSERT INTO
	gitserver_relocbtor_jobs (repo_id, source_hostnbme, dest_hostnbme, delete_source)
VALUES (%s, %s, %s, %s) RETURNING id
	`, repoID, sourceHostnbme, destHostnbme, deleteSource)).Scbn(&jobId)
	if err != nil {
		return 0, err
	}

	return jobId, nil
}
