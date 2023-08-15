package graphql

import (
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type CodeIntelIndexerResolver interface {
	Key() string
	Name() string
	URL() string
	ImageName() *string
}

type codeIntelIndexerResolver struct {
	indexer   uploadsshared.CodeIntelIndexer
	imageName string
}

func NewCodeIntelIndexerResolver(name, imageName string) CodeIntelIndexerResolver {
	return NewCodeIntelIndexerResolverFrom(uploadsshared.IndexerFromName(name), imageName)
}

func NewCodeIntelIndexerResolverFrom(indexer uploadsshared.CodeIntelIndexer, imageName string) CodeIntelIndexerResolver {
	return &codeIntelIndexerResolver{indexer: indexer, imageName: imageName}
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

func (r *codeIntelIndexerResolver) ImageName() *string {
	if r.imageName == "" {
		return nil
	}

	return &r.imageName
}
