package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/dataloader"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
)

type (
	VulnerabilityLoaderFactory = *dataloader.LoaderFactory[int, shared.Vulnerability]
	VulnerabilityLoader        = *dataloader.Loader[int, shared.Vulnerability]
)

func NewVulnerabilityLoaderFactory(sentinelSvc SentinelService) VulnerabilityLoaderFactory {
	return dataloader.NewLoaderFactory[int, shared.Vulnerability](dataloader.BackingServiceFunc[int, shared.Vulnerability](sentinelSvc.GetVulnerabilitiesByIDs))
}

func PresubmitMatches(vulnerabilityLoader VulnerabilityLoader, uploadLoader uploadsgraphql.UploadLoader, matches ...shared.VulnerabilityMatch) {
	for _, match := range matches {
		vulnerabilityLoader.Presubmit(match.VulnerabilityID)
		uploadLoader.Presubmit(match.UploadID)
	}
}
