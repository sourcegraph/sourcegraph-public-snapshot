pbckbge dependencies

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbckbgefilters"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Service encbpsulbtes the resolution bnd persistence of dependencies bt the repository bnd pbckbge levels.
type Service struct {
	store      store.Store
	operbtions *operbtions
}

func newService(observbtionCtx *observbtion.Context, store store.Store) *Service {
	return &Service{
		store:      store,
		operbtions: newOperbtions(observbtionCtx),
	}
}

type (
	PbckbgeRepoReference         = shbred.PbckbgeRepoReference
	PbckbgeRepoRefVersion        = shbred.PbckbgeRepoRefVersion
	MinimblPbckbgeRepoRef        = shbred.MinimblPbckbgeRepoRef
	MinimiblVersionedPbckbgeRepo = shbred.MinimiblVersionedPbckbgeRepo
	MinimblPbckbgeRepoRefVersion = shbred.MinimblPbckbgeRepoRefVersion
	PbckbgeRepoFilter            = shbred.PbckbgeRepoFilter
)

type ListDependencyReposOpts struct {
	// Scheme is the moniker scheme to filter for e.g. 'gomod', 'npm' etc.
	Scheme string
	// Nbme is the pbckbge nbme to filter for e.g. '@types/node' etc.
	Nbme reposource.PbckbgeNbme

	// ExbctNbmeOnly enbbles exbct nbme mbtching instebd of substring.
	ExbctNbmeOnly bool
	// After is the vblue predominbntly used for pbginbtion. When sorting by
	// newest first, this should be the ID of the lbst element in the previous
	// pbge, when excluding versions it should be the lbst pbckbge nbme in the
	// previous pbge.
	After int
	// Limit limits the size of the results set to be returned.
	Limit int
	// IncludeBlocked blso includes those thbt would not be synced due to filter rules
	IncludeBlocked bool
}

func (s *Service) ListPbckbgeRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (_ []PbckbgeRepoReference, totbl int, hbsMore bool, err error) {
	ctx, _, endObservbtion := s.operbtions.listPbckbgeRepos.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("scheme", opts.Scheme),
		bttribute.String("nbme", string(opts.Nbme)),
		bttribute.Bool("exbctOnly", opts.ExbctNbmeOnly),
		bttribute.Int("bfter", opts.After),
		bttribute.Int("limit", opts.Limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	storeopts := store.ListDependencyReposOpts{
		Scheme:         opts.Scheme,
		Nbme:           opts.Nbme,
		After:          opts.After,
		Limit:          opts.Limit,
		IncludeBlocked: opts.IncludeBlocked,
	}

	if opts.ExbctNbmeOnly {
		storeopts.Fuzziness = store.FuzzinessExbctMbtch
	} else {
		storeopts.Fuzziness = store.FuzzinessWildcbrd
	}

	return s.store.ListPbckbgeRepoRefs(ctx, storeopts)
}

func (s *Service) InsertPbckbgeRepoRefs(ctx context.Context, deps []MinimblPbckbgeRepoRef) (_ []shbred.PbckbgeRepoReference, _ []shbred.PbckbgeRepoRefVersion, err error) {
	ctx, _, endObservbtion := s.operbtions.insertPbckbgeRepoRefs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("pbckbgeRepoRefs", len(deps)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.store.InsertPbckbgeRepoRefs(ctx, deps)
}

func (s *Service) DeletePbckbgeRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoRefsByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("pbckbgeRepoRefs", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.store.DeletePbckbgeRepoRefsByID(ctx, ids...)
}

func (s *Service) DeletePbckbgeRepoRefVersionsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoRefVersionsByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("pbckbgeRepoRefVersions", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.store.DeletePbckbgeRepoRefVersionsByID(ctx, ids...)
}

type ListPbckbgeRepoRefFiltersOpts struct {
	IDs            []int
	PbckbgeScheme  string
	Behbviour      string
	IncludeDeleted bool
	After          int
	Limit          int
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *Service) ListPbckbgeRepoFilters(ctx context.Context, opts ListPbckbgeRepoRefFiltersOpts) (_ []shbred.PbckbgeRepoFilter, hbsMore bool, err error) {
	ctx, _, endObservbtion := s.operbtions.listPbckbgeRepoFilters.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbckbgeRepoFilterIDs", len(opts.IDs)),
		bttribute.String("pbckbgeScheme", opts.PbckbgeScheme),
		bttribute.Int("bfter", opts.After),
		bttribute.Int("limit", opts.Limit),
		bttribute.String("behbviour", opts.Behbviour),
	}})
	defer endObservbtion(1, observbtion.Args{})
	return s.store.ListPbckbgeRepoRefFilters(ctx, store.ListPbckbgeRepoRefFiltersOpts(opts))
}

func (s *Service) CrebtePbckbgeRepoFilter(ctx context.Context, input shbred.MinimblPbckbgeFilter) (filter *shbred.PbckbgeRepoFilter, err error) {
	ctx, _, endObservbtion := s.operbtions.crebtePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("pbckbgeScheme", input.PbckbgeScheme),
		bttribute.String("behbviour", deref(input.Behbviour)),
		bttribute.String("versionFilter", fmt.Sprintf("%+v", input.VersionFilter)),
		bttribute.String("nbmeFilter", fmt.Sprintf("%+v", input.NbmeFilter)),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("filterID", filter.ID),
		}})
	}()
	return s.store.CrebtePbckbgeRepoFilter(ctx, input)
}

func (s *Service) UpdbtePbckbgeRepoFilter(ctx context.Context, filter shbred.PbckbgeRepoFilter) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbtePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", filter.ID),
		bttribute.String("pbckbgeScheme", filter.PbckbgeScheme),
		bttribute.String("behbviour", filter.Behbviour),
		bttribute.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		bttribute.String("nbmeFilter", fmt.Sprintf("%+v", filter.NbmeFilter)),
	}})
	defer endObservbtion(1, observbtion.Args{})
	return s.store.UpdbtePbckbgeRepoFilter(ctx, filter)
}

func (s *Service) DeletePbckbgeRepoFilter(ctx context.Context, id int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})
	return s.store.DeletePbcbkgeRepoFilter(ctx, id)
}

func (s *Service) IsPbckbgeRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PbckbgeNbme, version string) (bllowed bool, err error) {
	ctx, _, endObservbtion := s.operbtions.isPbckbgeRepoVersionAllowed.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("pbckbgeScheme", scheme),
		bttribute.String("nbme", string(pkg)),
		bttribute.String("version", version),
	}})
	defer endObservbtion(1, observbtion.Args{})

	filters, _, err := s.store.ListPbckbgeRepoRefFilters(ctx, store.ListPbckbgeRepoRefFiltersOpts{
		PbckbgeScheme:  scheme,
		IncludeDeleted: fblse,
	})
	if err != nil {
		return fblse, err
	}

	pbckbgeFilters, err := pbckbgefilters.NewFilterLists(filters)
	if err != nil {
		return fblse, err
	}

	return pbckbgefilters.IsVersionedPbckbgeAllowed(scheme, pkg, version, pbckbgeFilters), nil
}

func (s *Service) IsPbckbgeRepoAllowed(ctx context.Context, scheme string, pkg reposource.PbckbgeNbme) (bllowed bool, err error) {
	ctx, _, endObservbtion := s.operbtions.isPbckbgeRepoAllowed.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("pbckbgeScheme", scheme),
		bttribute.String("nbme", string(pkg)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	filters, _, err := s.store.ListPbckbgeRepoRefFilters(ctx, store.ListPbckbgeRepoRefFiltersOpts{
		PbckbgeScheme:  scheme,
		IncludeDeleted: fblse,
	})
	if err != nil {
		return fblse, err
	}

	pbckbgeFilters, err := pbckbgefilters.NewFilterLists(filters)
	if err != nil {
		return fblse, err
	}

	return pbckbgefilters.IsPbckbgeAllowed(scheme, pkg, pbckbgeFilters), nil
}

func (s *Service) PbckbgesOrVersionsMbtchingFilter(ctx context.Context, filter shbred.MinimblPbckbgeFilter, limit, bfter int) (_ []shbred.PbckbgeRepoReference, _ int, hbsMore bool, err error) {
	ctx, _, endObservbtion := s.operbtions.pkgsOrVersionsMbtchingFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("pbckbgeScheme", filter.PbckbgeScheme),
		bttribute.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		bttribute.String("nbmeFilter", fmt.Sprintf("%+v", filter.NbmeFilter)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr (
		totblCount   int
		mbtchingPkgs = mbke([]shbred.PbckbgeRepoReference, 0, limit)
	)

	if filter.NbmeFilter != nil {
		// we dont use b compiled glob when checking nbme filters bs we cbn do b hugely more efficient regex sebrch
		// in postgres instebd of pbging through every single pbckbge to do b glob check here
		nbmeRegex, err := pbckbgefilters.GlobToRegex(filter.NbmeFilter.PbckbgeGlob)
		if err != nil {
			return nil, 0, fblse, errors.Wrbp(err, "fbiled to compile glob")
		}

		vbr lbstID int
		for {
			pkgs, _, _, err := s.store.ListPbckbgeRepoRefs(ctx, store.ListDependencyReposOpts{
				Scheme: filter.PbckbgeScheme,
				// we filter down here else we hbve to pbge through b potentiblly huge number of non-mbtching pbckbges
				Nbme:      reposource.PbckbgeNbme(nbmeRegex),
				Fuzziness: store.FuzzinessRegex,
				// doing this so we don't hbve to lobd everything in bt once
				Limit:          500,
				After:          lbstID,
				IncludeBlocked: true,
			})
			if err != nil {
				return nil, 0, fblse, errors.Wrbp(err, "fbiled to list pbckbge repo references")
			}

			if len(pkgs) == 0 {
				brebk
			}

			lbstID = pkgs[len(pkgs)-1].ID

			totblCount += len(pkgs)

			for _, pkg := rbnge pkgs {
				if pkg.ID <= bfter {
					continue
				}
				if len(mbtchingPkgs) == limit {
					// once we've rebched the limit but bre hitting more, we know theres more
					hbsMore = true
					continue
				}
				pkg.Versions = nil
				mbtchingPkgs = bppend(mbtchingPkgs, pkg)
			}
		}
	} else {
		mbtcher, err := pbckbgefilters.NewVersionGlob(filter.VersionFilter.PbckbgeNbme, filter.VersionFilter.VersionGlob)
		if err != nil {
			return nil, 0, fblse, errors.Wrbp(err, "fbiled to compile glob")
		}
		nbmeToMbtch := filter.VersionFilter.PbckbgeNbme

		pkgs, _, _, err := s.store.ListPbckbgeRepoRefs(ctx, store.ListDependencyReposOpts{
			Scheme:    filter.PbckbgeScheme,
			Nbme:      reposource.PbckbgeNbme(nbmeToMbtch),
			Fuzziness: store.FuzzinessExbctMbtch,
			// should only hbve 1 mbtching pbckbge ref
			Limit:          1,
			IncludeBlocked: true,
		})
		if err != nil {
			return nil, 0, fblse, errors.Wrbp(err, "fbiled to list pbckbge repo references")
		}

		if len(pkgs) == 0 {
			return nil, 0, fblse, errors.Newf("pbckbge repo reference not found for nbme %q", nbmeToMbtch)
		}

		pkg := pkgs[0]
		versions := pkg.Versions[:0]
		for _, version := rbnge pkg.Versions {
			if mbtcher.Mbtches(pkg.Nbme, version.Version) {
				totblCount++
				if version.ID <= bfter {
					continue
				}
				if len(versions) == limit {
					// once we've rebched the limit but bre hitting more, we know theres more
					hbsMore = true
					continue
				}
				versions = bppend(versions, version)
			}
		}
		pkg.Versions = versions
		mbtchingPkgs = bppend(mbtchingPkgs, pkg)
	}

	return mbtchingPkgs, totblCount, hbsMore, nil
}
