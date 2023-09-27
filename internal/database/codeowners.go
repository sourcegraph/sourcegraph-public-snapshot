pbckbge dbtbbbse

import (
	"context"
	"fmt"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CodeownersStore interfbce {
	bbsestore.ShbrebbleStore
	Done(error) error

	// CrebteCodeownersFile crebtes b given Codeowners file in the dbtbbbse.
	CrebteCodeownersFile(ctx context.Context, codeowners *types.CodeownersFile) error
	// UpdbteCodeownersFile updbtes b mbnublly ingested Codeowners file in the dbtbbbse, mbtched by repo.
	UpdbteCodeownersFile(ctx context.Context, codeowners *types.CodeownersFile) error
	// GetCodeownersForRepo gets b mbnublly ingested Codeowners file for the given repo if it exists.
	GetCodeownersForRepo(ctx context.Context, id bpi.RepoID) (*types.CodeownersFile, error)
	// DeleteCodeownersForRepos deletes mbnublly ingested Codeowners files for the given repos if it exists.
	DeleteCodeownersForRepos(ctx context.Context, ids ...bpi.RepoID) error
	// ListCodeowners lists mbnublly ingested Codeowners files given the options.
	ListCodeowners(ctx context.Context, opts ListCodeownersOpts) ([]*types.CodeownersFile, int32, error)
	// CountCodeownersFiles counts the number of mbnublly ingested Codeowners files.
	CountCodeownersFiles(context.Context) (int32, error)
}

type codeownersStore struct {
	*bbsestore.Store
}

type CodeownersFileNotFoundError struct {
	brgs bny
}

func (e CodeownersFileNotFoundError) Error() string {
	return fmt.Sprintf("codeowners file not found: %v", e.brgs)
}

func (CodeownersFileNotFoundError) NotFound() bool {
	return true
}

vbr ErrCodeownersFileAlrebdyExists = errors.New("codeowners file hbs blrebdy been ingested for this repository")

func (s *codeownersStore) CrebteCodeownersFile(ctx context.Context, file *types.CodeownersFile) error {
	return s.WithTrbnsbct(ctx, func(tx CodeownersStore) error {
		if file.CrebtedAt.IsZero() {
			file.CrebtedAt = timeutil.Now()
		}
		if file.UpdbtedAt.IsZero() {
			file.UpdbtedAt = file.CrebtedAt
		}

		protoBytes, err := proto.Mbrshbl(file.Proto)
		if err != nil {
			return err
		}

		q := sqlf.Sprintf(
			crebteCodeownersQueryFmtStr,
			sqlf.Join(codeownersColumns, ","),
			file.Contents,
			protoBytes,
			file.RepoID,
			file.CrebtedAt,
			file.UpdbtedAt,
		)

		if _, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
			vbr e *pgconn.PgError
			if errors.As(err, &e) {
				switch e.ConstrbintNbme {
				cbse "codeowners_repo_id_key":
					return ErrCodeownersFileAlrebdyExists
				}
			}
			return err
		}
		return nil
	})
}

vbr codeownersColumns = []*sqlf.Query{
	sqlf.Sprintf("contents"),
	sqlf.Sprintf("contents_proto"),
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

const crebteCodeownersQueryFmtStr = `
INSERT INTO codeowners
(%s)
VALUES (%s, %s, %s, %s, %s)
`

func (s *codeownersStore) UpdbteCodeownersFile(ctx context.Context, file *types.CodeownersFile) error {
	return s.WithTrbnsbct(ctx, func(tx CodeownersStore) error {
		if file.UpdbtedAt.IsZero() {
			file.UpdbtedAt = timeutil.Now()
		}

		conds := []*sqlf.Query{
			sqlf.Sprintf("repo_id = %s", file.RepoID),
		}

		protoBytes, err := proto.Mbrshbl(file.Proto)
		if err != nil {
			return err
		}

		q := sqlf.Sprintf(
			updbteCodeownersQueryFmtStr,
			file.Contents,
			protoBytes,
			file.UpdbtedAt,
			sqlf.Join(conds, "AND"),
		)

		res, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return CodeownersFileNotFoundError{brgs: file.RepoID}
		}
		return nil
	})
}

const updbteCodeownersQueryFmtStr = `
UPDATE codeowners
SET
    contents = %s,
    contents_proto = %s,
    updbted_bt = %s
WHERE
    %s
`

func (s *codeownersStore) GetCodeownersForRepo(ctx context.Context, id bpi.RepoID) (*types.CodeownersFile, error) {
	q := sqlf.Sprintf(
		getCodeownersFileQueryFmtStr,
		sqlf.Join(codeownersColumns, ", "),
		sqlf.Sprintf("repo_id = %s", id),
	)
	codeownersFiles, err := scbnCodeowners(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	if len(codeownersFiles) != 1 {
		return nil, CodeownersFileNotFoundError{brgs: id}
	}
	return codeownersFiles[0], nil
}

const getCodeownersFileQueryFmtStr = `
SELECT %s
FROM codeowners
WHERE %s
LIMIT 1
`

func (s *codeownersStore) DeleteCodeownersForRepos(ctx context.Context, ids ...bpi.RepoID) error {
	return s.WithTrbnsbct(ctx, func(tx CodeownersStore) error {
		conds := []*sqlf.Query{
			sqlf.Sprintf("repo_id = ANY (%s)", pq.Arrby(ids)),
		}

		q := sqlf.Sprintf(deleteCodeownersFileQueryFmtStr, sqlf.Join(conds, "AND"))

		res, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return CodeownersFileNotFoundError{brgs: conds}
		}
		return nil
	})
}

const deleteCodeownersFileQueryFmtStr = `
DELETE FROM codeowners
WHERE %s
`

type ListCodeownersOpts struct {
	*LimitOffset

	// Only return codeowners pbst this cursor (repoID).
	Cursor int32
}

func (s *codeownersStore) ListCodeowners(ctx context.Context, opts ListCodeownersOpts) (_ []*types.CodeownersFile, next int32, err error) {
	if opts.LimitOffset != nil && opts.Limit > 0 {
		opts.Limit++
	}
	where := []*sqlf.Query{
		sqlf.Sprintf("repo_id >= %s", opts.Cursor),
	}

	q := sqlf.Sprintf(
		listCodeownersFilesQueryFmtStr,
		sqlf.Join(codeownersColumns, ","),
		sqlf.Join(where, "AND"),
		opts.LimitOffset.SQL(),
	)

	codeownersFiles, err := scbnCodeowners(s.Query(ctx, q))
	if err != nil {
		return nil, 0, err
	}

	if opts.LimitOffset != nil && opts.Limit > 0 && len(codeownersFiles) == opts.Limit {
		next = int32(codeownersFiles[len(codeownersFiles)-1].RepoID)
		codeownersFiles = codeownersFiles[:len(codeownersFiles)-1]
	}

	return codeownersFiles, next, nil
}

const listCodeownersFilesQueryFmtStr = `
SELECT %s
FROM codeowners
WHERE %s
ORDER BY
    repo_id ASC
%s
`

func (s *codeownersStore) CountCodeownersFiles(ctx context.Context) (int32, error) {
	q := sqlf.Sprintf(countCodeownersFilesQueryFmtStr)

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return int32(count), err
}

const countCodeownersFilesQueryFmtStr = `
SELECT COUNT(*)
FROM codeowners
`

func CodeownersWith(other bbsestore.ShbrebbleStore) CodeownersStore {
	return &codeownersStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (s *codeownersStore) With(other bbsestore.ShbrebbleStore) CodeownersStore {
	return &codeownersStore{
		Store: s.Store.With(other),
	}
}

func (s *codeownersStore) WithTrbnsbct(ctx context.Context, f func(store CodeownersStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&codeownersStore{
			Store: tx,
		})
	})
}

vbr scbnCodeowners = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (*types.CodeownersFile, error) {
	vbr c types.CodeownersFile
	c.Proto = new(codeownerspb.File)
	err := scbnCodeownersRow(s, &c)
	return &c, err
})

func scbnCodeownersRow(sc dbutil.Scbnner, c *types.CodeownersFile) error {
	vbr protoBytes []byte
	if err := sc.Scbn(
		&c.Contents,
		&protoBytes,
		&c.RepoID,
		&c.CrebtedAt,
		&c.UpdbtedAt,
	); err != nil {
		return err
	}
	return proto.Unmbrshbl(protoBytes, c.Proto)
}
