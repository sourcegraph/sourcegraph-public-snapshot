package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type LocationResolver interface {
	Resource() *GitTreeEntryResolver
	Range() *rangeResolver
	URL(ctx context.Context) (string, error)
	CanonicalURL() (string, error)
}

type locationResolver struct {
	resource *GitTreeEntryResolver
	lspRange *lsp.Range
}

var _ LocationResolver = &locationResolver{}

func NewLocationResolver(resource *GitTreeEntryResolver, lspRange *lsp.Range) LocationResolver {
	return &locationResolver{
		resource: resource,
		lspRange: lspRange,
	}
}

func (r *locationResolver) Resource() *GitTreeEntryResolver { return r.resource }

func (r *locationResolver) Range() *rangeResolver {
	if r.lspRange == nil {
		return nil
	}
	return &rangeResolver{*r.lspRange}
}

func (r *locationResolver) URL(ctx context.Context) (string, error) {
	url, err := r.resource.URL(ctx)
	if err != nil {
		return "", err
	}
	return r.urlPath(url), nil
}

func (r *locationResolver) CanonicalURL() (string, error) {
	url, err := r.resource.CanonicalURL()
	if err != nil {
		return "", err
	}
	return r.urlPath(url), nil
}

func (r *locationResolver) urlPath(prefix string) string {
	url := prefix
	if r.lspRange != nil {
		url += "#L" + r.Range().urlFragment()
	}
	return url
}

type RangeResolver interface {
	Start() PositionResolver
	End() PositionResolver
}

type rangeResolver struct{ lspRange lsp.Range }

var _ RangeResolver = &rangeResolver{}

func NewRangeResolver(lspRange lsp.Range) RangeResolver {
	return &rangeResolver{
		lspRange: lspRange,
	}
}

func (r *rangeResolver) Start() PositionResolver { return r.start() }
func (r *rangeResolver) End() PositionResolver   { return r.end() }

func (r *rangeResolver) start() *positionResolver { return &positionResolver{r.lspRange.Start} }
func (r *rangeResolver) end() *positionResolver   { return &positionResolver{r.lspRange.End} }

func (r *rangeResolver) urlFragment() string {
	if r.lspRange.Start == r.lspRange.End {
		return r.start().urlFragment(false)
	}
	hasCharacter := r.lspRange.Start.Character != 0 || r.lspRange.End.Character != 0
	return r.start().urlFragment(hasCharacter) + "-" + r.end().urlFragment(hasCharacter)
}

type PositionResolver interface {
	Line() int32
	Character() int32
}

type positionResolver struct{ pos lsp.Position }

var _ PositionResolver = &positionResolver{}

func (r *positionResolver) Line() int32      { return int32(r.pos.Line) }
func (r *positionResolver) Character() int32 { return int32(r.pos.Character) }

func (r *positionResolver) urlFragment(forceIncludeCharacter bool) string {
	if !forceIncludeCharacter && r.pos.Character == 0 {
		return strconv.Itoa(r.pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Character+1)
}
