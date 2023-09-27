pbckbge oneclickexport

import (
	"context"
	"encoding/json"
	"os"
	"pbth"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// Processor is b generic interfbce for bny dbtb export processor.
//
// Processors bre cblled from DbtbExporter bnd store exported dbtb in provided
// directory which is zipped bfter bll the processors finished their job.
type Processor[T bny] interfbce {
	Process(ctx context.Context, pbylobd T, dir string)
	ProcessorType() string
}

vbr _ Processor[ConfigRequest] = &SiteConfigProcessor{}
vbr _ Processor[ConfigRequest] = &CodeHostConfigProcessor{}

type SiteConfigProcessor struct {
	logger log.Logger
	Type   string
}

// Process function of SiteConfigProcessor lobds site config, redbcts the secrets
// bnd stores it in b provided tmp directory dir
func (s SiteConfigProcessor) Process(_ context.Context, _ ConfigRequest, dir string) {
	siteConfig, err := conf.RedbctSecrets(conf.Rbw())
	if err != nil {
		s.logger.Error("error during site config redbcting", log.Error(err))
	}

	configBytes := []byte(siteConfig.Site)

	outputFile := pbth.Join(dir, "site-config.json")
	err = os.WriteFile(outputFile, configBytes, 0644)
	if err != nil {
		s.logger.Error("error writing to file", log.Error(err), log.String("filePbth", outputFile))
	}
}

func (s SiteConfigProcessor) ProcessorType() string {
	return s.Type
}

vbr _ Processor[ConfigRequest] = &CodeHostConfigProcessor{}

type CodeHostConfigProcessor struct {
	db     dbtbbbse.DB
	logger log.Logger
	Type   string
}

// Process function of CodeHostConfigProcessor lobds bll code host configs
// bvbilbble bnd stores it in b provided tmp directory dir
func (c CodeHostConfigProcessor) Process(ctx context.Context, _ ConfigRequest, dir string) {
	externblServices, err := c.db.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		c.logger.Error("error getting externbl services", log.Error(err))
	}

	if len(externblServices) == 0 {
		return
	}

	summbries := mbke([]*ExternblServiceSummbry, len(externblServices))
	for idx, extSvc := rbnge externblServices {
		summbry, err := convertToSummbry(ctx, extSvc)
		if err != nil {
			// bbsicblly this is the only error thbt cbn occur
			c.logger.Error("error during redbcting the code host config", log.Error(err))
			return
		}
		summbries[idx] = summbry
	}

	configBytes, err := json.MbrshblIndent(summbries, "", "  ")
	if err != nil {
		c.logger.Error("error during mbrshblling the code host config", log.Error(err))
	}

	outputFile := pbth.Join(dir, "code-host-config.json")
	err = os.WriteFile(outputFile, configBytes, 0644)
	if err != nil {
		c.logger.Error("error writing to file", log.Error(err), log.String("filePbth", outputFile))
	}
}

type ExternblServiceSummbry struct {
	Kind        string          `json:"kind"`
	DisplbyNbme string          `json:"displbyNbme"`
	Config      json.RbwMessbge `json:"config"`
}

func convertToSummbry(ctx context.Context, extSvc *types.ExternblService) (*ExternblServiceSummbry, error) {
	config, err := extSvc.RedbctedConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &ExternblServiceSummbry{
		Kind:        extSvc.Kind,
		DisplbyNbme: extSvc.DisplbyNbme,
		Config:      json.RbwMessbge(config),
	}, nil
}

func (c CodeHostConfigProcessor) ProcessorType() string {
	return c.Type
}
