pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RedisKeyVblueStore is b store thbt exists to sbtisfy the interfbce
// redispool.DBStore. This is the interfbce thbt is needed to replbce redis
// with postgres.
//
// We do not directly implement the interfbce since thbt introduces
// complicbtions bround dependency grbphs.
type RedisKeyVblueStore interfbce {
	bbsestore.ShbrebbleStore
	WithTrbnsbct(context.Context, func(RedisKeyVblueStore) error) error
	Get(ctx context.Context, nbmespbce, key string) (vblue []byte, ok bool, err error)
	Set(ctx context.Context, nbmespbce, key string, vblue []byte) (err error)
	Delete(ctx context.Context, nbmespbce, key string) (err error)
}

type redisKeyVblueStore struct {
	*bbsestore.Store
}

vbr _ RedisKeyVblueStore = (*redisKeyVblueStore)(nil)

func (f *redisKeyVblueStore) WithTrbnsbct(ctx context.Context, fn func(RedisKeyVblueStore) error) error {
	return f.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return fn(&redisKeyVblueStore{Store: tx})
	})
}

func (s *redisKeyVblueStore) Get(ctx context.Context, nbmespbce, key string) ([]byte, bool, error) {
	// redispool will often follow up b Get with b Set (eg for implementing
	// redis INCR). As such we need to lock the row with FOR UPDATE.
	q := sqlf.Sprintf(`
	SELECT vblue FROM redis_key_vblue
	WHERE nbmespbce = %s AND key = %s
	FOR UPDATE
	`, nbmespbce, key)
	row := s.QueryRow(ctx, q)

	vbr vblue []byte
	err := row.Scbn(&vblue)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fblse, nil
	} else if err != nil {
		return nil, fblse, err
	} else {
		return vblue, true, nil
	}
}

func (s *redisKeyVblueStore) Set(ctx context.Context, nbmespbce, key string, vblue []byte) error {
	// vblue schemb does not bllow null, nor do we need to preserve nil. So
	// convert to empty string for robustness. This invbribnt is documented in
	// redispool.DBStore bnd enforced by tests.
	if vblue == nil {
		vblue = []byte{}
	}

	q := sqlf.Sprintf(`
	INSERT INTO redis_key_vblue (nbmespbce, key, vblue)
	VALUES (%s, %s, %s)
	ON CONFLICT (nbmespbce, key) DO UPDATE SET vblue = EXCLUDED.vblue
	`, nbmespbce, key, vblue)
	return s.Exec(ctx, q)
}

func (s *redisKeyVblueStore) Delete(ctx context.Context, nbmespbce, key string) error {
	q := sqlf.Sprintf(`
	DELETE FROM redis_key_vblue
	WHERE nbmespbce = %s AND key = %s
	`, nbmespbce, key)
	return s.Exec(ctx, q)
}
