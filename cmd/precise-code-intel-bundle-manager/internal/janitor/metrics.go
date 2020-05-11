package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	UploadFilesRemoved          prometheus.Counter
	OprphanedBundleFilesRemoved prometheus.Counter
	EvictedBundleFilesRemoved   prometheus.Counter
	Errors                      prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	uploadFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_upload_files_removed_total",
		Help: "Total number of upload files removed (due to age)",
	})
	r.MustRegister(uploadFilesRemoved)

	oprphanedBundleFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_orphaned_bundle_files_removed_total",
		Help: "Total number of bundle files removed (with no corresponding database entry)",
	})
	r.MustRegister(oprphanedBundleFilesRemoved)

	evictedBundleFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_evicted_bundle_files_removed_total",
		Help: "Total number of bundles files removed (after evicting them from the database)",
	})
	r.MustRegister(evictedBundleFilesRemoved)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_errors_total",
		Help: "Total number of errors when running the janitor",
	})
	r.MustRegister(errors)

	return JanitorMetrics{
		UploadFilesRemoved:          uploadFilesRemoved,
		OprphanedBundleFilesRemoved: oprphanedBundleFilesRemoved,
		EvictedBundleFilesRemoved:   evictedBundleFilesRemoved,
		Errors:                      errors,
	}
}
