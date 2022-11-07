package graphql

import (
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
)

type PreciseSupportResolver interface {
	SupportLevel() string
	Indexers() *[]types.CodeIntelIndexerResolver
}

type preciseCodeIntelSupportType string

const (
	native     preciseCodeIntelSupportType = "NATIVE"
	thirdParty preciseCodeIntelSupportType = "THIRD_PARTY"
	unknown    preciseCodeIntelSupportType = "UNKNOWN"
)

type preciseCodeIntelSupportResolver struct {
	indexers []types.CodeIntelIndexerResolver
}

func NewPreciseCodeIntelSupportResolver(filepath string) PreciseSupportResolver {
	indexers := types.LanguageToIndexer[path.Ext(filepath)]

	resolvers := make([]types.CodeIntelIndexerResolver, 0, len(indexers))
	for _, indexer := range indexers {
		resolvers = append(resolvers, types.NewCodeIntelIndexerResolverFrom(indexer))
	}

	return &preciseCodeIntelSupportResolver{
		indexers: resolvers,
	}
}

func NewPreciseCodeIntelSupportResolverFromIndexers(indexers []types.CodeIntelIndexerResolver) PreciseSupportResolver {
	return &preciseCodeIntelSupportResolver{
		indexers: indexers,
	}
}

func (r *preciseCodeIntelSupportResolver) SupportLevel() string {
	// if the first indexer in a list is from us, consider native support
	nativeRecommendation := len(r.indexers) > 0 &&
		strings.HasPrefix(r.indexers[0].URL(), "https://github.com/sourcegraph")

	if nativeRecommendation {
		return string(native)
	} else if len(r.indexers) > 0 {
		return string(thirdParty)
	} else {
		return string(unknown)
	}
}

func (r *preciseCodeIntelSupportResolver) Indexers() *[]types.CodeIntelIndexerResolver {
	if len(r.indexers) == 0 {
		return nil
	}
	return &r.indexers
}
