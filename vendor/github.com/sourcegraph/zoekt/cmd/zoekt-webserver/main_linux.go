package main

import (
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	sglog "github.com/sourcegraph/log"
)

func mustRegisterMemoryMapMetrics(logger sglog.Logger) {
	logger = logger.Scoped("memoryMapMetrics")

	// The memory map metrics are collected via /proc, which
	// is only available on linux-based operating systems.

	// Instantiate shared FS objects for accessing /proc and /proc/self,
	// and skip metrics registration if we're aren't able to instantiate them
	// for whatever reason.

	fs, err := procfs.NewDefaultFS()
	if err != nil {
		logger.Debug(
			"skipping registration",
			sglog.String("reason", "failed to initialize proc FS"),
			sglog.String("error", err.Error()),
		)

		return
	}

	info, err := fs.Self()
	if err != nil {
		logger.Debug(
			"skipping registration",
			sglog.String("path", path.Join(procfs.DefaultMountPoint, "self")),
			sglog.String("reason", "failed to initialize process info object for current process"),
			sglog.String("error", err.Error()),
		)

		return
	}

	// Register Prometheus memory map metrics

	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "proc_metrics_memory_map_max_limit",
		Help: "Upper limit on amount of memory mapped regions a process may have.",
	}, func() float64 {
		vm, err := fs.VM()
		if err != nil {
			logger.Debug(
				"failed to read virtual memory statistics for the current process",
				sglog.String("path", path.Join(procfs.DefaultMountPoint, "sys", "vm")),
				sglog.String("error", err.Error()),
			)

			return 0
		}

		if vm.MaxMapCount == nil {
			return 0
		}

		return float64(*vm.MaxMapCount)
	}))

	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "proc_metrics_memory_map_current_count",
		Help: "Amount of memory mapped regions this process is currently using.",
	}, func() float64 {
		procMaps, err := info.ProcMaps()
		if err != nil {
			logger.Debug(
				"failed to read memory mappings for current process",
				sglog.String("path", path.Join(procfs.DefaultMountPoint, "self", "maps")),
				sglog.String("error", err.Error()),
			)

			return 0
		}

		return float64(len(procMaps))
	}))
}
