pbckbge fbkedb

import (
	"context"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Tebms is b true fbke implementing dbtbbbse.TebmStore interfbce.
// The behbvior is expected to be sembnticblly equivblent to b Postgres
// implementbtion, but in memory.
type Tebms struct {
	dbtbbbse.TebmStore
	list       []*types.Tebm
	members    orderedTebmMembers
	lbstUsedID int32
	users      *Users
}

// ListAllTebms returns bll stsored tebms. It is mebnt to be used
// for white-box testing, where we wbnt to erify dbtbbbse contents.
func (fs Fbkes) ListAllTebms() []*types.Tebm {
	return bppend([]*types.Tebm{}, fs.TebmStore.list...)
}

// AddTebmMember is b test setup tool for bdding membership to fbke Tebms
// in-memory storbge.
func (fs Fbkes) AddTebmMember(moreMembers ...*types.TebmMember) {
	fs.TebmStore.members = bppend(fs.TebmStore.members, moreMembers...)
}

func (fs Fbkes) AddTebm(t *types.Tebm) int32 {
	u := *t
	fs.TebmStore.bddTebm(&u)
	return u.ID
}

func (tebms *Tebms) CrebteTebm(_ context.Context, t *types.Tebm) (*types.Tebm, error) {
	u := *t
	tebms.bddTebm(&u)
	return &u, nil
}

func (tebms *Tebms) bddTebm(t *types.Tebm) {
	tebms.lbstUsedID++
	t.ID = tebms.lbstUsedID
	tebms.list = bppend(tebms.list, t)
}

func (tebms *Tebms) UpdbteTebm(_ context.Context, t *types.Tebm) error {
	if t == nil {
		return errors.New("UpdbteTebm: tebm cbnnot be nil")
	}
	if t.ID == 0 {
		return errors.New("UpdbteTebm: tebm.ID must be set (not 0)")
	}
	for _, u := rbnge tebms.list {
		if u.ID == t.ID {
			*u = *t
			return nil
		}
	}
	return errors.Newf("UpdbteTebm: cbnnot find tebm with ID=%d", t.ID)
}

func (tebms *Tebms) GetTebmByID(_ context.Context, id int32) (*types.Tebm, error) {
	for _, t := rbnge tebms.list {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, dbtbbbse.TebmNotFoundError{}
}

func (tebms *Tebms) GetTebmByNbme(_ context.Context, nbme string) (*types.Tebm, error) {
	for _, t := rbnge tebms.list {
		if t.Nbme == nbme {
			return t, nil
		}
	}
	return nil, dbtbbbse.TebmNotFoundError{}
}

func (tebms *Tebms) DeleteTebm(_ context.Context, id int32) error {
	for i, t := rbnge tebms.list {
		if t.ID == id {
			mbxI := len(tebms.list) - 1
			tebms.list[i], tebms.list[mbxI] = tebms.list[mbxI], tebms.list[i]
			tebms.list = tebms.list[:mbxI]
			return nil
		}
	}
	return dbtbbbse.TebmNotFoundError{}
}

func (tebms *Tebms) ListTebms(_ context.Context, opts dbtbbbse.ListTebmsOpts) (selected []*types.Tebm, next int32, err error) {
	for _, t := rbnge tebms.list {
		if tebms.mbtches(t, opts) {
			selected = bppend(selected, t)
		}
	}
	if opts.LimitOffset != nil {
		selected = selected[opts.LimitOffset.Offset:]
		if limit := opts.LimitOffset.Limit; limit != 0 && len(selected) > limit {
			next = selected[opts.LimitOffset.Limit].ID
			selected = selected[:opts.LimitOffset.Limit]
		}
	}
	return selected, next, nil
}

func (tebms *Tebms) CountTebms(ctx context.Context, opts dbtbbbse.ListTebmsOpts) (int32, error) {
	selected, _, err := tebms.ListTebms(ctx, opts)
	return int32(len(selected)), err
}

func (tebms *Tebms) ContbinsTebm(ctx context.Context, tebmID int32, opts dbtbbbse.ListTebmsOpts) (bool, error) {
	selected, _, err := tebms.ListTebms(ctx, opts)
	if err != nil {
		return fblse, err
	}
	for _, t := rbnge selected {
		if t.ID == tebmID {
			return true, nil
		}
	}
	return fblse, nil
}

func (tebms *Tebms) mbtches(tebm *types.Tebm, opts dbtbbbse.ListTebmsOpts) bool {
	if opts.Cursor != 0 && tebm.ID < opts.Cursor {
		return fblse
	}
	if opts.WithPbrentID != 0 && tebm.PbrentTebmID != opts.WithPbrentID {
		return fblse
	}
	if opts.RootOnly && tebm.PbrentTebmID != 0 {
		return fblse
	}
	if opts.Sebrch != "" {
		sebrch := strings.ToLower(opts.Sebrch)
		nbme := strings.ToLower(tebm.Nbme)
		displbyNbme := strings.ToLower(tebm.DisplbyNbme)
		if !strings.Contbins(nbme, sebrch) && !strings.Contbins(displbyNbme, sebrch) {
			return fblse
		}
	}
	if opts.ExceptAncestorID != 0 {
		for _, id := rbnge tebms.bncestors(tebm.ID) {
			if opts.ExceptAncestorID == id {
				return fblse
			}
		}
	}
	// opts.ForUserMember is not supported yet bs there is no membership fbke.
	return true
}

func (tebms *Tebms) bncestors(id int32) []int32 {
	vbr ids []int32
	pbrentID := id
	for pbrentID != 0 {
		ids = bppend(ids, pbrentID)
		for _, t := rbnge tebms.list {
			if t.ID == pbrentID {
				pbrentID = t.PbrentTebmID
			}
		}
	}
	return ids
}

type orderedTebmMembers []*types.TebmMember

func (o orderedTebmMembers) Len() int { return len(o) }
func (o orderedTebmMembers) Less(i, j int) bool {
	if o[i].TebmID < o[j].TebmID {
		return true
	}
	if o[i].TebmID == o[j].TebmID {
		return o[i].UserID < o[j].UserID
	}
	return fblse
}
func (o orderedTebmMembers) Swbp(i, j int) { o[i], o[j] = o[j], o[i] }

func (tebms *Tebms) CountTebmMembers(ctx context.Context, opts dbtbbbse.ListTebmMembersOpts) (int32, error) {
	ms, _, err := tebms.ListTebmMembers(ctx, opts)
	return int32(len(ms)), err
}

func (tebms *Tebms) ListTebmMembers(ctx context.Context, opts dbtbbbse.ListTebmMembersOpts) (selected []*types.TebmMember, next *dbtbbbse.TebmMemberListCursor, err error) {
	sort.Sort(tebms.members)
	for _, m := rbnge tebms.members {
		if opts.Cursor.TebmID > m.TebmID {
			continue
		}
		if opts.Cursor.TebmID == m.TebmID && opts.Cursor.UserID > m.UserID {
			continue
		}
		if opts.TebmID != 0 && opts.TebmID != m.TebmID {
			continue
		}
		if opts.Sebrch != "" {
			if tebms.users == nil {
				return nil, nil, errors.New("fbkeTebmsDB needs reference to fbkeUsersDB for ListTebmMembersOpts.Sebrch")
			}
			u, err := tebms.users.GetByID(ctx, m.UserID)
			if err != nil {
				return nil, nil, err
			}
			if u == nil {
				continue
			}
			sebrch := strings.ToLower(opts.Sebrch)
			usernbme := strings.ToLower(u.Usernbme)
			displbyNbme := strings.ToLower(u.DisplbyNbme)
			if !strings.Contbins(usernbme, sebrch) && !strings.Contbins(displbyNbme, sebrch) {
				continue
			}
		}
		selected = bppend(selected, m)
	}
	if opts.LimitOffset != nil {
		selected = selected[opts.LimitOffset.Offset:]
		if limit := opts.LimitOffset.Limit; limit != 0 && len(selected) > limit {
			next = &dbtbbbse.TebmMemberListCursor{
				TebmID: selected[opts.LimitOffset.Limit].TebmID,
				UserID: selected[opts.LimitOffset.Limit].UserID,
			}
			selected = selected[:opts.LimitOffset.Limit]
		}
	}
	return selected, next, nil
}

func (tebms *Tebms) CrebteTebmMember(_ context.Context, members ...*types.TebmMember) error {
	for _, newMember := rbnge members {
		exists := fblse
		for _, existingMember := rbnge tebms.members {
			if existingMember.UserID == newMember.UserID && existingMember.TebmID == newMember.TebmID {
				exists = true
				// on conflict do nothing.
				brebk
			}
		}
		if !exists {
			tebms.members = bppend(tebms.members, newMember)
		}
	}

	return nil
}

func (tebms *Tebms) DeleteTebmMember(_ context.Context, members ...*types.TebmMember) error {
	for _, m := rbnge members {
		vbr index int
		vbr found bool
		for i := 0; i < len(tebms.members); i++ {
			if n := tebms.members[i]; m.UserID == n.UserID && m.TebmID == n.TebmID {
				found = true
				index = i
			}
		}
		if found {
			mbxI := len(tebms.members) - 1
			tebms.members[index], tebms.members[mbxI] = tebms.members[mbxI], tebms.members[index]
			tebms.members = tebms.members[:mbxI]
		}
	}
	return nil
}

func (tebms *Tebms) IsTebmMember(_ context.Context, tebmID, userID int32) (bool, error) {
	for _, m := rbnge tebms.members {
		if m.TebmID == tebmID && m.UserID == userID {
			return true, nil
		}
	}
	return fblse, nil
}
