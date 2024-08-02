package graphql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
)

func newLocationConnectionResolver(locations []shared.UploadLocation, cursor *string, locationResolver *gitresolvers.CachedLocationResolver) resolverstubs.LocationConnectionResolver {
	return resolverstubs.NewLazyConnectionResolver(func(ctx context.Context) ([]resolverstubs.LocationResolver, error) {
		return resolveLocations(ctx, locationResolver, locations)
	}, encodeCursor(cursor))
}

// resolveLocations creates a slide of LocationResolvers for the given list of adjusted locations. The
// resulting list may be smaller than the input list as any locations with a commit not known by
// gitserver will be skipped.
func resolveLocations(ctx context.Context, locationResolver *gitresolvers.CachedLocationResolver, locations []shared.UploadLocation) ([]resolverstubs.LocationResolver, error) {
	resolvedLocations := make([]resolverstubs.LocationResolver, 0, len(locations))
	for i := range locations {
		resolver, err := resolveLocation(ctx, locationResolver, locations[i])
		if err != nil {
			return nil, err
		}
		if resolver == nil {
			continue
		}

		resolvedLocations = append(resolvedLocations, resolver)
	}

	return resolvedLocations, nil
}

// resolveLocation creates a LocationResolver for the given adjusted location. This function may return a
// nil resolver if the location's commit is not known by gitserver.
func resolveLocation(ctx context.Context, locationResolver *gitresolvers.CachedLocationResolver, location shared.UploadLocation) (resolverstubs.LocationResolver, error) {
	treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.Upload.RepositoryID), location.TargetCommit, location.Path.RawValue(), false)
	if err != nil || treeResolver == nil {
		return nil, err
	}

	return newLocationResolver(treeResolver, location.TargetRange.ToSCIPRange()), nil
}

type locationResolver struct {
	resource resolverstubs.GitTreeEntryResolver
	range_   scip.Range
}

func newLocationResolver(resource resolverstubs.GitTreeEntryResolver, range_ scip.Range) resolverstubs.LocationResolver {
	return &locationResolver{
		resource: resource,
		range_:   range_,
	}
}

func (r *locationResolver) Resource() resolverstubs.GitTreeEntryResolver { return r.resource }

func (r *locationResolver) Range() resolverstubs.RangeResolver {
	return r.rangeInternal()
}

func (r *locationResolver) rangeInternal() *rangeResolver {
	return &rangeResolver{r.range_}
}

func (r *locationResolver) URL(ctx context.Context) (string, error) {
	return r.urlPath(r.resource.URL()), nil
}

func (r *locationResolver) CanonicalURL() string {
	return r.urlPath(r.resource.URL())
}

func (r *locationResolver) urlPath(prefix string) string {
	return prefix + "?L" + r.rangeInternal().urlFragment()
}

type rangeResolver struct{ range_ scip.Range }

func newRangeResolver(range_ scip.Range) resolverstubs.RangeResolver {
	return &rangeResolver{
		range_: range_,
	}
}

func (r *rangeResolver) Start() resolverstubs.PositionResolver { return r.start() }
func (r *rangeResolver) End() resolverstubs.PositionResolver   { return r.end() }

func (r *rangeResolver) start() *positionResolver { return &positionResolver{r.range_.Start} }
func (r *rangeResolver) end() *positionResolver   { return &positionResolver{r.range_.End} }

func (r *rangeResolver) urlFragment() string {
	if r.range_.Start == r.range_.End {
		return r.start().urlFragment(false)
	}
	hasCharacter := r.range_.Start.Character != 0 || r.range_.End.Character != 0
	return r.start().urlFragment(hasCharacter) + "-" + r.end().urlFragment(hasCharacter)
}

type positionResolver struct{ pos scip.Position }

func (r *positionResolver) Line() int32      { return r.pos.Line }
func (r *positionResolver) Character() int32 { return r.pos.Character }

func (r *positionResolver) urlFragment(forceIncludeCharacter bool) string {
	if !forceIncludeCharacter && r.pos.Character == 0 {
		return strconv.Itoa(int(r.pos.Line + 1))
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Character+1)
}
