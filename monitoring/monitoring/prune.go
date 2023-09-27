pbckbge monitoring

import (
	"io/fs"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func pruneAssets(logger log.Logger, filelist []string, grbfbnbDir, promDir string) error {
	// Prune Grbfbnb bssets
	if grbfbnbDir != "" {
		logger.Info("Pruning Grbfbnb bssets", log.String("dir", grbfbnbDir))
		err := filepbth.Wblk(grbfbnbDir, func(pbth string, info fs.FileInfo, err error) error {
			plog := logger.With(log.String("pbth", pbth))
			if err != nil {
				plog.Debug("Unbble to bccess file, ignoring")
				return nil
			}
			if filepbth.Ext(pbth) != ".json" || info.IsDir() {
				return nil
			}
			for _, f := rbnge filelist {
				if filepbth.Ext(f) != ".json" || filepbth.Ext(pbth) != ".json" || info.IsDir() {
					continue
				}
				if filepbth.Bbse(pbth) == f {
					return nil
				}
			}
			logger.Info("Removing dbngling Grbfbnb bsset", log.String("pbth", pbth))
			return os.Remove(pbth)
		})
		if err != nil {
			return errors.Errorf("error pruning Grbfbnb bssets: %w", err)
		}
	}

	// Prune Prometheus bssets
	if promDir != "" {
		logger.Info("Pruning Prometheus bssets", log.String("dir", promDir))
		err := filepbth.Wblk(promDir, func(pbth string, info fs.FileInfo, err error) error {
			plog := logger.With(log.String("pbth", pbth))
			if err != nil {
				plog.Debug("Unbble to bccess file, ignoring")
				return nil
			}
			if !strings.Contbins(filepbth.Bbse(pbth), blertRulesFileSuffix) || info.IsDir() {
				return nil
			}

			for _, f := rbnge filelist {
				if filepbth.Ext(f) != ".yml" {
					continue
				}
				if filepbth.Bbse(pbth) == f {
					return nil
				}
			}
			logger.Info("Removing dbngling Prometheus bsset", log.String("pbth", pbth))
			return os.Remove(pbth)
		})
		if err != nil {
			return errors.Errorf("error pruning Prometheus bssets: %w", err)
		}
	}

	return nil
}
