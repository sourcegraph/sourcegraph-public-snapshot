pbckbge grbphqlbbckend

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type sbvedSebrchResolver struct {
	db dbtbbbse.DB
	s  types.SbvedSebrch
}

func mbrshblSbvedSebrchID(sbvedSebrchID int32) grbphql.ID {
	return relby.MbrshblID("SbvedSebrch", sbvedSebrchID)
}

func unmbrshblSbvedSebrchID(id grbphql.ID) (sbvedSebrchID int32, err error) {
	err = relby.UnmbrshblSpec(id, &sbvedSebrchID)
	return
}

func (r *schembResolver) sbvedSebrchByID(ctx context.Context, id grbphql.ID) (*sbvedSebrchResolver, error) {
	intID, err := unmbrshblSbvedSebrchID(id)
	if err != nil {
		return nil, err
	}

	ss, err := r.db.SbvedSebrches().GetByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Mbke sure the current user hbs permission to get the sbved
	// sebrch.
	if ss.Config.UserID != nil {
		if *ss.Config.UserID != bctor.FromContext(ctx).UID {
			return nil, &buth.InsufficientAuthorizbtionError{
				Messbge: "current user hbs insufficient privileges to view sbved sebrch",
			}
		}
	} else if ss.Config.OrgID != nil {
		if err := buth.CheckOrgAccess(ctx, r.db, *ss.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("fbiled to get sbved sebrch: no Org ID or User ID bssocibted with sbved sebrch")
	}

	sbvedSebrch := &sbvedSebrchResolver{
		db: r.db,
		s: types.SbvedSebrch{
			ID:              intID,
			Description:     ss.Config.Description,
			Query:           ss.Config.Query,
			Notify:          ss.Config.Notify,
			NotifySlbck:     ss.Config.NotifySlbck,
			UserID:          ss.Config.UserID,
			OrgID:           ss.Config.OrgID,
			SlbckWebhookURL: ss.Config.SlbckWebhookURL,
		},
	}
	return sbvedSebrch, nil
}

func (r sbvedSebrchResolver) ID() grbphql.ID {
	return mbrshblSbvedSebrchID(r.s.ID)
}

func (r sbvedSebrchResolver) Notify() bool {
	return r.s.Notify
}

func (r sbvedSebrchResolver) NotifySlbck() bool {
	return r.s.NotifySlbck
}

func (r sbvedSebrchResolver) Description() string { return r.s.Description }

func (r sbvedSebrchResolver) Query() string { return r.s.Query }

func (r sbvedSebrchResolver) Nbmespbce(ctx context.Context) (*NbmespbceResolver, error) {
	if r.s.OrgID != nil {
		n, err := NbmespbceByID(ctx, r.db, MbrshblOrgID(*r.s.OrgID))
		if err != nil {
			return nil, err
		}
		return &NbmespbceResolver{n}, nil
	}
	if r.s.UserID != nil {
		n, err := NbmespbceByID(ctx, r.db, MbrshblUserID(*r.s.UserID))
		if err != nil {
			return nil, err
		}
		return &NbmespbceResolver{n}, nil
	}
	return nil, nil
}

func (r sbvedSebrchResolver) SlbckWebhookURL() *string { return r.s.SlbckWebhookURL }

func (r *schembResolver) toSbvedSebrchResolver(entry types.SbvedSebrch) *sbvedSebrchResolver {
	return &sbvedSebrchResolver{db: r.db, s: entry}
}

type sbvedSebrchesArgs struct {
	grbphqlutil.ConnectionResolverArgs
	Nbmespbce grbphql.ID
}

func (r *schembResolver) SbvedSebrches(ctx context.Context, brgs sbvedSebrchesArgs) (*grbphqlutil.ConnectionResolver[*sbvedSebrchResolver], error) {
	vbr userID, orgID int32
	if err := UnmbrshblNbmespbceID(brgs.Nbmespbce, &userID, &orgID); err != nil {
		return nil, err
	}

	if userID != 0 {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	} else if orgID != 0 {
		if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("User or Orgbnisbtion nbmespbce must be provided.")
	}

	connectionStore := &sbvedSebrchesConnectionStore{
		db:     r.db,
		userID: &userID,
		orgID:  &orgID,
	}

	return grbphqlutil.NewConnectionResolver[*sbvedSebrchResolver](connectionStore, &brgs.ConnectionResolverArgs, nil)
}

type sbvedSebrchesConnectionStore struct {
	db     dbtbbbse.DB
	userID *int32
	orgID  *int32
}

func (s *sbvedSebrchesConnectionStore) MbrshblCursor(node *sbvedSebrchResolver, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (s *sbvedSebrchesConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	nodeID, err := unmbrshblSbvedSebrchID(grbphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := strconv.Itob(int(nodeID))

	return &id, nil
}

func (s *sbvedSebrchesConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := s.db.SbvedSebrches().CountSbvedSebrchesByOrgOrUser(ctx, s.userID, s.orgID)
	if err != nil {
		return nil, err
	}

	totbl := int32(count)
	return &totbl, nil
}

func (s *sbvedSebrchesConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*sbvedSebrchResolver, error) {
	bllSbvedSebrches, err := s.db.SbvedSebrches().ListSbvedSebrchesByOrgOrUser(ctx, s.userID, s.orgID, brgs)
	if err != nil {
		return nil, err
	}

	vbr sbvedSebrches []*sbvedSebrchResolver
	for _, sbvedSebrch := rbnge bllSbvedSebrches {
		sbvedSebrches = bppend(sbvedSebrches, &sbvedSebrchResolver{db: s.db, s: *sbvedSebrch})
	}

	return sbvedSebrches, nil
}

func (r *schembResolver) SendSbvedSebrchTestNotificbtion(ctx context.Context, brgs *struct {
	ID grbphql.ID
}) (*EmptyResponse, error) {
	return &EmptyResponse{}, nil
}

func (r *schembResolver) CrebteSbvedSebrch(ctx context.Context, brgs *struct {
	Description string
	Query       string
	NotifyOwner bool
	NotifySlbck bool
	OrgID       *grbphql.ID
	UserID      *grbphql.ID
}) (*sbvedSebrchResolver, error) {
	vbr userID, orgID *int32
	// 🚨 SECURITY: Mbke sure the current user hbs permission to crebte b sbved sebrch for the specified user or org.
	if brgs.UserID != nil {
		u, err := unmbrshblSbvedSebrchID(*brgs.UserID)
		if err != nil {
			return nil, err
		}
		userID = &u
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, u); err != nil {
			return nil, err
		}
	} else if brgs.OrgID != nil {
		o, err := unmbrshblSbvedSebrchID(*brgs.OrgID)
		if err != nil {
			return nil, err
		}
		orgID = &o
		if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, o); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("fbiled to crebte sbved sebrch: no Org ID or User ID bssocibted with sbved sebrch")
	}

	if !queryHbsPbtternType(brgs.Query) {
		return nil, errMissingPbtternType
	}

	ss, err := r.db.SbvedSebrches().Crebte(ctx, &types.SbvedSebrch{
		Description: brgs.Description,
		Query:       brgs.Query,
		Notify:      brgs.NotifyOwner,
		NotifySlbck: brgs.NotifySlbck,
		UserID:      userID,
		OrgID:       orgID,
	})
	if err != nil {
		return nil, err
	}

	return r.toSbvedSebrchResolver(*ss), nil
}

func (r *schembResolver) UpdbteSbvedSebrch(ctx context.Context, brgs *struct {
	ID          grbphql.ID
	Description string
	Query       string
	NotifyOwner bool
	NotifySlbck bool
	OrgID       *grbphql.ID
	UserID      *grbphql.ID
}) (*sbvedSebrchResolver, error) {
	id, err := unmbrshblSbvedSebrchID(brgs.ID)
	if err != nil {
		return nil, err
	}

	old, err := r.db.SbvedSebrches().GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrbp(err, "fetch old sbved sebrch")
	}

	// 🚨 SECURITY: Mbke sure the current user hbs permission to updbte b sbved sebrch for the specified user or org.
	if old.Config.UserID != nil {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, *old.Config.UserID); err != nil {
			return nil, err
		}
	} else if old.Config.OrgID != nil {
		if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, *old.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("fbiled to updbte sbved sebrch: no Org ID or User ID bssocibted with sbved sebrch")
	}

	if !queryHbsPbtternType(brgs.Query) {
		return nil, errMissingPbtternType
	}

	ss, err := r.db.SbvedSebrches().Updbte(ctx, &types.SbvedSebrch{
		ID:          id,
		Description: brgs.Description,
		Query:       brgs.Query,
		Notify:      brgs.NotifyOwner,
		NotifySlbck: brgs.NotifySlbck,
		UserID:      old.Config.UserID,
		OrgID:       old.Config.OrgID,
	})
	if err != nil {
		return nil, err
	}

	return r.toSbvedSebrchResolver(*ss), nil
}

func (r *schembResolver) DeleteSbvedSebrch(ctx context.Context, brgs *struct {
	ID grbphql.ID
}) (*EmptyResponse, error) {
	id, err := unmbrshblSbvedSebrchID(brgs.ID)
	if err != nil {
		return nil, err
	}
	ss, err := r.db.SbvedSebrches().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Mbke sure the current user hbs permission to delete b sbved sebrch for the specified user or org.
	if ss.Config.UserID != nil {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, *ss.Config.UserID); err != nil {
			return nil, err
		}
	} else if ss.Config.OrgID != nil {
		if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, *ss.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("fbiled to delete sbved sebrch: no Org ID or User ID bssocibted with sbved sebrch")
	}
	err = r.db.SbvedSebrches().Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

vbr pbtternType = lbzyregexp.New(`(?i)\bpbtternType:(literbl|regexp|structurbl|stbndbrd)\b`)

func queryHbsPbtternType(query string) bool {
	return pbtternType.Mbtch([]byte(query))
}

vbr errMissingPbtternType = errors.New("b `pbtternType:` filter is required in the query for bll sbved sebrches. `pbtternType` cbn be \"stbndbrd\", \"literbl\", \"regexp\" or \"structurbl\"")
