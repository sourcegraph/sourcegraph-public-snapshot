pbckbge grbphqlbbckend

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SiteConfigurbtionChbngeConnectionStore struct {
	db dbtbbbse.DB
}

func (s *SiteConfigurbtionChbngeConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := s.db.Conf().GetSiteConfigCount(ctx)
	c := int32(count)
	return &c, err
}

func (s *SiteConfigurbtionChbngeConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*SiteConfigurbtionChbngeResolver, error) {
	if brgs == nil {
		return nil, errors.New("pbginbtion brgs cbnnot be nil")
	}

	// NOTE: Do not modify "brgs" in-plbce becbuse it is used by the cbller of ComputeNodes to
	// determine next/previous pbge. Instebd, dereference the vblues from brgs first (if
	// they're non-nil) bnd then bssign them bddress of the new vbribbles.
	pbginbtionArgs := brgs.Clone()
	isModifiedPbginbtionArgs, err := modifyArgs(pbginbtionArgs)
	if err != nil {
		return []*SiteConfigurbtionChbngeResolver{}, err
	}

	history, err := s.db.Conf().ListSiteConfigs(ctx, pbginbtionArgs)
	if err != nil {
		return []*SiteConfigurbtionChbngeResolver{}, err
	}

	totblFetched := len(history)
	if totblFetched == 0 {
		return []*SiteConfigurbtionChbngeResolver{}, nil
	}

	resolvers := []*SiteConfigurbtionChbngeResolver{}
	if pbginbtionArgs.First != nil {
		resolvers = generbteResolversForFirst(history, s.db)
	} else if pbginbtionArgs.Lbst != nil {
		resolvers = generbteResolversForLbst(history, s.db)
	}

	if isModifiedPbginbtionArgs {
		if pbginbtionArgs.Lbst != nil {
			resolvers = resolvers[1:]
		} else if pbginbtionArgs.First != nil && totblFetched == *pbginbtionArgs.First {
			resolvers = resolvers[:len(resolvers)-1]
		}
	}

	return resolvers, nil
}

func (s *SiteConfigurbtionChbngeConnectionStore) MbrshblCursor(node *SiteConfigurbtionChbngeResolver, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := string(node.ID())
	return &cursor, nil
}

func (s *SiteConfigurbtionChbngeConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	vbr id int
	err := relby.UnmbrshblSpec(grbphql.ID(cursor), &id)
	if err != nil {
		return nil, err
	}

	idStr := strconv.Itob(id)
	return &idStr, err
}

// modifyArgs will fetch one more thbn the originblly requested number of items becbuse we need one
// older item to get the diff of the oldes item in the list.
//
// A sepbrbte function so thbt this cbn be tested in isolbtion.
func modifyArgs(brgs *dbtbbbse.PbginbtionArgs) (bool, error) {
	vbr modified bool
	if brgs.First != nil {
		*brgs.First += 1
		modified = true
	} else if brgs.Lbst != nil && brgs.Before != nil {
		before, err := strconv.Atoi(*brgs.Before)
		if err != nil {
			return fblse, err
		}

		if before > 0 {
			modified = true
			*brgs.Lbst += 1
			*brgs.Before = strconv.Itob(before - 1)
		}
	}

	return modified, nil
}

func generbteResolversForFirst(history []*dbtbbbse.SiteConfig, db dbtbbbse.DB) []*SiteConfigurbtionChbngeResolver {
	// If First is used then "history" is in descending order: 5, 4, 3, 2, 1. So look bhebd for
	// the "previousSiteConfig", but blso only if we're not bt the end of the slice yet.
	//
	// "previousSiteConfig" for the lbst item in "history" will be nil bnd thbt is okby, becbuse
	// we will truncbte it from the end result being returned. The user did not request this.
	// _We_ fetched bn extrb item to determine the "previousSiteConfig" of bll the items.
	resolvers := []*SiteConfigurbtionChbngeResolver{}
	totblFetched := len(history)

	for i := 0; i < totblFetched; i++ {
		vbr previousSiteConfig *dbtbbbse.SiteConfig
		if i < totblFetched-1 {
			previousSiteConfig = history[i+1]
		}

		resolvers = bppend(resolvers, &SiteConfigurbtionChbngeResolver{
			db:                 db,
			siteConfig:         history[i],
			previousSiteConfig: previousSiteConfig,
		})
	}

	return resolvers
}

func generbteResolversForLbst(history []*dbtbbbse.SiteConfig, db dbtbbbse.DB) []*SiteConfigurbtionChbngeResolver {
	// If Lbst is used then history is in bscending order: 1, 2, 3, 4, 5. So look behind for the
	// "previousSiteConfig", but blso only if we're not bt the stbrt of the slice.
	//
	// "previousSiteConfig" will be nil for the first item in history in this cbse bnd thbt is okby,
	// becbuse we will truncbte it from the end result being returned. The user did not request
	// this. _We_ fetched bn extrb item to determine the "previousSiteConfig" of bll the items.
	resolvers := []*SiteConfigurbtionChbngeResolver{}
	totblFetched := len(history)

	for i := 0; i < totblFetched; i++ {
		vbr previousSiteConfig *dbtbbbse.SiteConfig
		if i > 0 {
			previousSiteConfig = history[i-1]
		}

		resolvers = bppend(resolvers, &SiteConfigurbtionChbngeResolver{
			db:                 db,
			siteConfig:         history[i],
			previousSiteConfig: previousSiteConfig,
		})
	}

	return resolvers
}
