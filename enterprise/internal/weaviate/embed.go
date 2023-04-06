package weaviate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type readFile func(fileName string) ([]byte, error)

func EmbedRepo(
	ctx context.Context,
	logger log.Logger,
	repoName api.RepoName,
	revision api.CommitID,
	fileNames []string,
	excludedFilePathPatterns []*paths.GlobPattern,
	client *weaviate.Client,
	readFile readFile,
) error {

	batchSize := 100
	batch := client.Batch().ObjectsBatcher()

	cnt := 0
	for _, fileName := range fileNames {
		if isExcludedFilePath(fileName, excludedFilePathPatterns) {
			continue
		}

		typ := "code"
		if isValidTextFile(fileName) {
			typ = "text"
		}

		b, err := readFile(fileName)
		if err != nil {
			return err
		}

		cnt++
		logger.Info("adding object")
		batch.WithObjects(&models.Object{
			Class:              "Code",
			LastUpdateTimeUnix: 0,
			Properties: map[string]string{
				"filename":   fileName,
				"content":    string(b),
				"repository": string(repoName),
				"type":       typ,
				"revision":   string(revision),
			},
		})

		if cnt%batchSize == 0 {
			_, err := batch.Do(ctx)
			if err != nil {
				return err
			}
			cnt = 0
		}
	}

	if cnt > 0 {
		logger.Info("sending final batch")
		resp, err := batch.Do(ctx)
		if err != nil {
			logger.Error("error sending final batch", log.Error(err))
		}
		for _, r := range resp {
			if r.Result.Errors == nil {
				continue
			}
			for _, e := range r.Result.Errors.Error {
				logger.Error(fmt.Sprintf("error adding object: %s", e.Message))
			}
		}

		return err
	}

	return nil
}

func isExcludedFilePath(filePath string, excludedFilePathPatterns []*paths.GlobPattern) bool {
	for _, excludedFilePathPattern := range excludedFilePathPatterns {
		if excludedFilePathPattern.Match(filePath) {
			return true
		}
	}
	return false
}

func isValidTextFile(fileName string) bool {
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	_, ok := textFileExtensions[strings.ToLower(ext)]
	if ok {
		return true
	}
	basename := strings.ToLower(filepath.Base(fileName))
	return strings.HasPrefix(basename, "license")
}

var textFileExtensions = map[string]struct{}{
	"md":       {},
	"markdown": {},
	"rst":      {},
	"txt":      {},
}
