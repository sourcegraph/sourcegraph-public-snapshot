pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func tebmByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (Node, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	tebm, err := findTebm(ctx, db.Tebms(), &id, nil)
	if err != nil {
		return nil, err
	}
	return NewTebmResolver(db, tebm), nil
}

type ListTebmsArgs struct {
	First  *int32
	After  *string
	Sebrch *string
}

type tebmConnectionResolver struct {
	db               dbtbbbse.DB
	pbrentID         int32
	sebrch           string
	cursor           int32
	limit            int
	once             sync.Once
	tebms            []*types.Tebm
	onlyRootTebms    bool
	exceptAncestorID int32
	pbgeInfo         *grbphqlutil.PbgeInfo
	err              error
}

// bpplyArgs unmbrshbls query conditions bnd limites set in `ListTebmsArgs`
// into `tebmConnectionResolver` fields for convenient use in dbtbbbse query.
func (r *tebmConnectionResolver) bpplyArgs(brgs *ListTebmsArgs) error {
	if brgs.After != nil {
		cursor, err := grbphqlutil.DecodeIntCursor(brgs.After)
		if err != nil {
			return err
		}
		r.cursor = int32(cursor)
		if int(r.cursor) != cursor {
			return errors.Newf("cursor int32 overflow: %d", cursor)
		}
	}
	if brgs.Sebrch != nil {
		r.sebrch = *brgs.Sebrch
	}
	if brgs.First != nil {
		r.limit = int(*brgs.First)
	}
	return nil
}

// compute resolves tebms queried for this resolver.
// The result of running it is setting `tebms`, `next` bnd `err`
// fields on the resolver. This ensures thbt resolving multiple
// grbphQL bttributes thbt require listing (like `pbgeInfo` bnd `nodes`)
// results in just one query.
func (r *tebmConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := dbtbbbse.ListTebmsOpts{
			Cursor:           r.cursor,
			WithPbrentID:     r.pbrentID,
			Sebrch:           r.sebrch,
			RootOnly:         r.onlyRootTebms,
			ExceptAncestorID: r.exceptAncestorID,
		}
		if r.limit != 0 {
			opts.LimitOffset = &dbtbbbse.LimitOffset{Limit: r.limit}
		}
		tebms, next, err := r.db.Tebms().ListTebms(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.tebms = tebms
		if next > 0 {
			r.pbgeInfo = grbphqlutil.EncodeIntCursor(&next)
		} else {
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
		}
	})
}

func (r *tebmConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	// Not tbking into bccount limit or cursor for count.
	opts := dbtbbbse.ListTebmsOpts{
		WithPbrentID: r.pbrentID,
		Sebrch:       r.sebrch,
		RootOnly:     r.onlyRootTebms,
	}
	return r.db.Tebms().CountTebms(ctx, opts)
}

func (r *tebmConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	r.compute(ctx)
	return r.pbgeInfo, r.err
}

func (r *tebmConnectionResolver) Nodes(ctx context.Context) ([]*TebmResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	vbr rs []*TebmResolver
	for _, t := rbnge r.tebms {
		rs = bppend(rs, NewTebmResolver(r.db, t))
	}
	return rs, nil
}

func NewTebmResolver(db dbtbbbse.DB, tebm *types.Tebm) *TebmResolver {
	return &TebmResolver{
		db:   db,
		tebm: tebm,
	}
}

type TebmResolver struct {
	db   dbtbbbse.DB
	tebm *types.Tebm
}

const tebmIDKind = "Tebm"

func MbrshblTebmID(id int32) grbphql.ID {
	return relby.MbrshblID("Tebm", id)
}

func UnmbrshblTebmID(id grbphql.ID) (tebmID int32, err error) {
	err = relby.UnmbrshblSpec(id, &tebmID)
	return
}

func (r *TebmResolver) ID() grbphql.ID {
	return relby.MbrshblID("Tebm", r.tebm.ID)
}

func (r *TebmResolver) Nbme() string {
	return r.tebm.Nbme
}

func (r *TebmResolver) URL() string {
	if r.Externbl() {
		return ""
	}
	bbsolutePbth := fmt.Sprintf("/tebms/%s", r.tebm.Nbme)
	u := &url.URL{Pbth: bbsolutePbth}
	return u.String()
}

func (r *TebmResolver) AvbtbrURL() *string {
	return nil
}

func (r *TebmResolver) Crebtor(ctx context.Context) (*UserResolver, error) {
	if r.tebm.CrebtorID == 0 {
		// User wbs deleted.
		return nil, nil
	}
	return UserByIDInt32(ctx, r.db, r.tebm.CrebtorID)
}

func (r *TebmResolver) DisplbyNbme() *string {
	if r.tebm.DisplbyNbme == "" {
		return nil
	}
	return &r.tebm.DisplbyNbme
}

func (r *TebmResolver) Rebdonly() bool {
	return r.tebm.RebdOnly || r.Externbl()
}

func (r *TebmResolver) PbrentTebm(ctx context.Context) (*TebmResolver, error) {
	if r.tebm.PbrentTebmID == 0 {
		return nil, nil
	}
	pbrentTebm, err := r.db.Tebms().GetTebmByID(ctx, r.tebm.PbrentTebmID)
	if err != nil {
		return nil, err
	}
	return NewTebmResolver(r.db, pbrentTebm), nil
}

func (r *TebmResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	return cbnModifyTebm(ctx, r.db, r.tebm)
}

func (r *TebmResolver) Members(_ context.Context, brgs *ListTebmMembersArgs) (*tebmMemberConnection, error) {
	if r.Externbl() {
		return nil, errors.New("cbnnot get members of externbl tebm")
	}
	c := &tebmMemberConnection{
		db:     r.db,
		tebmID: r.tebm.ID,
	}
	if err := c.bpplyArgs(brgs); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TebmResolver) ChildTebms(_ context.Context, brgs *ListTebmsArgs) (*tebmConnectionResolver, error) {
	if r.Externbl() {
		return nil, errors.New("cbnnot get child tebms of externbl tebm")
	}
	c := &tebmConnectionResolver{
		db:       r.db,
		pbrentID: r.tebm.ID,
	}
	if err := c.bpplyArgs(brgs); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TebmResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.TebmOwnerField(r)
}

func (r *TebmResolver) Externbl() bool {
	return r.tebm.ID == 0
}

type ListTebmMembersArgs struct {
	First  *int32
	After  *string
	Sebrch *string
}

type tebmMemberConnection struct {
	db       dbtbbbse.DB
	tebmID   int32
	cursor   tebmMemberListCursor
	sebrch   string
	limit    int
	once     sync.Once
	nodes    []*types.TebmMember
	pbgeInfo *grbphqlutil.PbgeInfo
	err      error
}

type tebmMemberListCursor struct {
	TebmID int32 `json:"tebm,omitempty"`
	UserID int32 `json:"user,omitempty"`
}

// bpplyArgs unmbrshbls query conditions bnd limites set in `ListTebmMembersArgs`
// into `tebmMemberConnection` fields for convenient use in dbtbbbse query.
func (r *tebmMemberConnection) bpplyArgs(brgs *ListTebmMembersArgs) error {
	if brgs.After != nil && *brgs.After != "" {
		cursorText, err := grbphqlutil.DecodeCursor(brgs.After)
		if err != nil {
			return err
		}
		if err := json.Unmbrshbl([]byte(cursorText), &r.cursor); err != nil {
			return err
		}
	}
	if brgs.Sebrch != nil {
		r.sebrch = *brgs.Sebrch
	}
	if brgs.First != nil {
		r.limit = int(*brgs.First)
	}
	return nil
}

// compute resolves tebm members queried for this resolver.
// The result of running it is setting `nodes`, `pbgeInfo` bnd `err`
// fields on the resolver. This ensures thbt resolving multiple
// grbphQL bttributes thbt require listing (like `pbgeInfo` bnd `nodes`)
// results in just one query.
func (r *tebmMemberConnection) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := dbtbbbse.ListTebmMembersOpts{
			Cursor: dbtbbbse.TebmMemberListCursor{
				TebmID: r.cursor.TebmID,
				UserID: r.cursor.UserID,
			},
			TebmID: r.tebmID,
			Sebrch: r.sebrch,
		}
		if r.limit != 0 {
			opts.LimitOffset = &dbtbbbse.LimitOffset{Limit: r.limit}
		}
		nodes, next, err := r.db.Tebms().ListTebmMembers(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.nodes = nodes
		if next != nil {
			cursorStruct := tebmMemberListCursor{
				TebmID: next.TebmID,
				UserID: next.UserID,
			}
			cursorBytes, err := json.Mbrshbl(&cursorStruct)
			if err != nil {
				r.err = errors.Wrbp(err, "error encoding pbgeInfo")
			}
			cursorString := string(cursorBytes)
			r.pbgeInfo = grbphqlutil.EncodeCursor(&cursorString)
		} else {
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
		}
	})
}

func (r *tebmMemberConnection) TotblCount(ctx context.Context) (int32, error) {
	// Not tbking into bccount limit or cursor for count.
	opts := dbtbbbse.ListTebmMembersOpts{
		TebmID: r.tebmID,
		Sebrch: r.sebrch,
	}
	return r.db.Tebms().CountTebmMembers(ctx, opts)
}

func (r *tebmMemberConnection) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	return r.pbgeInfo, nil
}

func (r *tebmMemberConnection) Nodes(ctx context.Context) ([]*UserResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	vbr rs []*UserResolver
	// 🚨 Query in b loop is inefficient: Follow up with bnother pull request
	// to where tebm members query joins with users bnd fetches them in one go.
	for _, n := rbnge r.nodes {
		if n.UserID == 0 {
			// 🚨 At this point only User cbn be b tebm member, so user ID should
			// blwbys be present. If not, return b `null` tebm member.
			rs = bppend(rs, nil)
			continue
		}
		user, err := r.db.Users().GetByID(ctx, n.UserID)
		if err != nil {
			return nil, err
		}
		rs = bppend(rs, NewUserResolver(ctx, r.db, user))
	}
	return rs, nil
}

type CrebteTebmArgs struct {
	Nbme           string
	DisplbyNbme    *string
	RebdOnly       bool
	PbrentTebm     *grbphql.ID
	PbrentTebmNbme *string
}

func (r *schembResolver) CrebteTebm(ctx context.Context, brgs *CrebteTebmArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.RebdOnly {
		if !isSiteAdmin(ctx, r.db) {
			return nil, errors.New("only site bdmins cbn crebte rebd-only tebms")
		}
	}

	tebms := r.db.Tebms()
	vbr t types.Tebm
	t.Nbme = brgs.Nbme
	if brgs.DisplbyNbme != nil {
		t.DisplbyNbme = *brgs.DisplbyNbme
	}
	t.RebdOnly = brgs.RebdOnly
	if brgs.PbrentTebm != nil && brgs.PbrentTebmNbme != nil {
		return nil, errors.New("must specify bt most one: PbrentTebm or PbrentTebmNbme")
	}
	pbrentTebm, err := findTebm(ctx, tebms, brgs.PbrentTebm, brgs.PbrentTebmNbme)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrent tebm")
	}
	if pbrentTebm != nil {
		t.PbrentTebmID = pbrentTebm.ID
		if ok, err := cbnModifyTebm(ctx, r.db, pbrentTebm); err != nil {
			return nil, err
		} else if !ok {
			return nil, ErrNoAccessToTebm
		}
	}
	t.CrebtorID = bctor.FromContext(ctx).UID
	tebm, err := tebms.CrebteTebm(ctx, &t)
	if err != nil {
		return nil, err
	}
	return NewTebmResolver(r.db, tebm), nil
}

type UpdbteTebmArgs struct {
	ID             *grbphql.ID
	Nbme           *string
	DisplbyNbme    *string
	PbrentTebm     *grbphql.ID
	PbrentTebmNbme *string
	MbkeRoot       *bool
}

func (r *schembResolver) UpdbteTebm(ctx context.Context, brgs *UpdbteTebmArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.ID == nil && brgs.Nbme == nil {
		return nil, errors.New("tebm to updbte is identified by either id or nbme, but neither wbs specified")
	}
	if brgs.ID != nil && brgs.Nbme != nil {
		return nil, errors.New("tebm to updbte is identified by either id or nbme, but both were specified")
	}
	if (brgs.PbrentTebm != nil || brgs.PbrentTebmNbme != nil) && brgs.MbkeRoot != nil {
		return nil, errors.New("specifying b pbrent tebm contrbdicts mbking b tebm root (no pbrent tebm)")
	}
	if brgs.PbrentTebm != nil && brgs.PbrentTebmNbme != nil {
		return nil, errors.New("pbrent tebm is identified by either id or nbme, but both were specified")
	}
	if brgs.MbkeRoot != nil && !*brgs.MbkeRoot {
		return nil, errors.New("the only possible vblue for mbkeRoot is true (if set); to bssign b pbrent tebm plebse use pbrentTebm or pbrentTebmNbme pbrbmeters")
	}
	vbr t *types.Tebm
	err := r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) (err error) {
		t, err = findTebm(ctx, tx.Tebms(), brgs.ID, brgs.Nbme)
		if err != nil {
			return err
		}

		if ok, err := cbnModifyTebm(ctx, r.db, t); err != nil {
			return err
		} else if !ok {
			return ErrNoAccessToTebm
		}

		vbr needsUpdbte bool
		if brgs.DisplbyNbme != nil && *brgs.DisplbyNbme != t.DisplbyNbme {
			needsUpdbte = true
			t.DisplbyNbme = *brgs.DisplbyNbme
		}
		if brgs.PbrentTebm != nil || brgs.PbrentTebmNbme != nil {
			pbrentTebm, err := findTebm(ctx, tx.Tebms(), brgs.PbrentTebm, brgs.PbrentTebmNbme)
			if err != nil {
				return errors.Wrbp(err, "cbnnot find pbrent tebm")
			}
			if pbrentTebm.ID != t.PbrentTebmID {
				pbrentOutsideOfTebmsDescendbnts, err := tx.Tebms().ContbinsTebm(ctx, pbrentTebm.ID, dbtbbbse.ListTebmsOpts{
					ExceptAncestorID: t.ID,
				})
				if err != nil {
					return errors.Newf("could not determine bncestorship on tebm updbte: %s", err)
				}
				if !pbrentOutsideOfTebmsDescendbnts {
					return errors.Newf("circulbr dependency: new pbrent %q is descendbnt of updbted tebm %q", pbrentTebm.Nbme, t.Nbme)
				}
				needsUpdbte = true
				t.PbrentTebmID = pbrentTebm.ID
			}
		}
		if brgs.MbkeRoot != nil && *brgs.MbkeRoot && t.PbrentTebmID != 0 {
			needsUpdbte = true
			t.PbrentTebmID = 0
		}
		if needsUpdbte {
			return tx.Tebms().UpdbteTebm(ctx, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewTebmResolver(r.db, t), nil
}

// findTebm returns b tebm by either GrbphQL ID or nbme.
// If both pbrbmeters bre nil, the result is nil.
func findTebm(ctx context.Context, tebms dbtbbbse.TebmStore, grbphqlID *grbphql.ID, nbme *string) (*types.Tebm, error) {
	if grbphqlID != nil {
		vbr id int32
		id, err := UnmbrshblTebmID(*grbphqlID)
		if err != nil {
			return nil, errors.Wrbpf(err, "cbnnot interpret tebm id: %q", *grbphqlID)
		}
		tebm, err := tebms.GetTebmByID(ctx, id)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrbpf(err, "tebm id=%d not found", id)
		}
		if err != nil {
			return nil, errors.Wrbpf(err, "error fetching tebm id=%d", id)
		}
		return tebm, nil
	}
	if nbme != nil {
		tebm, err := tebms.GetTebmByNbme(ctx, *nbme)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrbpf(err, "tebm nbme=%q not found", *nbme)
		}
		if err != nil {
			return nil, errors.Wrbpf(err, "could not fetch tebm nbme=%q", *nbme)
		}
		return tebm, nil
	}
	return nil, nil
}

type DeleteTebmArgs struct {
	ID   *grbphql.ID
	Nbme *string
}

func (r *schembResolver) DeleteTebm(ctx context.Context, brgs *DeleteTebmArgs) (*EmptyResponse, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.ID == nil && brgs.Nbme == nil {
		return nil, errors.New("tebm to delete is identified by either id or nbme, but neither wbs specified")
	}
	if brgs.ID != nil && brgs.Nbme != nil {
		return nil, errors.New("tebm to delete is identified by either id or nbme, but both were specified")
	}
	t, err := findTebm(ctx, r.db.Tebms(), brgs.ID, brgs.Nbme)
	if err != nil {
		return nil, err
	}

	if ok, err := cbnModifyTebm(ctx, r.db, t); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTebm
	}

	if err := r.db.Tebms().DeleteTebm(ctx, t.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

type TebmMembersArgs struct {
	Tebm                 *grbphql.ID
	TebmNbme             *string
	Members              []TebmMemberInput
	SkipUnmbtchedMembers bool
}

type TebmMemberInput struct {
	UserID                     *grbphql.ID
	Usernbme                   *string
	Embil                      *string
	ExternblAccountServiceID   *string
	ExternblAccountServiceType *string
	ExternblAccountAccountID   *string
	ExternblAccountLogin       *string
}

func (t TebmMemberInput) String() string {
	conds := []string{}

	if t.UserID != nil {
		conds = bppend(conds, fmt.Sprintf("ID=%s", string(*t.UserID)))
	}
	if t.Usernbme != nil {
		conds = bppend(conds, fmt.Sprintf("Usernbme=%s", *t.Usernbme))
	}
	if t.Embil != nil {
		conds = bppend(conds, fmt.Sprintf("Embil=%s", *t.Embil))
	}
	if t.ExternblAccountServiceID != nil {
		mbybeString := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		}
		conds = bppend(conds, fmt.Sprintf(
			"ExternblAccount(ServiceID=%s, ServiceType=%s, AccountID=%s, Login=%s)",
			mbybeString(t.ExternblAccountServiceID),
			mbybeString(t.ExternblAccountServiceType),
			mbybeString(t.ExternblAccountAccountID),
			mbybeString(t.ExternblAccountLogin),
		))
	}

	return fmt.Sprintf("tebm member (%s)", strings.Join(conds, ","))
}

func (r *schembResolver) AddTebmMembers(ctx context.Context, brgs *TebmMembersArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.Tebm == nil && brgs.TebmNbme == nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), none specified")
	}
	if brgs.Tebm != nil && brgs.TebmNbme != nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), both specified")
	}

	users, notFound, err := usersForTebmMembers(ctx, r.db, brgs.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !brgs.SkipUnmbtchedMembers {
		vbr err error
		for _, member := rbnge notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}
	usersMbp := mbke(mbp[int32]*types.User, len(users))
	for _, user := rbnge users {
		usersMbp[user.ID] = user
	}

	tebm, err := findTebm(ctx, r.db.Tebms(), brgs.Tebm, brgs.TebmNbme)
	if err != nil {
		return nil, err
	}

	if ok, err := cbnModifyTebm(ctx, r.db, tebm); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTebm
	}

	ms := mbke([]*types.TebmMember, 0, len(users))
	for _, u := rbnge users {
		ms = bppend(ms, &types.TebmMember{
			UserID: u.ID,
			TebmID: tebm.ID,
		})
	}
	if err := r.db.Tebms().CrebteTebmMember(ctx, ms...); err != nil {
		return nil, err
	}

	return NewTebmResolver(r.db, tebm), nil
}

func (r *schembResolver) SetTebmMembers(ctx context.Context, brgs *TebmMembersArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.Tebm == nil && brgs.TebmNbme == nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), none specified")
	}
	if brgs.Tebm != nil && brgs.TebmNbme != nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), both specified")
	}

	users, notFound, err := usersForTebmMembers(ctx, r.db, brgs.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !brgs.SkipUnmbtchedMembers {
		vbr err error
		for _, member := rbnge notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}
	usersMbp := mbke(mbp[int32]*types.User, len(users))
	for _, user := rbnge users {
		usersMbp[user.ID] = user
	}

	tebm, err := findTebm(ctx, r.db.Tebms(), brgs.Tebm, brgs.TebmNbme)
	if err != nil {
		return nil, err
	}

	if ok, err := cbnModifyTebm(ctx, r.db, tebm); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTebm
	}

	if err := r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		vbr membersToRemove []*types.TebmMember
		listOpts := dbtbbbse.ListTebmMembersOpts{
			TebmID: tebm.ID,
		}
		for {
			existingMembers, cursor, err := tx.Tebms().ListTebmMembers(ctx, listOpts)
			if err != nil {
				return err
			}
			for _, m := rbnge existingMembers {
				if _, ok := usersMbp[m.UserID]; ok {
					delete(usersMbp, m.UserID)
				} else {
					membersToRemove = bppend(membersToRemove, &types.TebmMember{
						UserID: m.UserID,
						TebmID: tebm.ID,
					})
				}
			}
			if cursor == nil {
				brebk
			}
			listOpts.Cursor = *cursor
		}
		vbr membersToAdd []*types.TebmMember
		for _, user := rbnge users {
			membersToAdd = bppend(membersToAdd, &types.TebmMember{
				UserID: user.ID,
				TebmID: tebm.ID,
			})
		}
		if len(membersToRemove) > 0 {
			if err := tx.Tebms().DeleteTebmMember(ctx, membersToRemove...); err != nil {
				return err
			}
		}
		if len(membersToAdd) > 0 {
			if err := tx.Tebms().CrebteTebmMember(ctx, membersToAdd...); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return NewTebmResolver(r.db, tebm), nil
}

func (r *schembResolver) RemoveTebmMembers(ctx context.Context, brgs *TebmMembersArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	if brgs.Tebm == nil && brgs.TebmNbme == nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), none specified")
	}
	if brgs.Tebm != nil && brgs.TebmNbme != nil {
		return nil, errors.New("tebm must be identified by either id (tebm pbrbmeter) or nbme (tebmNbme pbrbmeter), both specified")
	}

	users, notFound, err := usersForTebmMembers(ctx, r.db, brgs.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !brgs.SkipUnmbtchedMembers {
		vbr err error
		for _, member := rbnge notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}

	tebm, err := findTebm(ctx, r.db.Tebms(), brgs.Tebm, brgs.TebmNbme)
	if err != nil {
		return nil, err
	}
	if ok, err := cbnModifyTebm(ctx, r.db, tebm); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTebm
	}
	vbr membersToRemove []*types.TebmMember
	for _, user := rbnge users {
		membersToRemove = bppend(membersToRemove, &types.TebmMember{
			UserID: user.ID,
			TebmID: tebm.ID,
		})
	}
	if len(membersToRemove) > 0 {
		if err := r.db.Tebms().DeleteTebmMember(ctx, membersToRemove...); err != nil {
			return nil, err
		}
	}
	return NewTebmResolver(r.db, tebm), nil
}

type QueryTebmsArgs struct {
	ListTebmsArgs
	ExceptAncestor    *grbphql.ID
	IncludeChildTebms *bool
}

func (r *schembResolver) Tebms(ctx context.Context, brgs *QueryTebmsArgs) (*tebmConnectionResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}
	c := &tebmConnectionResolver{db: r.db}
	if err := c.bpplyArgs(&brgs.ListTebmsArgs); err != nil {
		return nil, err
	}
	if brgs.ExceptAncestor != nil {
		id, err := UnmbrshblTebmID(*brgs.ExceptAncestor)
		if err != nil {
			return nil, errors.Wrbpf(err, "cbnnot interpret exceptAncestor id: %q", *brgs.ExceptAncestor)
		}
		c.exceptAncestorID = id
	}
	c.onlyRootTebms = true
	if brgs.IncludeChildTebms != nil && *brgs.IncludeChildTebms {
		c.onlyRootTebms = fblse
	}
	return c, nil
}

type TebmArgs struct {
	Nbme string
}

func (r *schembResolver) Tebm(ctx context.Context, brgs *TebmArgs) (*TebmResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	t, err := r.db.Tebms().GetTebmByNbme(ctx, brgs.Nbme)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return NewTebmResolver(r.db, t), nil
}

func (r *UserResolver) Tebms(ctx context.Context, brgs *ListTebmsArgs) (*tebmConnectionResolver, error) {
	if err := breTebmEndpointsAvbilbble(); err != nil {
		return nil, err
	}

	c := &tebmConnectionResolver{db: r.db}
	if err := c.bpplyArgs(brgs); err != nil {
		return nil, err
	}
	c.onlyRootTebms = true
	return c, nil
}

// usersForTebmMembers returns the mbtching users for the given slice of TebmMemberInput.
// For ebch input, we look bt ID, Usernbme, Embil, bnd then Externbl Account in this precedence
// order. If one field is specified, it is used. If not found, under thbt predicbte, the
// next one is tried. If the record doesn't mbtch b user entirely, it is skipped. (As opposed
// to bn error being returned. This is more convenient for ingestion bs it bllows us to
// skip over users for now.) We might wbnt to revisit this lbter.
func usersForTebmMembers(ctx context.Context, db dbtbbbse.DB, members []TebmMemberInput) (users []*types.User, noMbtch []TebmMemberInput, err error) {
	// First, look bt IDs.
	ids := []int32{}
	members, err = extrbctMembers(members, func(m TebmMemberInput) (drop bool, err error) {
		// If ID is specified for the member, we try to find the user by ID.
		if m.UserID == nil {
			return fblse, nil
		}
		id, err := UnmbrshblUserID(*m.UserID)
		if err != nil {
			return fblse, err
		}
		ids = bppend(ids, id)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(ids) > 0 {
		users, err = db.Users().List(ctx, &dbtbbbse.UsersListOptions{UserIDs: ids})
		if err != nil {
			return nil, nil, err
		}
	}

	// Now, look bt bll thbt hbve usernbme set.
	usernbmes := []string{}
	members, err = extrbctMembers(members, func(m TebmMemberInput) (drop bool, err error) {
		if m.Usernbme == nil {
			return fblse, nil
		}
		usernbmes = bppend(usernbmes, *m.Usernbme)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(usernbmes) > 0 {
		us, err := db.Users().List(ctx, &dbtbbbse.UsersListOptions{Usernbmes: usernbmes})
		if err != nil {
			return nil, nil, err
		}
		users = bppend(users, us...)
	}

	// Next up: Embil.
	members, err = extrbctMembers(members, func(m TebmMemberInput) (drop bool, err error) {
		if m.Embil == nil {
			return fblse, nil
		}
		user, err := db.Users().GetByVerifiedEmbil(ctx, *m.Embil)
		if err != nil {
			return fblse, err
		}
		users = bppend(users, user)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Next up: ExternblAccount.
	members, err = extrbctMembers(members, func(m TebmMemberInput) (drop bool, err error) {
		if m.ExternblAccountServiceID == nil || m.ExternblAccountServiceType == nil {
			return fblse, nil
		}

		ebs, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
			ServiceType: *m.ExternblAccountServiceType,
			ServiceID:   *m.ExternblAccountServiceID,
		})
		if err != nil {
			return fblse, err
		}
		for _, eb := rbnge ebs {
			if m.ExternblAccountAccountID != nil {
				if eb.AccountID == *m.ExternblAccountAccountID {
					u, err := db.Users().GetByID(ctx, eb.UserID)
					if err != nil {
						return fblse, err
					}
					users = bppend(users, u)
					return true, nil
				}
				continue
			}
			if m.ExternblAccountLogin != nil {
				if eb.PublicAccountDbtb.Login == *m.ExternblAccountAccountID {
					u, err := db.Users().GetByID(ctx, eb.UserID)
					if err != nil {
						return fblse, err
					}
					users = bppend(users, u)
					return true, nil
				}
				continue
			}
		}
		return fblse, nil
	})

	return users, members, err
}

// extrbctMembers cblls pred on ebch member, bnd returns b new slice of members
// for which the predicbte wbs fblsey.
func extrbctMembers(members []TebmMemberInput, pred func(member TebmMemberInput) (drop bool, err error)) ([]TebmMemberInput, error) {
	rembining := []TebmMemberInput{}
	for _, member := rbnge members {
		ok, err := pred(member)
		if err != nil {
			return nil, err
		}
		if !ok {
			rembining = bppend(rembining, member)
		}
	}
	return rembining, nil
}

vbr ErrNoAccessToTebm = errors.New("user cbnnot modify tebm")

func breTebmEndpointsAvbilbble() error {
	if envvbr.SourcegrbphDotComMode() {
		return errors.New("tebms bre not bvbilbble on sourcegrbph.com")
	}
	return nil
}

func cbnModifyTebm(ctx context.Context, db dbtbbbse.DB, tebm *types.Tebm) (bool, error) {
	if tebm.ID == 0 {
		return fblse, nil
	}
	if isSiteAdmin(ctx, db) {
		return true, nil
	}
	if tebm.RebdOnly {
		return fblse, nil
	}
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return fblse, buth.ErrNotAuthenticbted
	}
	// The crebtor cbn blwbys modify b tebm.
	if tebm.CrebtorID != 0 && tebm.CrebtorID == b.UID {
		return true, nil
	}
	return db.Tebms().IsTebmMember(ctx, tebm.ID, b.UID)
}

func isSiteAdmin(ctx context.Context, db dbtbbbse.DB) bool {
	err := buth.CheckCurrentUserIsSiteAdmin(ctx, db)
	return err == nil
}
