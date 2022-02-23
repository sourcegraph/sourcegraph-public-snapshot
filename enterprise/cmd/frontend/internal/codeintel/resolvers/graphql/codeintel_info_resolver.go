package graphql

import (
	"context"
	"path"
	"strings"

	"github.com/gobwas/glob"
	"github.com/sourcegraph/go-ctags"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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

var (
	lsifNode = codeIntelIndexerResolver{
		name: "lsif-tsc",
		urn:  "github.com/sourcegraph/lsif-node",
	}
	msftNode = codeIntelIndexerResolver{
		name: "msft/lsif-node",
		urn:  "github.com/Microsoft/lsif-node",
	}
	lsifTypescript = codeIntelIndexerResolver{
		name: "lsif-typescript",
		urn:  "github.com/sourcegraph/lsif-typescript",
	}
	lsifJava = codeIntelIndexerResolver{
		name: "lsif-java",
		urn:  "github.com/sourcegraph/lsif-java",
	}
	msftJava = codeIntelIndexerResolver{
		name: "msft/lsif-java",
		urn:  "github.com/Microsoft/lsif-java",
	}
	lsifGo = codeIntelIndexerResolver{
		name: "lsif-go",
		urn:  "github.com/sourcegraph/lsif-go",
	}
	lsifClang = codeIntelIndexerResolver{
		name: "lsif-clang",
		urn:  "github.com/sourcegraph/lsif-clang",
	}
	lsifCpp = codeIntelIndexerResolver{
		name: "lsif-cpp",
		urn:  "github.com/sourcegraph/lsif-cpp",
	}
	lsifDart = codeIntelIndexerResolver{
		name: "lsif-dart",
		urn:  "github.com/sourcegraph/lsif-dart",
	}
	workivaDart = codeIntelIndexerResolver{
		name: "lsif_indexer",
		urn:  "github.com/Workiva/lsif_indexer",
	}
	hieLsif = codeIntelIndexerResolver{
		name: "hie-lsif",
		urn:  "github.com/mpickering/hie-lsif",
	}
	lsifJsonnet = codeIntelIndexerResolver{
		name: "lsif-jsonnet",
		urn:  "github.com/sourcegraph/lsif-jsonnet",
	}
	lsifOcaml = codeIntelIndexerResolver{
		name: "lsif-ocaml",
		urn:  "github.com/rvantonder/lsif-ocaml",
	}
	lsifPy = codeIntelIndexerResolver{
		name: "lsif-py",
		urn:  "github.com/sourcegraph/lsif-py",
	}
	rustAnalyzer = codeIntelIndexerResolver{
		name: "rust-analyzer",
		urn:  "github.com/rust-analyzer/rust-analyzer",
	}
	lsifPhp = codeIntelIndexerResolver{
		name: "lsif-php",
		urn:  "github.com/davidrjenni/lsif-php",
	}
	lsifTerraform = codeIntelIndexerResolver{
		name: "lsif-terraform",
		urn:  "github.com/juliosueiras/lsif-terraform",
	}
)

var languageToIndexer = map[string][]gql.CodeIntelIndexerResolver{
	".go":      {&lsifGo},
	".java":    {&lsifJava, &msftJava},
	".kt":      {&lsifJava},
	".scala":   {&lsifJava},
	".js":      {&lsifTypescript, &lsifNode, &msftNode},
	".jsx":     {&lsifTypescript, &lsifNode, &msftNode},
	".ts":      {&lsifTypescript, &lsifNode, &msftNode},
	".tsx":     {&lsifTypescript, &lsifNode, &msftNode},
	".dart":    {&workivaDart, &lsifDart},
	".c":       {&lsifClang, &lsifCpp},
	".cc":      {&lsifClang, &lsifCpp},
	".cpp":     {&lsifClang, &lsifCpp},
	".cxx":     {&lsifClang, &lsifCpp},
	".h":       {&lsifClang, &lsifCpp},
	".hs":      {&hieLsif},
	".jsonnet": {&lsifJsonnet},
	".py":      {&lsifPy},
	".ml":      {&lsifOcaml},
	".rs":      {&rustAnalyzer},
	".php":     {&lsifPhp},
	".tf":      {&lsifTerraform},
}

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
		indexers: languageToIndexer[path.Ext(filepath)],
	}
}

func (r *preciseCodeIntelSupportResolver) SupportLevel() string {
	var hasNative bool
	for _, indexer := range r.indexers {
		if strings.HasPrefix(indexer.URL(), "https://github.com/sourcegraph") {
			hasNative = true
			break
		}
	}

	if hasNative {
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

type codeIntelIndexerResolver struct {
	name, urn string
}

func (r *codeIntelIndexerResolver) Name() string {
	return r.name
}

func (r *codeIntelIndexerResolver) URL() string {
	return "https://" + r.urn
}
