package oneclickexport

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const DefaultLimit = 1000

var _ Processor[Limit] = &ExtSvcQueryProcessor{}

type ExtSvcQueryProcessor struct {
	db     database.DB
	logger log.Logger
	Type   string
}

func (d ExtSvcQueryProcessor) Process(ctx context.Context, payload Limit, dir string) {
	externalServices, err := d.db.ExternalServices().List(
		ctx,
		database.ExternalServicesListOptions{LimitOffset: &database.LimitOffset{Limit: payload.getOrDefault(DefaultLimit)}},
	)
	if err != nil {
		d.logger.Error("error during fetching external services from the DB", log.Error(err))
		return
	}

	redactedExtSvc := make([]*RedactedExternalService, len(externalServices))
	for idx, extSvc := range externalServices {
		redacted, err := convertExtSvcToRedacted(ctx, extSvc)
		if err != nil {
			d.logger.Error("error during redacting external service code host config", log.Error(err))
			return
		}
		redactedExtSvc[idx] = redacted
	}

	bytes, err := json.MarshalIndent(redactedExtSvc, "", "  ")
	if err != nil {
		d.logger.Error("error during marshalling the result", log.Error(err))
		return
	}

	outputFile := path.Join(dir, "db-external-services.txt")
	err = os.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		d.logger.Error("error writing to file", log.Error(err), log.String("filePath", outputFile))
	}
}

func (d ExtSvcQueryProcessor) ProcessorType() string {
	return d.Type
}

type RedactedExternalService struct {
	ID          int64
	Kind        string
	DisplayName string
	// This is the redacted config which is the only difference between this type and
	// types.ExternalService
	Config         json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
	LastSyncAt     time.Time
	NextSyncAt     time.Time
	Unrestricted   bool
	HasWebhooks    *bool
	TokenExpiresAt *time.Time
}

func convertExtSvcToRedacted(ctx context.Context, extSvc *types.ExternalService) (*RedactedExternalService, error) {
	config, err := extSvc.RedactedConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &RedactedExternalService{
		ID:             extSvc.ID,
		Kind:           extSvc.Kind,
		DisplayName:    extSvc.DisplayName,
		Config:         json.RawMessage(config),
		CreatedAt:      extSvc.CreatedAt,
		UpdatedAt:      extSvc.UpdatedAt,
		DeletedAt:      extSvc.DeletedAt,
		LastSyncAt:     extSvc.LastSyncAt,
		NextSyncAt:     extSvc.NextSyncAt,
		Unrestricted:   extSvc.Unrestricted,
		HasWebhooks:    extSvc.HasWebhooks,
		TokenExpiresAt: extSvc.TokenExpiresAt,
	}, nil
}

// ExtSvcReposQueryProcessor is the query processor for the
// external_service_repos table.
type ExtSvcReposQueryProcessor struct {
	db     database.DB
	logger log.Logger
	Type   string
}

func (e ExtSvcReposQueryProcessor) Process(ctx context.Context, payload Limit, dir string) {
	externalServiceRepos, err := e.db.ExternalServices().ListRepos(
		ctx,
		database.ExternalServiceReposListOptions{LimitOffset: &database.LimitOffset{Limit: payload.getOrDefault(DefaultLimit)}},
	)
	if err != nil {
		e.logger.Error("error during fetching external service repos from the DB", log.Error(err))
		return
	}

	bytes, err := json.MarshalIndent(externalServiceRepos, "", "  ")
	if err != nil {
		e.logger.Error("error during marshalling the result", log.Error(err))
		return
	}

	outputFile := path.Join(dir, "db-external-service-repos.txt")
	err = os.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		e.logger.Error("error writing to file", log.Error(err), log.String("filePath", outputFile))
	}
}

func (e ExtSvcReposQueryProcessor) ProcessorType() string {
	return e.Type
}
