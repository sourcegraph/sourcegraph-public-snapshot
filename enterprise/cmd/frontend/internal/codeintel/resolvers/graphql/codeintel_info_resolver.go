package graphql

import (
	"context"
	"path"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	indexerconsts "github.com/sourcegraph/sourcegraph/internal/codeintel/indexer-consts"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

type codeIntelSupportResolver struct {
	gitBlobMeta *gql.GitBlobCodeIntelInfoArgs
	resolver    resolvers.Resolver
	errTracer   *observation.ErrCollector
}

func NewCodeIntelSupportResolver(resolver resolvers.Resolver, args *gql.GitBlobCodeIntelInfoArgs, errTracer *observation.ErrCollector) gql.CodeIntelSupportResolver {
	return &codeIntelSupportResolver{
		gitBlobMeta: args,
		resolver:    resolver,
		errTracer:   errTracer,
	}
}

func (r *codeIntelSupportResolver) SearchBasedSupport(ctx context.Context) (_ gql.SearchBasedCodeIntelSupportResolver, err error) {
	var (
		ctagsSupported bool
		language       string
	)

	defer func() {
		r.errTracer.Collect(&err,
			log.String("codeIntelSupportResolver.field", "searchBasedSupport"),
			log.String("inferredLanguage", language),
			log.Bool("ctagsSupported", ctagsSupported))
	}()

	ctagsSupported, language, err = r.resolver.SupportedByCtags(ctx, r.gitBlobMeta.Path, r.gitBlobMeta.Repo)
	if err != nil {
		return nil, err
	}

	if ctagsSupported {
		return NewSearchBasedCodeIntelResolver(&language), nil
	}

	return NewSearchBasedCodeIntelResolver(nil), nil
}

func (r *codeIntelSupportResolver) PreciseSupport(ctx context.Context) (gql.PreciseCodeIntelSupportResolver, error) {
	return NewPreciseCodeIntelSupportResolver(r.gitBlobMeta.Path), nil
}

type searchBasedSupportResolver struct {
	language *string
}

func NewSearchBasedCodeIntelResolver(language *string) gql.SearchBasedCodeIntelSupportResolver {
	return &searchBasedSupportResolver{language}
}

func (r *searchBasedSupportResolver) SupportLevel() string {
	if r.language != nil && *r.language != "" {
		return string(basic)
	}
	return string(unsupported)
}

func (r *searchBasedSupportResolver) Language() *string {
	return r.language
}

type preciseCodeIntelSupportResolver struct {
	indexers []gql.CodeIntelIndexerResolver
}

func NewPreciseCodeIntelSupportResolver(filepath string) gql.PreciseCodeIntelSupportResolver {
	return &preciseCodeIntelSupportResolver{
		indexers: indexerconsts.LanguageToIndexer[path.Ext(filepath)],
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
