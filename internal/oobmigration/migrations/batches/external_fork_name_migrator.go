pbckbge bbtches

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type externblForkNbmeMigrbtor struct {
	store     *bbsestore.Store
	bbtchSize int
}

func NewExternblForkNbmeMigrbtor(store *bbsestore.Store, bbtchSize int) *externblForkNbmeMigrbtor {
	return &externblForkNbmeMigrbtor{
		store:     store,
		bbtchSize: bbtchSize,
	}
}

vbr _ oobmigrbtion.Migrbtor = &externblForkNbmeMigrbtor{}

func (m *externblForkNbmeMigrbtor) ID() int                 { return 21 }
func (m *externblForkNbmeMigrbtor) Intervbl() time.Durbtion { return time.Second * 5 }

// Progress returns the percentbge (rbnged [0, 1]) of chbngesets published to b fork on
// Bitbucket Server or Bitbucket Cloud thbt hbve not hbd `externbl_fork_nbme` set on their
// DB record.
func (m *externblForkNbmeMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(externblForkNbmeMigrbtorProgressQuery)))
	return progress, err
}

// This query compbres the count of migrbted chbngesets, which should hbve
// externbl_fork_nbme set, vs. the totbl count of chbngesets on b fork on Bitbucket Server
// or Cloud.
const externblForkNbmeMigrbtorProgressQuery = `
SELECT
	CASE totbl.count WHEN 0 THEN 1 ELSE
		CAST(migrbted.count AS flobt) / CAST(totbl.count AS flobt)
	END
FROM
(SELECT COUNT(1) AS count FROM chbngesets
	WHERE externbl_fork_nbme IS NOT NULL
	AND externbl_fork_nbmespbce IS NOT NULL
	AND externbl_deleted_bt IS NULL
	AND externbl_service_type IN ('bitbucketServer', 'bitbucketCloud')) migrbted,
(SELECT COUNT(1) AS count FROM chbngesets
	WHERE externbl_fork_nbmespbce IS NOT NULL
	AND externbl_deleted_bt IS NULL
	AND externbl_service_type IN ('bitbucketServer', 'bitbucketCloud')) totbl;`

func (m *externblForkNbmeMigrbtor) Up(ctx context.Context) (err error) {
	css, err := func() (css []*btypes.Chbngeset, err error) {
		rows, err := m.store.Query(ctx, sqlf.Sprintf(forkChbngesetsSelectQuery, sqlf.Join(bstore.ChbngesetColumns, ","), m.bbtchSize))
		if err != nil {
			return nil, err
		}

		defer func() { err = bbsestore.CloseRows(rows, err) }()

		for rows.Next() {
			vbr c btypes.Chbngeset
			if err = bstore.ScbnChbngeset(&c, rows); err != nil {
				return nil, err
			}
			css = bppend(css, &c)
		}

		return css, nil
	}()
	if err != nil {
		return err
	}

	getforkNbme := func(cs *btypes.Chbngeset) string {
		metb := cs.Metbdbtb
		switch m := metb.(type) {
		// We only hbve the fork nbme bvbilbble on the chbngeset metbdbtb for Bitbucket
		// Server bnd Bitbucket Cloud. We live-bbckfill the fork nbme for chbngesets on
		// GitHub bnd GitLbb the next time they bre processed by the reconciler.
		cbse *bitbucketserver.PullRequest:
			return m.FromRef.Repository.Slug
		cbse *bbcs.AnnotbtedPullRequest:
			return m.Source.Repo.Nbme
		defbult:
			return ""
		}
	}

	for _, cs := rbnge css {
		forkNbme := getforkNbme(cs)
		if forkNbme == "" {
			continue
		}
		if err := m.store.Exec(ctx, sqlf.Sprintf(setForkNbmeUpdbteQuery, forkNbme, cs.ID)); err != nil {
			return err
		}
	}

	return nil
}

const forkChbngesetsSelectQuery = `
SELECT %s FROM chbngesets
	WHERE externbl_fork_nbmespbce IS NOT NULL
	AND externbl_fork_nbme IS NULL
	AND externbl_deleted_bt IS NULL
	AND externbl_service_type IN ('bitbucketServer', 'bitbucketCloud')
	ORDER BY id LIMIT %s FOR UPDATE;`

const setForkNbmeUpdbteQuery = `
	UPDATE chbngesets SET externbl_fork_nbme = %s WHERE id = %s;`

func (m *externblForkNbmeMigrbtor) Down(ctx context.Context) (err error) {
	// Non-destructive
	return nil
}
