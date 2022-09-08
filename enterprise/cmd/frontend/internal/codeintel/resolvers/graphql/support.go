package graphql

import (
	"context"
	"path"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type searchBasedCodeIntelSupportType string

const (
	unsupported searchBasedCodeIntelSupportType = "UNSUPPORTED"
	basic       searchBasedCodeIntelSupportType = "BASIC"
)

type preciseCodeIntelSupportType string

const (
	native     preciseCodeIntelSupportType = "NATIVE"
	thirdParty preciseCodeIntelSupportType = "THIRD_PARTY"
	unknown    preciseCodeIntelSupportType = "UNKNOWN"
)

type searchBasedSupportResolver struct {
	language string
}

func NewSearchBasedCodeIntelResolver(language string) gql.SearchBasedSupportResolver {
	return &searchBasedSupportResolver{language}
}

func (r *searchBasedSupportResolver) SupportLevel() string {
	if r.language != "" {
		return string(basic)
	}
	return string(unsupported)
}

func (r *searchBasedSupportResolver) Language() string {
	return r.language
}

type preciseCodeIntelSupportResolver struct {
	indexers []gql.CodeIntelIndexerResolver
}

func NewPreciseCodeIntelSupportResolver(filepath string) gql.PreciseSupportResolver {
	return &preciseCodeIntelSupportResolver{
		indexers: languageToIndexer[path.Ext(filepath)],
	}
}

func NewPreciseCodeIntelSupportResolverFromIndexers(indexers []gql.CodeIntelIndexerResolver) gql.PreciseSupportResolver {
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

func (r *preciseCodeIntelSupportResolver) Indexers() *[]gql.CodeIntelIndexerResolver {
	if len(r.indexers) == 0 {
		return nil
	}
	return &r.indexers
}

func (r *Resolver) RequestLanguageSupport(ctx context.Context, args *gql.RequestLanguageSupportArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.observationContext.requestLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userID := int(actor.FromContext(ctx).UID)
	if userID == 0 {
		return nil, errors.Newf("language support requests only logged for authenticated users")
	}

	codeNavResolver := r.resolver.CodeNavResolver()
	if err := codeNavResolver.SetRequestLanguageSupport(ctx, userID, args.Language); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) RequestedLanguageSupport(ctx context.Context) (_ []string, err error) {
	ctx, _, endObservation := r.observationContext.requestedLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userID := int(actor.FromContext(ctx).UID)
	if userID == 0 {
		return nil, errors.Newf("language support requests only logged for authenticated users")
	}

	codeNavResolver := r.resolver.CodeNavResolver()
	return codeNavResolver.GetLanguagesRequestedBy(ctx, userID)
}
