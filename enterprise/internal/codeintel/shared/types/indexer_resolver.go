package types

type CodeIntelIndexerResolver interface {
	Name() string
	URL() string
}

type codeIntelIndexerResolver struct {
	indexer CodeIntelIndexer
}

func NewIndexerResolver(indexerName string) CodeIntelIndexerResolver {
	for _, indexer := range AllIndexers {
		if indexer.Name == indexerName {
			return NewCodeIntelIndexerResolverFrom(indexer)
		}
	}

	return NewCodeIntelIndexerResolver(indexerName)
}

func NewCodeIntelIndexerResolver(name string) CodeIntelIndexerResolver {
	return NewCodeIntelIndexerResolverFrom(CodeIntelIndexer{Name: name})
}

func NewCodeIntelIndexerResolverFrom(indexer CodeIntelIndexer) CodeIntelIndexerResolver {
	return &codeIntelIndexerResolver{indexer: indexer}
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
