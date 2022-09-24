package store

import (
	"context"
	"sort"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/storage"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

var UnsupportedConversionErr = errors.New("unsupported conversion for Code Insights storage")

type converter struct {
	store *Store
}

func NewConverter(store *Store) *converter {
	return &converter{store: store}
}

func (c *converter) Convert(ctx context.Context, seriesDefinition types.InsightSeries, desiredFormat storage.DataFormat) error {
	if seriesDefinition.DataFormat == storage.Uncompressed && desiredFormat == storage.Gorilla {
		return c.uncompressedToGorilla(ctx, seriesDefinition)
	} else {
		return UnsupportedConversionErr
	}
}

func (c *converter) uncompressedToGorilla(ctx context.Context, seriesDefinition types.InsightSeries) error {
	repos, err := c.allReposForSeriesOnSeriesPoints(ctx, seriesDefinition.SeriesID)
	if err != nil {
		return errors.Wrap(err, "error selecting repo_ids for series")
	}

	log15.Info("repos", "r", repos)

	for _, repo := range repos {
		id := api.RepoID(repo)
		points, err := c.store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesDefinition.SeriesID, RepoID: &id})
		if err != nil {
			return err
		}

		getKey := func(cap *string) string {
			if cap == nil {
				return ""
			}
			return *cap
		}

		// group points by capture
		byCapture := make(map[string][]SeriesPoint)
		for _, point := range points {
			byCapture[getKey(point.Capture)] = append(byCapture[getKey(point.Capture)], point)
		}

		for key, capturePoints := range byCapture {
			if key == "" {
				continue
			}
			sort.Slice(capturePoints, func(i, j int) bool {
				return capturePoints[i].Time.Before(capturePoints[j].Time)
			})

			samples := make([]RawSample, 0, len(capturePoints))
			for _, point := range capturePoints {
				samples = append(samples, RawSample{
					Time:  uint32(point.Time.Unix()),
					Value: point.Value,
				})
			}

			row := UncompressedRow{
				altFormatRowMetadata: altFormatRowMetadata{
					RepoId:  repo,
					Capture: &key,
				},
				Samples: samples,
			}

			err = c.store.StoreAlternateFormat(ctx, row, uint32(seriesDefinition.ID))
			if err != nil {
				return errors.Wrap(err, "StoreAlternateFormat")
			}
		}
	}

	return nil
}

func (c *converter) allReposForSeriesOnSeriesPoints(ctx context.Context, seriesId string) (_ []uint32, err error) {
	q := `select distinct repo_id from series_points where series_id = %s order by repo_id;`
	return basestore.Scanuint32s(c.store.Query(ctx, sqlf.Sprintf(q, seriesId)))
}
