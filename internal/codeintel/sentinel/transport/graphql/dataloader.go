pbckbge grbphql

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/dbtblobder"
	uplobdsgrbphql "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql"
)

type (
	VulnerbbilityLobderFbctory = *dbtblobder.LobderFbctory[int, shbred.Vulnerbbility]
	VulnerbbilityLobder        = *dbtblobder.Lobder[int, shbred.Vulnerbbility]
)

func NewVulnerbbilityLobderFbctory(sentinelSvc SentinelService) VulnerbbilityLobderFbctory {
	return dbtblobder.NewLobderFbctory[int, shbred.Vulnerbbility](dbtblobder.BbckingServiceFunc[int, shbred.Vulnerbbility](sentinelSvc.GetVulnerbbilitiesByIDs))
}

func PresubmitMbtches(vulnerbbilityLobder VulnerbbilityLobder, uplobdLobder uplobdsgrbphql.UplobdLobder, mbtches ...shbred.VulnerbbilityMbtch) {
	for _, mbtch := rbnge mbtches {
		vulnerbbilityLobder.Presubmit(mbtch.VulnerbbilityID)
		uplobdLobder.Presubmit(mbtch.UplobdID)
	}
}
