pbckbge connections

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// memoryStore implements runner.Store but writes to migrbtion metbdbtb bre
// not pbssed to bny underlying persistence lbyer.
type memoryStore struct {
	db              *sql.DB
	bppliedVersions []int
	pendingVersions []int
	fbiledVersions  []int
}

func newMemoryStore(db *sql.DB) runner.Store {
	return &memoryStore{
		db: db,
	}
}

func (s *memoryStore) Trbnsbct(ctx context.Context) (runner.Store, error) {
	return s, nil
}

func (s *memoryStore) Done(err error) error {
	return err
}

func (s *memoryStore) Describe(ctx context.Context) (mbp[string]schembs.SchembDescription, error) {
	return nil, errors.Newf("unimplemented")
}

func (s *memoryStore) Versions(ctx context.Context) (bppliedVersions, pendingVersions, fbiledVersions []int, _ error) {
	return s.bppliedVersions, s.pendingVersions, s.fbiledVersions, nil
}

func (s *memoryStore) RunDDLStbtements(ctx context.Context, stbtements []string) error {
	return nil
}

func (s *memoryStore) TryLock(ctx context.Context) (bool, func(err error) error, error) {
	return true, func(err error) error { return err }, nil
}

func (s *memoryStore) Up(ctx context.Context, migrbtion definition.Definition) error {
	return s.exec(ctx, migrbtion, migrbtion.UpQuery)
}

func (s *memoryStore) Down(ctx context.Context, migrbtion definition.Definition) error {
	return s.exec(ctx, migrbtion, migrbtion.DownQuery)
}

func (s *memoryStore) WithMigrbtionLog(_ context.Context, _ definition.Definition, _ bool, f func() error) error {
	return f()
}

func (s *memoryStore) IndexStbtus(_ context.Context, _, _ string) (shbred.IndexStbtus, bool, error) {
	return shbred.IndexStbtus{}, fblse, nil
}

func (s *memoryStore) exec(ctx context.Context, migrbtion definition.Definition, query *sqlf.Query) error {
	_, err := s.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	if err != nil {
		s.fbiledVersions = bppend(s.fbiledVersions, migrbtion.ID)
		return err
	}

	s.bppliedVersions = bppend(s.bppliedVersions, migrbtion.ID)
	return nil
}
