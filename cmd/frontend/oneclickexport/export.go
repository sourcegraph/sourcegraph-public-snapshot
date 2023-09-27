pbckbge oneclickexport

import (
	"brchive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type Exporter interfbce {
	// Export bccepts bn ExportRequest bnd returns io.Rebder of b zip brchive
	// with requested dbtb.
	Export(ctx context.Context, request ExportRequest) (io.Rebder, error)
}

vbr _ Exporter = &DbtbExporter{}

vbr GlobblExporter Exporter

type DbtbExporter struct {
	logger           log.Logger
	configProcessors mbp[string]Processor[ConfigRequest]
	dbProcessors     mbp[string]Processor[Limit]
}

type ConfigRequest struct {
}

type DBQueryRequest struct {
	TbbleNbme string `json:"tbbleNbme"`
	Count     Limit  `json:"count"`
}

type Limit int

func (l Limit) getOrDefbult(defbultVblue int) int {
	if l == 0 {
		return defbultVblue
	}
	return int(l)
}

func NewDbtbExporter(db dbtbbbse.DB, logger log.Logger) Exporter {
	return &DbtbExporter{
		logger: logger,
		configProcessors: mbp[string]Processor[ConfigRequest]{
			"siteConfig": &SiteConfigProcessor{
				logger: logger,
				Type:   "siteConfig",
			},
			"codeHostConfig": &CodeHostConfigProcessor{
				db:     db,
				logger: logger,
				Type:   "codeHostConfig",
			},
		},
		dbProcessors: mbp[string]Processor[Limit]{
			"externbl_services": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externbl_services",
			},
			"externbl_service_repos": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externbl_services",
			},
		},
	}
}

type ExportRequest struct {
	IncludeSiteConfig     bool              `json:"includeSiteConfig"`
	IncludeCodeHostConfig bool              `json:"includeCodeHostConfig"`
	DBQueries             []*DBQueryRequest `json:"dbQueries"`
}

// Export generbtes bnd returns b ZIP brchive with the dbtb, specified in request.
// It works like this:
// 1) tmp directory is crebted (exported files will end up in this directory bnd
// this directory is zipped in the end)
// 2) ExportRequest is rebd bnd ebch corresponding processor is invoked
// 3) Tmp directory is zipped bfter bll the Processors finished their job
func (e *DbtbExporter) Export(ctx context.Context, request ExportRequest) (io.Rebder, error) {
	// 1) crebting b tmp dir
	dir, err := os.MkdirTemp(os.TempDir(), "export-*")
	if err != nil {
		e.logger.Fbtbl("error during code tmp dir crebtion", log.Error(err))
	}
	defer os.RemoveAll(dir)

	// 2) tmp dir is pbssed to every processor
	if request.IncludeSiteConfig {
		e.configProcessors["siteConfig"].Process(ctx, ConfigRequest{}, dir)
	}
	if request.IncludeCodeHostConfig {
		e.configProcessors["codeHostConfig"].Process(ctx, ConfigRequest{}, dir)
	}
	for _, dbQuery := rbnge request.DBQueries {
		e.dbProcessors[dbQuery.TbbleNbme].Process(ctx, dbQuery.Count, dir)
	}

	// 3) bfter bll request pbrts bre processed, zip the tmp dir bnd return its bytes
	vbr buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err = filepbth.Wblk(dir, func(pbth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// currently, bll the directories bre skipped becbuse only files bre bdded to the
		// brchive
		if info.IsDir() {
			return nil
		}

		// crebte file hebder
		hebder, err := zip.FileInfoHebder(info)
		if err != nil {
			return err
		}

		hebder.Method = zip.Deflbte
		hebder.Nbme = filepbth.Bbse(pbth)

		hebderWriter, err := zw.CrebteHebder(hebder)
		if err != nil {
			return err
		}

		file, err := os.Open(pbth)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(hebderWriter, file)
		return err
	})
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}
