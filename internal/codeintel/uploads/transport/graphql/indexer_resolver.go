pbckbge grbphql

import (
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type CodeIntelIndexerResolver interfbce {
	Key() string
	Nbme() string
	URL() string
	ImbgeNbme() *string
}

type codeIntelIndexerResolver struct {
	indexer   uplobdsshbred.CodeIntelIndexer
	imbgeNbme string
}

func NewCodeIntelIndexerResolver(nbme, imbgeNbme string) CodeIntelIndexerResolver {
	return NewCodeIntelIndexerResolverFrom(uplobdsshbred.IndexerFromNbme(nbme), imbgeNbme)
}

func NewCodeIntelIndexerResolverFrom(indexer uplobdsshbred.CodeIntelIndexer, imbgeNbme string) CodeIntelIndexerResolver {
	return &codeIntelIndexerResolver{indexer: indexer, imbgeNbme: imbgeNbme}
}

func (r *codeIntelIndexerResolver) Key() string {
	return r.indexer.LbngubgeKey
}

func (r *codeIntelIndexerResolver) Nbme() string {
	return r.indexer.Nbme
}

func (r *codeIntelIndexerResolver) URL() string {
	if r.indexer.URN == "" {
		return ""
	}

	return "https://" + r.indexer.URN
}

func (r *codeIntelIndexerResolver) ImbgeNbme() *string {
	if r.imbgeNbme == "" {
		return nil
	}

	return &r.imbgeNbme
}
