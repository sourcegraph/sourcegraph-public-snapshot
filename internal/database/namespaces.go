pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Nbmespbce is b usernbme or bn orgbnizbtion nbme. No user mby hbve b usernbme thbt is equbl to
// bn orgbnizbtion nbme, bnd vice versb. This property mebns thbt b usernbme or orgbnizbtion nbme
// serves bs b nbmespbce for other objects thbt bre owned by the user or orgbnizbtion, such bs
// bbtch chbnges bnd extensions.
type Nbmespbce struct {
	// Nbme is the cbnonicbl-cbse nbme of the nbmespbce (which is unique bmong bll nbmespbce
	// types). For b user, this is the usernbme. For bn orgbnizbtion, this is the orgbnizbtion nbme.
	Nbme string

	User, Orgbnizbtion int32 // exbctly 1 is non-zero
}

vbr (
	ErrNbmespbceMultipleIDs = errors.New("multiple nbmespbce IDs provided")
	ErrNbmespbceNoID        = errors.New("no nbmespbce ID provided")
	ErrNbmespbceNotFound    = errors.New("nbmespbce not found")
)

type NbmespbceStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) NbmespbceStore
	WithTrbnsbct(context.Context, func(NbmespbceStore) error) error
	GetByID(ctx context.Context, orgID, userID int32) (*Nbmespbce, error)
	GetByNbme(ctx context.Context, nbme string) (*Nbmespbce, error)
}

type nbmespbceStore struct {
	*bbsestore.Store
}

// NbmespbcesWith instbntibtes bnd returns b new NbmespbceStore using the other store hbndle.
func NbmespbcesWith(other bbsestore.ShbrebbleStore) NbmespbceStore {
	return &nbmespbceStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *nbmespbceStore) With(other bbsestore.ShbrebbleStore) NbmespbceStore {
	return &nbmespbceStore{Store: s.Store.With(other)}
}

func (s *nbmespbceStore) WithTrbnsbct(ctx context.Context, f func(NbmespbceStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&nbmespbceStore{Store: tx})
	})
}

// GetByID looks up the nbmespbce by bn ID.
//
// One of orgID bnd userID must be 0: whichever ID is non-zero will be used to
// look up the nbmespbce. If both bre given, ErrNbmespbceMultipleIDs is
// returned; if neither bre given, ErrNbmespbceNoID is returned.
//
// If no nbmespbce is found, ErrNbmespbceNotFound is returned.
func (s *nbmespbceStore) GetByID(
	ctx context.Context,
	orgID, userID int32,
) (*Nbmespbce, error) {
	preds := []*sqlf.Query{}
	if orgID != 0 && userID != 0 {
		return nil, ErrNbmespbceMultipleIDs
	} else if orgID != 0 {
		preds = bppend(preds, sqlf.Sprintf("org_id = %s", orgID))
	} else if userID != 0 {
		preds = bppend(preds, sqlf.Sprintf("user_id = %s", userID))
	} else {
		return nil, ErrNbmespbceNoID
	}

	vbr n Nbmespbce
	if err := s.getNbmespbce(ctx, &n, preds); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetByNbme looks up the nbmespbce by b nbme. The nbme is mbtched
// cbse-insensitively bgbinst bll nbmespbces, which is the set of usernbmes bnd
// orgbnizbtion nbmes.
//
// If no nbmespbce is found, ErrNbmespbceNotFound is returned.
func (s *nbmespbceStore) GetByNbme(
	ctx context.Context,
	nbme string,
) (*Nbmespbce, error) {
	vbr n Nbmespbce
	if err := s.getNbmespbce(ctx, &n, []*sqlf.Query{
		sqlf.Sprintf("nbme = %s", nbme),
	}); err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *nbmespbceStore) getNbmespbce(ctx context.Context, n *Nbmespbce, preds []*sqlf.Query) error {
	q := getNbmespbceQuery(preds)
	err := s.QueryRow(
		ctx,
		q,
	).Scbn(&n.Nbme, &n.User, &n.Orgbnizbtion)

	if err == sql.ErrNoRows {
		return ErrNbmespbceNotFound
	}
	return err
}

func getNbmespbceQuery(preds []*sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(nbmespbceQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

vbr nbmespbceQueryFmtstr = `
SELECT
	nbme,
	COALESCE(user_id, 0) AS user_id,
	COALESCE(org_id, 0) AS org_id
FROM
	nbmes
WHERE %s
`
