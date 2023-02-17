package types

type CodeIntelIndexerResolver interface {
	Key() string
	Name() string
	URL() string
}

type codeIntelIndexerResolver struct {
	indexer CodeIntelIndexer
}

func NewCodeIntelIndexerResolver(name string) CodeIntelIndexerResolver {
	for _, indexer := range AllIndexers {
		if indexer.Name == name {
			return NewCodeIntelIndexerResolverFrom(indexer)
		}
	}

	return NewCodeIntelIndexerResolverFrom(CodeIntelIndexer{Name: name})
}

func NewCodeIntelIndexerResolverFrom(indexer CodeIntelIndexer) CodeIntelIndexerResolver {
	return &codeIntelIndexerResolver{indexer: indexer}
}

func (r *codeIntelIndexerResolver) Key() string {
	return r.indexer.LanguageKey
}

func (r *codeIntelIndexerResolver) Name() string {
	return r.indexer.Name
}

func (r *codeIntelIndexerResolver) URL() string {
	if r.indexer.URN == "" {
		return ""
	}

	return "https://" + r.indexer.URN
}
