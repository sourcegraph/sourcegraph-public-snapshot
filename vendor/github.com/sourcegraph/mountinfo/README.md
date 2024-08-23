# `github.com/sourcegraph/mountinfo`

This Go package provides a Prometheus collector that advertises the names of block storage devices backing the requested file paths.

See the doc comment for `NewCollector` in [info.go](./info.go) for more information.

(snippet):

```go
// NewCollector returns a Prometheus collector that collects a single metric, "mount_point_info",
// that contains the names of the block storage devices backing each of the requested mounts.
//
// Mounts is a set of name -> file path mappings (example: {"indexDir": "/home/.zoekt"}).
//
// The metric "mount_point_info" has a constant value of 1 and two labels:
//   - mount_name: caller-provided name for the given mount (example: "indexDir")
//   - device: name of the block device that backs the given mount file path (example: "sdb")
```
