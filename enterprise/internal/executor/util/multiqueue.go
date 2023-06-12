package util

import (
	"strings"
)

func FormatQueueNamesForMetrics(queueName string, queueNames []string) string {
	var formatted string
	if len(queueNames) > 0 {
		formatted = strings.Join(queueNames, "_")
	} else {
		formatted = queueName
	}
	return formatted
}
