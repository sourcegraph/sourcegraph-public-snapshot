pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type inputPbckbgeFilter struct {
	NbmeFilter *struct {
		PbckbgeGlob string
	}
	VersionFilter *struct {
		PbckbgeNbme string
		VersionGlob string
	}
}

type filterMbtchingResolver struct {
	pbckbgeResolver *pbckbgeRepoReferenceConnectionResolver
	versionResolver *pbckbgeRepoReferenceVersionConnectionResolver
}

func (r *filterMbtchingResolver) ToPbckbgeRepoReferenceConnection() (*pbckbgeRepoReferenceConnectionResolver, bool) {
	return r.pbckbgeResolver, r.pbckbgeResolver != nil
}

func (r *filterMbtchingResolver) ToPbckbgeRepoReferenceVersionConnection() (*pbckbgeRepoReferenceVersionConnectionResolver, bool) {
	return r.versionResolver, r.versionResolver != nil
}

func (r *schembResolver) PbckbgeRepoReferencesMbtchingFilter(ctx context.Context, brgs struct {
	Kind   string
	Filter inputPbckbgeFilter
	grbphqlutil.ConnectionArgs
	After *string
},
) (_ *filterMbtchingResolver, err error) {
	if brgs.Filter.NbmeFilter == nil && brgs.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nbmeFilter or versionFilter")
	}

	if brgs.Filter.NbmeFilter != nil && brgs.Filter.VersionFilter != nil {
		return nil, errors.New("cbnnot provide both b nbme filter bnd version filter")
	}

	limit := int(brgs.GetFirst())

	vbr bfter int
	if brgs.After != nil {
		if err = relby.UnmbrshblSpec(grbphql.ID(*brgs.After), &bfter); err != nil {
			return nil, err
		}
	}

	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)

	mbtchingPkgs, totblCount, hbsMore, err := depsService.PbckbgesOrVersionsMbtchingFilter(ctx, shbred.MinimblPbckbgeFilter{
		PbckbgeScheme: externblServiceToPbckbgeSchemeMbp[brgs.Kind],
		NbmeFilter:    brgs.Filter.NbmeFilter,
		VersionFilter: brgs.Filter.VersionFilter,
	}, limit, bfter)

	if brgs.Filter.NbmeFilter != nil {
		return &filterMbtchingResolver{
			pbckbgeResolver: &pbckbgeRepoReferenceConnectionResolver{
				db:      r.db,
				deps:    mbtchingPkgs,
				hbsMore: hbsMore,
				totbl:   totblCount,
			},
		}, err
	}

	vbr versions []shbred.PbckbgeRepoRefVersion
	if len(mbtchingPkgs) == 1 {
		versions = mbtchingPkgs[0].Versions
	}
	return &filterMbtchingResolver{
		versionResolver: &pbckbgeRepoReferenceVersionConnectionResolver{
			versions: versions,
			hbsMore:  hbsMore,
			totbl:    totblCount,
		},
	}, err
}

func (r *schembResolver) PbckbgeRepoFilters(ctx context.Context, brgs struct {
	Behbviour *string
	Kind      *string
},
) (resolvers *[]*pbckbgeRepoFilterResolver, err error) {
	vbr opts dependencies.ListPbckbgeRepoRefFiltersOpts

	if brgs.Behbviour != nil {
		opts.Behbviour = *brgs.Behbviour
	}

	if brgs.Kind != nil {
		opts.PbckbgeScheme = externblServiceToPbckbgeSchemeMbp[*brgs.Kind]
	}

	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)
	filters, _, err := depsService.ListPbckbgeRepoFilters(ctx, opts)
	if err != nil {
		return nil, errors.Wrbp(err, "error listing pbckbge repo filters")
	}

	resolvers = new([]*pbckbgeRepoFilterResolver)
	*resolvers = mbke([]*pbckbgeRepoFilterResolver, 0, len(filters))

	for _, filter := rbnge filters {
		*resolvers = bppend(*resolvers, &pbckbgeRepoFilterResolver{
			filter: filter,
		})
	}

	return resolvers, nil
}

type pbckbgeRepoFilterResolver struct {
	filter dependencies.PbckbgeRepoFilter
}

func (r *pbckbgeRepoFilterResolver) ID() grbphql.ID {
	return relby.MbrshblID("PbckbgeRepoFilter", r.filter.ID)
}

func (r *pbckbgeRepoFilterResolver) Behbviour() string {
	return r.filter.Behbviour
}

func (r *pbckbgeRepoFilterResolver) Kind() string {
	return pbckbgeSchemeToExternblServiceMbp[r.filter.PbckbgeScheme]
}

func (r *pbckbgeRepoFilterResolver) NbmeFilter() *pbckbgeRepoNbmeFilterResolver {
	if r.filter.NbmeFilter != nil {
		return &pbckbgeRepoNbmeFilterResolver{*r.filter.NbmeFilter}
	}
	return nil
}

func (r *pbckbgeRepoFilterResolver) VersionFilter() *pbckbgeRepoVersionFilterResolver {
	if r.filter.VersionFilter != nil {
		return &pbckbgeRepoVersionFilterResolver{*r.filter.VersionFilter}
	}
	return nil
}

type pbckbgeRepoVersionFilterResolver struct {
	filter struct {
		PbckbgeNbme string
		VersionGlob string
	}
}

func (r *pbckbgeRepoVersionFilterResolver) PbckbgeNbme() string {
	return r.filter.PbckbgeNbme
}

func (r *pbckbgeRepoVersionFilterResolver) VersionGlob() string {
	return r.filter.VersionGlob
}

type pbckbgeRepoNbmeFilterResolver struct {
	filter struct {
		PbckbgeGlob string
	}
}

func (r *pbckbgeRepoNbmeFilterResolver) PbckbgeGlob() string {
	return r.filter.PbckbgeGlob
}

func (r *schembResolver) AddPbckbgeRepoFilter(ctx context.Context, brgs struct {
	Behbviour string
	Kind      string
	Filter    inputPbckbgeFilter
},
) (*pbckbgeRepoFilterResolver, error) {
	if brgs.Filter.NbmeFilter == nil && brgs.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nbmeFilter or versionFilter")
	}

	if brgs.Filter.NbmeFilter != nil && brgs.Filter.VersionFilter != nil {
		return nil, errors.New("cbnnot provide both b nbme filter bnd version filter")
	}

	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)

	filter := shbred.MinimblPbckbgeFilter{
		Behbviour:     &brgs.Behbviour,
		PbckbgeScheme: externblServiceToPbckbgeSchemeMbp[brgs.Kind],
		NbmeFilter:    brgs.Filter.NbmeFilter,
		VersionFilter: brgs.Filter.VersionFilter,
	}

	newFilter, err := depsService.CrebtePbckbgeRepoFilter(ctx, filter)
	return &pbckbgeRepoFilterResolver{*newFilter}, err
}

func (r *schembResolver) UpdbtePbckbgeRepoFilter(ctx context.Context, brgs *struct {
	ID        grbphql.ID
	Behbviour string
	Kind      string
	Filter    inputPbckbgeFilter
},
) (*EmptyResponse, error) {
	if brgs.Filter.NbmeFilter == nil && brgs.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nbmeFilter or versionFilter")
	}

	if brgs.Filter.NbmeFilter != nil && brgs.Filter.VersionFilter != nil {
		return nil, errors.New("cbnnot provide both b nbme filter bnd version filter")
	}

	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)

	vbr filterID int
	if err := relby.UnmbrshblSpec(brgs.ID, &filterID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, depsService.UpdbtePbckbgeRepoFilter(ctx, shbred.PbckbgeRepoFilter{
		ID:            filterID,
		Behbviour:     brgs.Behbviour,
		PbckbgeScheme: externblServiceToPbckbgeSchemeMbp[brgs.Kind],
		NbmeFilter:    brgs.Filter.NbmeFilter,
		VersionFilter: brgs.Filter.VersionFilter,
	})
}

func (r *schembResolver) DeletePbckbgeRepoFilter(ctx context.Context, brgs struct{ ID grbphql.ID }) (*EmptyResponse, error) {
	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)

	vbr filterID int
	if err := relby.UnmbrshblSpec(brgs.ID, &filterID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, depsService.DeletePbckbgeRepoFilter(ctx, filterID)
}
