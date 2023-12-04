package oneclickexport

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Exporter interface {
	// Export accepts an ExportRequest and returns io.Reader of a zip archive
	// with requested data.
	Export(ctx context.Context, request ExportRequest) (io.Reader, error)
}

var _ Exporter = &DataExporter{}

var GlobalExporter Exporter

type DataExporter struct {
	logger           log.Logger
	configProcessors map[string]Processor[ConfigRequest]
	dbProcessors     map[string]Processor[Limit]
}

type ConfigRequest struct {
}

type DBQueryRequest struct {
	TableName string `json:"tableName"`
	Count     Limit  `json:"count"`
}

type Limit int

func (l Limit) getOrDefault(defaultValue int) int {
	if l == 0 {
		return defaultValue
	}
	return int(l)
}

func NewDataExporter(db database.DB, logger log.Logger) Exporter {
	return &DataExporter{
		logger: logger,
		configProcessors: map[string]Processor[ConfigRequest]{
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
		dbProcessors: map[string]Processor[Limit]{
			"external_services": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "external_services",
			},
			"external_service_repos": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "external_services",
			},
		},
	}
}

type ExportRequest struct {
	IncludeSiteConfig     bool              `json:"includeSiteConfig"`
	IncludeCodeHostConfig bool              `json:"includeCodeHostConfig"`
	DBQueries             []*DBQueryRequest `json:"dbQueries"`
}

// Export generates and returns a ZIP archive with the data, specified in request.
// It works like this:
// 1) tmp directory is created (exported files will end up in this directory and
// this directory is zipped in the end)
// 2) ExportRequest is read and each corresponding processor is invoked
// 3) Tmp directory is zipped after all the Processors finished their job
func (e *DataExporter) Export(ctx context.Context, request ExportRequest) (io.Reader, error) {
	// 1) creating a tmp dir
	dir, err := os.MkdirTemp(os.TempDir(), "export-*")
	if err != nil {
		e.logger.Fatal("error during code tmp dir creation", log.Error(err))
	}
	defer os.RemoveAll(dir)

	// 2) tmp dir is passed to every processor
	if request.IncludeSiteConfig {
		e.configProcessors["siteConfig"].Process(ctx, ConfigRequest{}, dir)
	}
	if request.IncludeCodeHostConfig {
		e.configProcessors["codeHostConfig"].Process(ctx, ConfigRequest{}, dir)
	}
	for _, dbQuery := range request.DBQueries {
		e.dbProcessors[dbQuery.TableName].Process(ctx, dbQuery.Count, dir)
	}

	// 3) after all request parts are processed, zip the tmp dir and return its bytes
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// currently, all the directories are skipped because only files are added to the
		// archive
		if info.IsDir() {
			return nil
		}

		// create file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate
		header.Name = filepath.Base(path)

		headerWriter, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(headerWriter, file)
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
