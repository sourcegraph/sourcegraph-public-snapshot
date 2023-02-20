package database

import (
	"context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CodeownersStore interface {
	basestore.ShareableStore
	Done(error) error

	CreateCodeownersFile(ctx context.Context, codeowners *types.CodeownersFile) error
	GetCodeownersForRepo(ctx context.Context, id api.RepoID) (*types.CodeownersFile, error)
	DeleteCodeownersForRepo(ctx context.Context, id api.RepoID) error
	ListCodeowners(ctx context.Context) ([]*types.CodeownersFile, error)
}

type codeownersStore struct {
	*basestore.Store
}

func (s *codeownersStore) CreateCodeownersFile(ctx context.Context, file *types.CodeownersFile) error {
	return s.WithTransact(ctx, func(tx CodeownersStore) error {
		if file.CreatedAt.IsZero() {
			file.CreatedAt = timeutil.Now()
		}
		if file.UpdatedAt.IsZero() {
			file.UpdatedAt = file.CreatedAt
		}

		q := sqlf.Sprintf(
			createCodeownersQueryFmtStr,
			sqlf.Join(codeownersColumns, ","),
			file.Contents,
			file.Proto,
			file.RepoID,
			file.CreatedAt,
			file.UpdatedAt,
		)

		_, err := tx.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return err
		}
		return nil
	})
}

var codeownersColumns = []*sqlf.Query{
	sqlf.Sprintf("contents"),
	sqlf.Sprintf("contents_proto"),
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

const createCodeownersQueryFmtStr = `
INSERT INTO codeowners 
(%s)
VALUES (%s, %s, %s, %s, %s)
ON CONFLICT (repo_id) 
DO UPDATE SET contents = EXCLUDED.contents, contents_proto = EXCLUDED.contents_proto, updated_at = EXCLUDED.updated_at
`

func (s *codeownersStore) GetCodeownersForRepo(ctx context.Context, id api.RepoID) (*types.CodeownersFile, error) {
	q := sqlf.Sprintf(
		getCodeownersFileQueryFmtStr,
		sqlf.Join(codeownersColumns, ", "),
		sqlf.Sprintf("repo_id = %s", id),
	)
	codeownersFiles, err := scanCodeowners(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	if len(codeownersFiles) != 1 {
		return nil, errors.New("add a not found error")
	}
	return codeownersFiles[0], nil
}

const getCodeownersFileQueryFmtStr = `
SELECT %s
FROM codeowners 
WHERE %s
LIMIT 1
`

func (s *codeownersStore) DeleteCodeownersForRepo(ctx context.Context, id api.RepoID) error {
	//TODO implement me
	panic("implement me")
}

func (s *codeownersStore) ListCodeowners(ctx context.Context) ([]*types.CodeownersFile, error) {
	//TODO implement me
	panic("implement me")
}

func CodeownersWith(other basestore.ShareableStore) CodeownersStore {
	return &codeownersStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

func (s *codeownersStore) With(other basestore.ShareableStore) CodeownersStore {
	return &codeownersStore{
		Store: s.Store.With(other),
	}
}

func (s *codeownersStore) WithTransact(ctx context.Context, f func(store CodeownersStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&codeownersStore{
			Store: tx,
		})
	})
}

var scanCodeowners = basestore.NewSliceScanner(func(s dbutil.Scanner) (*types.CodeownersFile, error) {
	var c types.CodeownersFile
	c.Proto = new(codeownerspb.File)
	err := scanCodeownersRow(s, &c)
	return &c, err
})

func scanCodeownersRow(sc dbutil.Scanner, c *types.CodeownersFile) error {
	var protoString string
	if err := sc.Scan(
		&c.Contents,
		&protoString,
		&c.RepoID,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return err
	}
	return jsonpb.UnmarshalString(protoString, c.Proto)
}
