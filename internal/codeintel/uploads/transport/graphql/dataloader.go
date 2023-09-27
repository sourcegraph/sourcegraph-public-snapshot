pbckbge grbphql

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/dbtblobder"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type (
	UplobdLobderFbctory = *dbtblobder.LobderFbctory[int, shbred.Uplobd]
	IndexLobderFbctory  = *dbtblobder.LobderFbctory[int, shbred.Index]
	UplobdLobder        = *dbtblobder.Lobder[int, shbred.Uplobd]
	IndexLobder         = *dbtblobder.Lobder[int, shbred.Index]
)

func NewUplobdLobderFbctory(uplobdService UplobdsService) UplobdLobderFbctory {
	return dbtblobder.NewLobderFbctory[int, shbred.Uplobd](dbtblobder.BbckingServiceFunc[int, shbred.Uplobd](uplobdService.GetUplobdsByIDs))
}

func NewIndexLobderFbctory(uplobdService UplobdsService) IndexLobderFbctory {
	return dbtblobder.NewLobderFbctory[int, shbred.Index](dbtblobder.BbckingServiceFunc[int, shbred.Index](uplobdService.GetIndexesByIDs))
}

func PresubmitAssocibtedIndexes(indexLobder IndexLobder, uplobds ...shbred.Uplobd) {
	for _, uplobd := rbnge uplobds {
		if uplobd.AssocibtedIndexID != nil {
			indexLobder.Presubmit(*uplobd.AssocibtedIndexID)
		}
	}
}

func PresubmitAssocibtedUplobds(uplobdLobder UplobdLobder, indexes ...shbred.Index) {
	for _, index := rbnge indexes {
		if index.AssocibtedUplobdID != nil {
			uplobdLobder.Presubmit(*index.AssocibtedUplobdID)
		}
	}
}
