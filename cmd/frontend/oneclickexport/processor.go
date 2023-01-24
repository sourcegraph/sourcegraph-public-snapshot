package oneclickexport

import (
	"context"
	"encoding/json"
	"os"
	"path"

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

type SiteConfigProcessor struct {
	logger log.Logger
	Type   string
}

// Process function of SiteConfigProcessor loads site config, redacts the secrets
// and stores it in a provided tmp directory dir
func (s SiteConfigProcessor) Process(_ context.Context, _ ConfigRequest, dir string) {
	siteConfig, err := conf.RedactSecrets(conf.Raw())
	if err != nil {
		s.logger.Error("error during site config redacting", log.Error(err))
	}

	configBytes := []byte(siteConfig.Site)

	outputFile := path.Join(dir, "site-config.json")
	err = os.WriteFile(outputFile, configBytes, 0644)
	if err != nil {
		s.logger.Error("error writing to file", log.Error(err), log.String("filePath", outputFile))
	}
}

func (s SiteConfigProcessor) ProcessorType() string {
	return s.Type
}

var _ Processor[ConfigRequest] = &CodeHostConfigProcessor{}

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
		c.logger.Error("error getting external services", log.Error(err))
	}

	if len(externalServices) == 0 {
		return
	}

	summaries := make([]*ExternalServiceSummary, len(externalServices))
	for idx, extSvc := range externalServices {
		summary, err := convertToSummary(ctx, extSvc)
		if err != nil {
			// basically this is the only error that can occur
			c.logger.Error("error during redacting the code host config", log.Error(err))
			return
		}
		summaries[idx] = summary
	}

	configBytes, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		c.logger.Error("error during marshalling the code host config", log.Error(err))
	}

	outputFile := path.Join(dir, "code-host-config.json")
	err = os.WriteFile(outputFile, configBytes, 0644)
	if err != nil {
		c.logger.Error("error writing to file", log.Error(err), log.String("filePath", outputFile))
	}
}

type ExternalServiceSummary struct {
	Kind        string          `json:"kind"`
	DisplayName string          `json:"displayName"`
	Config      json.RawMessage `json:"config"`
}

func convertToSummary(ctx context.Context, extSvc *types.ExternalService) (*ExternalServiceSummary, error) {
	config, err := extSvc.RedactedConfig(ctx)
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
