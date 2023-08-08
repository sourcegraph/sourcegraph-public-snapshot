package util

import (
	"sort"
	"strings"
)

// FormatQueueNamesForMetrics returns a single string that is used to publish autoscaling metrics.
// When queueName is not empty, the same value is returned ("batches" -> "batches").
// When queueNames is not empty, the elements are alphabetically sorted and concatenated with underscores (["codeintel", "batches'] -> "batches_codeintel")
func FormatQueueNamesForMetrics(queueName string, queueNames []string) string {
	var formatted string
	if len(queueNames) > 0 {
		// sort alphabetically to ensure order of definition by users doesn't matter
		sort.Strings(queueNames)
		formatted = strings.Join(queueNames, "_")
	} else {
		formatted = queueName
	}
	return formatted
}
