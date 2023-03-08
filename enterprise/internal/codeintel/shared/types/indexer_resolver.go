package types

import "strings"

type CodeIntelIndexerResolver interface {
	Key() string
	Name() string
	URL() string
	ImageName() *string
}

type codeIntelIndexerResolver struct {
	indexer   CodeIntelIndexer
	imageName string
}

func NewCodeIntelIndexerResolver(name, imageName string) CodeIntelIndexerResolver {
	return NewCodeIntelIndexerResolverFrom(indexerFromName(name), imageName)
}

func indexerFromName(name string) CodeIntelIndexer {
	// drop the Docker image tag if one exists
	name = strings.Split(name, "@sha256:")[0]
	name = strings.Split(name, ":")[0]

	if indexer, ok := imageToIndexer[name]; ok {
		return indexer
	}

	for _, indexer := range allIndexers {
		if indexer.Name == name {
			return indexer
		}
	}

	return CodeIntelIndexer{Name: name}
}

func NewCodeIntelIndexerResolverFrom(indexer CodeIntelIndexer, imageName string) CodeIntelIndexerResolver {
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
