package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	UploadFilesRemoved        prometheus.Counter
	PartFilesRemoved          prometheus.Counter
	OrphanedFilesRemoved      prometheus.Counter
	EvictedBundleFilesRemoved prometheus.Counter
	UploadRecordsRemoved      prometheus.Counter
	Errors                    prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	uploadFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_upload_files_removed_total",
		Help: "Total number of upload files removed (due to age)",
	})
	r.MustRegister(uploadFilesRemoved)

	partFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_part_files_removed_total",
		Help: "Total number of upload and database part files removed (due to age)",
	})
	r.MustRegister(partFilesRemoved)

	orphanedFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_orphaned_files_removed_total",
		Help: "Total number of bundle files removed (with no corresponding successful database entry)",
	})
	r.MustRegister(orphanedFilesRemoved)

	evictedBundleFilesRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_evicted_bundle_files_removed_total",
		Help: "Total number of bundle files removed (after evicting them from the database)",
	})
	r.MustRegister(evictedBundleFilesRemoved)

	uploadRecordsRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_upload_records_removed_total",
		Help: "Total number of processed upload records removed",
	})
	r.MustRegister(uploadRecordsRemoved)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_bundle_manager_janitor_errors_total",
		Help: "Total number of errors when running the janitor",
	})
	r.MustRegister(errors)

	return JanitorMetrics{
		UploadFilesRemoved:        uploadFilesRemoved,
		PartFilesRemoved:          partFilesRemoved,
		OrphanedFilesRemoved:      orphanedFilesRemoved,
		EvictedBundleFilesRemoved: evictedBundleFilesRemoved,
		UploadRecordsRemoved:      uploadRecordsRemoved,
		Errors:                    errors,
	}
}
