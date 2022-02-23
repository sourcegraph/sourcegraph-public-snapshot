package graphql

import (
	"context"
	"path"
	"strings"

	"github.com/gobwas/glob"
	"github.com/sourcegraph/go-ctags"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	indexerconsts "github.com/sourcegraph/sourcegraph/internal/codeintel/indexer-consts"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
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

type gitBlobCodeIntelInfoResolver struct {
	gitBlobMeta *gql.GitBlobCodeIntelInfoArgs
	errTracer   *observation.ErrCollector
}

func NewGitBlobCodeIntelInfoResolver(args *gql.GitBlobCodeIntelInfoArgs, errTracer *observation.ErrCollector) gql.GitBlobCodeIntelInfoResolver {
	return &gitBlobCodeIntelInfoResolver{
		gitBlobMeta: args,
		errTracer:   errTracer,
	}
}

func (r *gitBlobCodeIntelInfoResolver) Support(ctx context.Context) gql.CodeIntelSupportResolver {
	return NewCodeIntelSupportResolver(r.gitBlobMeta, r.errTracer)
}

func (r *gitBlobCodeIntelInfoResolver) LSIFUploads(ctx context.Context) (gql.LSIFUploadConnectionResolver, error) {
	return nil, nil
}

type codeIntelSupportResolver struct {
	gitBlobMeta *gql.GitBlobCodeIntelInfoArgs
	errTracer   *observation.ErrCollector
}

func NewCodeIntelSupportResolver(args *gql.GitBlobCodeIntelInfoArgs, errTracer *observation.ErrCollector) gql.CodeIntelSupportResolver {
	return &codeIntelSupportResolver{
		gitBlobMeta: args,
		errTracer:   errTracer,
	}
}

func (r *codeIntelSupportResolver) SearchBasedSupport(ctx context.Context) (_ gql.SearchBasedCodeIntelSupportResolver, err error) {
	defer r.errTracer.Collect(&err)

	mappings, err := symbols.DefaultClient.ListLanguageMappings(ctx, r.gitBlobMeta.Repo)
	if err != nil {
		return nil, err
	}

	for _, allowedLanguage := range ctags.SupportedLanguages {
		for _, pattern := range mappings[allowedLanguage] {
			compiled, err := glob.Compile(pattern)
			if err != nil {
				return nil, err
			}

			if compiled.Match(path.Base(r.gitBlobMeta.Path)) {
				return NewSearchBasedCodeIntelResolver(&allowedLanguage), nil
			}
		}
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
