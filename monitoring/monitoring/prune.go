package monitoring

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func pruneAssets(logger log.Logger, filelist []string, grafanaDir, promDir string) error {
	// Prune Grafana assets
	if grafanaDir != "" {
		logger.Info("Pruning Grafana assets", log.String("dir", grafanaDir))
		err := filepath.Walk(grafanaDir, func(path string, info fs.FileInfo, err error) error {
			plog := logger.With(log.String("path", path))
			if err != nil {
				plog.Debug("Unable to access file, ignoring")
				return nil
			}
			if filepath.Ext(path) != ".json" || info.IsDir() {
				return nil
			}
			for _, f := range filelist {
				if filepath.Ext(f) != ".json" || filepath.Ext(path) != ".json" || info.IsDir() {
					continue
				}
				if filepath.Base(path) == f {
					return nil
				}
			}
			logger.Info("Removing dangling Grafana asset", log.String("path", path))
			return os.Remove(path)
		})
		if err != nil {
			return errors.Errorf("error pruning Grafana assets: %w", err)
		}
	}

	// Prune Prometheus assets
	if promDir != "" {
		logger.Info("Pruning Prometheus assets", log.String("dir", promDir))
		err := filepath.Walk(promDir, func(path string, info fs.FileInfo, err error) error {
			plog := logger.With(log.String("path", path))
			if err != nil {
				plog.Debug("Unable to access file, ignoring")
				return nil
			}
			if !strings.Contains(filepath.Base(path), alertRulesFileSuffix) || info.IsDir() {
				return nil
			}

			for _, f := range filelist {
				if filepath.Ext(f) != ".yml" {
					continue
				}
				if filepath.Base(path) == f {
					return nil
				}
			}
			logger.Info("Removing dangling Prometheus asset", log.String("path", path))
			return os.Remove(path)
		})
		if err != nil {
			return errors.Errorf("error pruning Prometheus assets: %w", err)
		}
	}

	return nil
}
