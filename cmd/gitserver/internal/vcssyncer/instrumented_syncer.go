package vcssyncer

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"io"
	"strconv"
	"strings"
	"time"
)

// fetchBuckets are the b
var fetchBuckets = append([]float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30}, prometheus.LinearBuckets(60, 5*60, 24)...) // 50ms -> 120 minutes

var (
	metricFetchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vcssyncer_fetch_duration_seconds",
		Help:    "Time taken to fetch a repository",
		Buckets: fetchBuckets,
	}, []string{"type", "success"})

	metricCloneDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vcssyncer_clone_duration_seconds",
		Help:    "Time taken to clone a repository",
		Buckets: fetchBuckets,
	}, []string{"type", "success"})

	metricIsCloneableDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vcssyncer_is_cloneable_duration_seconds",
		Help:    "Time taken to check if a repository is cloneable",
		Buckets: prometheus.DefBuckets,
	}, []string{"type", "success"})
)

func newInstrumentedSyncer(syncer VCSSyncer) VCSSyncer {
	typ := syncer.Type()
	typ = strings.ToLower(typ)
	typ = strings.ReplaceAll(typ, " ", "_")
	typ = strings.ReplaceAll(typ, "-", "_")

	return &instrumentedSyncer{
		base:               syncer,
		formattedTypeLabel: typ,
	}
}

type instrumentedSyncer struct {
	base               VCSSyncer
	formattedTypeLabel string
}

func (i *instrumentedSyncer) Type() string {
	return i.base.Type()
}

func (i *instrumentedSyncer) IsCloneable(ctx context.Context, repoName api.RepoName) error {
	start := time.Now()
	succeeded := true

	err := i.base.IsCloneable(ctx, repoName)
	if err != nil {
		succeeded = false
	}

	metricIsCloneableDuration.WithLabelValues(i.formattedTypeLabel, strconv.FormatBool(succeeded)).Observe(time.Since(start).Seconds())

	return err
}

func (i *instrumentedSyncer) Clone(ctx context.Context, repo api.RepoName, targetDir common.GitDir, tmpPath string, progressWriter io.Writer) error {
	start := time.Now()
	succeeded := true

	err := i.base.Clone(ctx, repo, targetDir, tmpPath, progressWriter)
	if err != nil {
		succeeded = false
	}

	metricCloneDuration.WithLabelValues(i.formattedTypeLabel, strconv.FormatBool(succeeded)).Observe(time.Since(start).Seconds())

	return err
}

func (i *instrumentedSyncer) Fetch(ctx context.Context, repoName api.RepoName, dir common.GitDir, revspec string) ([]byte, error) {
	start := time.Now()
	succeeded := true

	data, err := i.base.Fetch(ctx, repoName, dir, revspec)
	if err != nil {
		succeeded = false
	}

	metricFetchDuration.WithLabelValues(i.formattedTypeLabel, strconv.FormatBool(succeeded)).Observe(time.Since(start).Seconds())

	return data, err
}

var _ VCSSyncer = &instrumentedSyncer{}
