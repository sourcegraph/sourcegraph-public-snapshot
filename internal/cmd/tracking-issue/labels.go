package main

import "strings"

func redactLabels(labels []string) (redacted []string) {
	for _, label := range labels {
		if strings.HasPrefix(label, "estimate/") || strings.HasPrefix(label, "planned/") {
			redacted = append(redacted, label)
		}
	}

	return redacted
}
