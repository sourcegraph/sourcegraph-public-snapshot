package local

import (
	"errors"
	"log"
	"sort"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var Units sourcegraph.UnitsServer = &units{}

type units struct{}

var _ sourcegraph.UnitsServer = (*units)(nil)

func (s *units) Get(ctx context.Context, unitSpec *sourcegraph.UnitSpec) (*unit.RepoSourceUnit, error) {
	if unitSpec.RepoSpec.URI == "" || unitSpec.CommitID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "UnitSpec URI and CommitID must be set")
	}
	us, err := store.GraphFromContext(ctx).Units(
		srcstore.ByUnits(unit.ID2{Type: unitSpec.UnitType, Name: unitSpec.Unit}),
		srcstore.ByCommitIDs(unitSpec.CommitID),
		srcstore.ByRepos(unitSpec.RepoSpec.URI),
	)
	if err != nil {
		return nil, err
	}
	if len(us) == 0 {
		return nil, errors.New("unit does not exist")
	}
	return unit.NewRepoSourceUnit(us[0])
}

func (s *units) List(ctx context.Context, opt *sourcegraph.UnitListOptions) (*sourcegraph.RepoSourceUnitList, error) {
	if opt == nil {
		opt = new(sourcegraph.UnitListOptions)
	}

	unitFilters := []srcstore.UnitFilter{}
	if opt.UnitType != "" && opt.Unit != "" {
		unitFilters = []srcstore.UnitFilter{srcstore.ByUnits(unit.ID2{Type: opt.UnitType, Name: opt.Unit})}
	}
	if q := strings.ToLower(opt.NameQuery); q != "" {
		unitFilters = append(unitFilters, srcstore.UnitFilterFunc(func(u *unit.SourceUnit) bool {
			return strings.Contains(strings.ToLower(u.Name), q)
		}))
	}
	hasRepoRevFilter := false
	if len(opt.RepoRevs) > 0 {
		vs := make([]srcstore.Version, 0, len(opt.RepoRevs))
		for _, repoRev := range opt.RepoRevs {
			repoURI, commitID := sourcegraph.ParseRepoAndCommitID(repoRev)
			if len(commitID) != 40 {
				repoRev := sourcegraph.RepoRevSpec{
					RepoSpec: sourcegraph.RepoSpec{URI: repoURI},
					Rev:      commitID,
				}
				if err := (&repos{}).resolveRepoRev(ctx, &repoRev); err != nil {
					log.Printf("In UnitsService.List, resolving repoRev entry %q failed: %s. (Skipping.)", repoRev, err)
					continue
				}
				commitID = string(repoRev.CommitID)
			}
			if commitID != "" {
				vs = append(vs, srcstore.Version{Repo: repoURI, CommitID: commitID})
			}
		}
		if len(vs) > 0 {
			hasRepoRevFilter = true
			unitFilters = append(unitFilters, srcstore.ByRepoCommitIDs(vs...))
		}
	}
	if !hasRepoRevFilter {
		return nil, grpc.Errorf(codes.InvalidArgument, "Units.List requires at least 1 RepoRevs entry to narrow scope")
	}

	units, err := store.GraphFromContext(ctx).Units(unitFilters...)

	if err != nil {
		return nil, err
	}

	// Apply limit and pagination.
	offset := opt.Offset()
	min := len(units)
	if x := opt.Offset() + opt.Limit(); x < min {
		min = x
	}
	repoSourceUnits := make([]*unit.RepoSourceUnit, min)
	for i := range repoSourceUnits {
		u, err := unit.NewRepoSourceUnit(units[i+offset])
		if err != nil {
			return nil, err
		}
		repoSourceUnits[i] = u
	}

	sortable := sortableRepoSourceUnits(repoSourceUnits)
	sort.Sort(sortable)

	return &sourcegraph.RepoSourceUnitList{Units: []*unit.RepoSourceUnit(sortable)}, nil
}

type sortableRepoSourceUnits []*unit.RepoSourceUnit

func (s sortableRepoSourceUnits) Len() int      { return len(s) }
func (s sortableRepoSourceUnits) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableRepoSourceUnits) Less(i, j int) bool {
	return s[i].Repo+s[i].CommitID+s[i].UnitType+s[i].Unit < s[j].Repo+s[j].CommitID+s[j].UnitType+s[j].Unit
}
