// Package mountinfo provides a Prometheus collector that advertises
// the names of block storage devices backing the requested file paths.
package mountinfo

import (
	"github.com/prometheus/client_golang/prometheus"
	sglog "github.com/sourcegraph/log"
)

// CollectorOpts modifies the behavior of the metric created
// by NewCollector.
type CollectorOpts struct {
	// If non-empty, Namespace prefixes the "mount_point_info" metric by the provided string and
	// an underscore ("_").
	Namespace string
}

// NewCollector returns a Prometheus collector that collects a single metric, "mount_point_info",
// that contains the names of the block storage devices backing each of the requested mounts.
//
// Mounts is a set of name -> file path mappings (example: {"indexDir": "/home/.zoekt"}).
//
// The metric "mount_point_info" has a constant value of 1 and two labels:
//   - mount_name: caller-provided name for the given mount (example: "indexDir")
//   - device: name of the block device that backs the given mount file path (example: "sdb")
//
// This metric currently works only on Linux-based operating systems that have access to the sysfs pseudo-filesystem.
// On all other operating systems, this metric will not emit any values.
func NewCollector(logger sglog.Logger, opts CollectorOpts, mounts map[string]string) prometheus.Collector {
	logger = logger.Scoped("mountPointInfo")

	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Name:      "mount_point_info",
		Help:      "An info metric with a constant '1' value that contains mount_name, device mappings",
	}, []string{"mount_name", "device"})

	for name, filePath := range mounts {
		// for each <mountName>:<mountFilePath> pairing,
		// discover the name of the block device that stores <mountFilePath>.
		discoveryLogger := logger.Scoped("deviceNameDiscovery").With(
			sglog.String("mountName", name),
			sglog.String("mountFilePath", filePath),
		)

		device, err := discoverDeviceName(discoveryLogger, filePath)
		if err != nil {
			discoveryLogger.Warn("skipping metric registration",
				sglog.String("reason", "failed to discover device name"),
				sglog.Error(err),
			)

			continue
		}

		discoveryLogger.Debug("discovered device name",
			sglog.String("deviceName", device),
		)

		metric.WithLabelValues(name, device).Set(1)
	}

	return metric
}
