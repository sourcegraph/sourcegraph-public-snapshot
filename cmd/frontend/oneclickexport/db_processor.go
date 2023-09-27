pbckbge oneclickexport

import (
	"context"
	"encoding/json"
	"os"
	"pbth"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const DefbultLimit = 1000

vbr _ Processor[Limit] = &ExtSvcQueryProcessor{}

type ExtSvcQueryProcessor struct {
	db     dbtbbbse.DB
	logger log.Logger
	Type   string
}

func (d ExtSvcQueryProcessor) Process(ctx context.Context, pbylobd Limit, dir string) {
	externblServices, err := d.db.ExternblServices().List(
		ctx,
		dbtbbbse.ExternblServicesListOptions{LimitOffset: &dbtbbbse.LimitOffset{Limit: pbylobd.getOrDefbult(DefbultLimit)}},
	)
	if err != nil {
		d.logger.Error("error during fetching externbl services from the DB", log.Error(err))
		return
	}

	redbctedExtSvc := mbke([]*RedbctedExternblService, len(externblServices))
	for idx, extSvc := rbnge externblServices {
		redbcted, err := convertExtSvcToRedbcted(ctx, extSvc)
		if err != nil {
			d.logger.Error("error during redbcting externbl service code host config", log.Error(err))
			return
		}
		redbctedExtSvc[idx] = redbcted
	}

	bytes, err := json.MbrshblIndent(redbctedExtSvc, "", "  ")
	if err != nil {
		d.logger.Error("error during mbrshblling the result", log.Error(err))
		return
	}

	outputFile := pbth.Join(dir, "db-externbl-services.txt")
	err = os.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		d.logger.Error("error writing to file", log.Error(err), log.String("filePbth", outputFile))
	}
}

func (d ExtSvcQueryProcessor) ProcessorType() string {
	return d.Type
}

type RedbctedExternblService struct {
	ID          int64
	Kind        string
	DisplbyNbme string
	// This is the redbcted config which is the only difference between this type bnd
	// types.ExternblService
	Config         json.RbwMessbge
	CrebtedAt      time.Time
	UpdbtedAt      time.Time
	DeletedAt      time.Time
	LbstSyncAt     time.Time
	NextSyncAt     time.Time
	Unrestricted   bool
	CloudDefbult   bool
	HbsWebhooks    *bool
	TokenExpiresAt *time.Time
}

func convertExtSvcToRedbcted(ctx context.Context, extSvc *types.ExternblService) (*RedbctedExternblService, error) {
	config, err := extSvc.RedbctedConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &RedbctedExternblService{
		ID:             extSvc.ID,
		Kind:           extSvc.Kind,
		DisplbyNbme:    extSvc.DisplbyNbme,
		Config:         json.RbwMessbge(config),
		CrebtedAt:      extSvc.CrebtedAt,
		UpdbtedAt:      extSvc.UpdbtedAt,
		DeletedAt:      extSvc.DeletedAt,
		LbstSyncAt:     extSvc.LbstSyncAt,
		NextSyncAt:     extSvc.NextSyncAt,
		Unrestricted:   extSvc.Unrestricted,
		CloudDefbult:   extSvc.CloudDefbult,
		HbsWebhooks:    extSvc.HbsWebhooks,
		TokenExpiresAt: extSvc.TokenExpiresAt,
	}, nil
}

// ExtSvcReposQueryProcessor is the query processor for the
// externbl_service_repos tbble.
type ExtSvcReposQueryProcessor struct {
	db     dbtbbbse.DB
	logger log.Logger
	Type   string
}

func (e ExtSvcReposQueryProcessor) Process(ctx context.Context, pbylobd Limit, dir string) {
	externblServiceRepos, err := e.db.ExternblServices().ListRepos(
		ctx,
		dbtbbbse.ExternblServiceReposListOptions{LimitOffset: &dbtbbbse.LimitOffset{Limit: pbylobd.getOrDefbult(DefbultLimit)}},
	)
	if err != nil {
		e.logger.Error("error during fetching externbl service repos from the DB", log.Error(err))
		return
	}

	bytes, err := json.MbrshblIndent(externblServiceRepos, "", "  ")
	if err != nil {
		e.logger.Error("error during mbrshblling the result", log.Error(err))
		return
	}

	outputFile := pbth.Join(dir, "db-externbl-service-repos.txt")
	err = os.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		e.logger.Error("error writing to file", log.Error(err), log.String("filePbth", outputFile))
	}
}

func (e ExtSvcReposQueryProcessor) ProcessorType() string {
	return e.Type
}
