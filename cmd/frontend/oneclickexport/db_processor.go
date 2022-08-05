package oneclickexport

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const DefaultLimit = 1000

var _ Processor[Limit] = &ExtSvcDBQueryProcessor{}

type ExtSvcDBQueryProcessor struct {
	db     database.DB
	logger log.Logger
	Type   string
}

func (d ExtSvcDBQueryProcessor) Process(ctx context.Context, payload Limit, dir string) {
	externalServices, err := d.db.ExternalServices().List(
		ctx,
		database.ExternalServicesListOptions{LimitOffset: &database.LimitOffset{Limit: payload.getOrDefault(DefaultLimit)}},
	)
	if err != nil {
		d.logger.Error("Error during fetching external services from the DB", log.Error(err))
		return
	}

	redactedExtSvc := make([]*RedactedExternalService, len(externalServices))

	for idx, extSvc := range externalServices {
		redacted, err := convertExtSvcToRedacted(extSvc)
		if err != nil {
			d.logger.Error("Error during redacting external service code host config", log.Error(err))
			return
		}
		redactedExtSvc[idx] = redacted
	}

	bytes, err := json.MarshalIndent(redactedExtSvc, "", "  ")
	if err != nil {
		d.logger.Error("Error during marshalling the result", log.Error(err))
		return
	}

	err = ioutil.WriteFile(dir+"/db-external-services.txt", bytes, 0644)

	if err != nil {
		d.logger.Error("Error during external_services export", log.Error(err))
	}
}

func (d ExtSvcDBQueryProcessor) ProcessorType() string {
	return d.Type
}

type RedactedExternalService struct {
	ID          int64
	Kind        string
	DisplayName string
	// This is the redacted config which is the only difference between this type and
	// types.ExternalService
	Config          json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       time.Time
	LastSyncAt      time.Time
	NextSyncAt      time.Time
	NamespaceUserID int32
	NamespaceOrgID  int32
	Unrestricted    bool
	CloudDefault    bool
	HasWebhooks     *bool
	TokenExpiresAt  *time.Time
}

func convertExtSvcToRedacted(extSvc *types.ExternalService) (*RedactedExternalService, error) {
	config, err := extSvc.RedactedConfig()
	if err != nil {
		return nil, err
	}
	return &RedactedExternalService{
		ID:              extSvc.ID,
		Kind:            extSvc.Kind,
		DisplayName:     extSvc.DisplayName,
		Config:          json.RawMessage(config),
		CreatedAt:       extSvc.CreatedAt,
		UpdatedAt:       extSvc.UpdatedAt,
		DeletedAt:       extSvc.DeletedAt,
		LastSyncAt:      extSvc.LastSyncAt,
		NextSyncAt:      extSvc.NextSyncAt,
		NamespaceUserID: extSvc.NamespaceUserID,
		NamespaceOrgID:  extSvc.NamespaceOrgID,
		Unrestricted:    extSvc.Unrestricted,
		CloudDefault:    extSvc.CloudDefault,
		HasWebhooks:     extSvc.HasWebhooks,
		TokenExpiresAt:  extSvc.TokenExpiresAt,
	}, nil
}

type Limit int

func (l Limit) getOrDefault(defaultValue int) int {
	if l == 0 {
		return defaultValue
	}
	return int(l)
}

type DBQueryRequest struct {
	TableName string `json:"tableName"`
	Count     Limit  `json:"count"`
}
