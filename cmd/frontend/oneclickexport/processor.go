package oneclickexport

import (
	"io/ioutil"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// Processor is a generic interface for any data export processor.
//
// Processors are called from DataExporter and store exported data in provided
// directory which is zipped after all the processors finished their job.
type Processor[T any] interface {
	Process(payload T, dir string)
	ProcessorType() string
}

var _ Processor[ConfigRequest] = &SiteConfigProcessor{}

type SiteConfigProcessor struct {
	logger log.Logger
	Type   string
}

type ConfigRequest struct {
}

// Process function of SiteConfigProcessor loads site config, redacts the secrets
// and stores it in a provided tmp directory dir.
func (g SiteConfigProcessor) Process(_ ConfigRequest, dir string) {
	siteConfig, err := conf.RedactSecrets(conf.Raw())
	if err != nil {
		g.logger.Error("Error during site config redacting", log.Error(err))
	}

	configBytes := []byte(siteConfig.Site)

	err = ioutil.WriteFile(dir+"/site-config.json", configBytes, 0644)

	if err != nil {
		g.logger.Error("Error during site config export", log.Error(err))
	}
}

func (g SiteConfigProcessor) ProcessorType() string {
	return g.Type
}
