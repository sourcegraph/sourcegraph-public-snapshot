package util

import (
	"sort"
	"strings"
)

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
