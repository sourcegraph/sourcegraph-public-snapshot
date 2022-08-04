package oneclickexport

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Processor is a generic interface for any data export processor.
//
// Processors are called from DataExporter and store exported data in provided
// directory which is zipped after all the processors finished their job.
type Processor[T any] interface {
	Process(ctx context.Context, payload T, dir string)
	ProcessorType() string
}

var _ Processor[ConfigRequest] = &SiteConfigProcessor{}
var _ Processor[ConfigRequest] = &CodeHostConfigProcessor{}

type ConfigRequest struct {
}

type SiteConfigProcessor struct {
	logger log.Logger
	Type   string
}

// Process function of SiteConfigProcessor loads site config, redacts the secrets
// and stores it in a provided tmp directory dir
func (s SiteConfigProcessor) Process(_ context.Context, _ ConfigRequest, dir string) {
	siteConfig, err := conf.RedactSecrets(conf.Raw())
	if err != nil {
		s.logger.Error("Error during site config redacting", log.Error(err))
	}

	configBytes := []byte(siteConfig.Site)

	err = ioutil.WriteFile(dir+"/site-config.json", configBytes, 0644)

	if err != nil {
		s.logger.Error("Error during site config export", log.Error(err))
	}
}

func (s SiteConfigProcessor) ProcessorType() string {
	return s.Type
}

type CodeHostConfigProcessor struct {
	db     database.DB
	logger log.Logger
	Type   string
}

// Process function of CodeHostConfigProcessor loads all code host configs
// available and stores it in a provided tmp directory dir
func (c CodeHostConfigProcessor) Process(ctx context.Context, _ ConfigRequest, dir string) {
	externalServices, err := c.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{})
	if err != nil {
		c.logger.Error("Error getting external services", log.Error(err))
	}

	if len(externalServices) == 0 {
		return
	}

	summaries := make([]*ExternalServiceSummary, len(externalServices))
	for idx, extSvc := range externalServices {
		summary, err := convertToSummary(extSvc)
		if err != nil {
			// basically this is the only error that can occur
			c.logger.Error("Error during redacting the code host config", log.Error(err))
			return
		}
		summaries[idx] = summary
	}

	configBytes, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		c.logger.Error("Error during marshalling the code host config", log.Error(err))
	}

	err = ioutil.WriteFile(dir+"/code-host-config.json", configBytes, 0644)

	if err != nil {
		c.logger.Error("Error during code host config export", log.Error(err))
	}
}

type ExternalServiceSummary struct {
	Kind        string          `json:"kind"`
	DisplayName string          `json:"displayName"`
	Config      json.RawMessage `json:"config"`
}

func convertToSummary(extSvc *types.ExternalService) (*ExternalServiceSummary, error) {
	config, err := extSvc.RedactedConfig()
	if err != nil {
		return nil, err
	}
	return &ExternalServiceSummary{
		Kind:        extSvc.Kind,
		DisplayName: extSvc.DisplayName,
		Config:      json.RawMessage(config),
	}, nil
}

func (c CodeHostConfigProcessor) ProcessorType() string {
	return c.Type
}
