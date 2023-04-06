package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/dataloader"
)

func NewVulnerabilityLoaderFactory(sentinelSvc SentinelService) *dataloader.DataloaderFactory[int, shared.Vulnerability] {
	return dataloader.NewDataloaderFactory[int, shared.Vulnerability](dataloader.BackingServiceFunc[int, shared.Vulnerability](sentinelSvc.GetVulnerabilitiesByIDs))
}
