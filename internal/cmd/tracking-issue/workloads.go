package main

import (
	"sort"
	"strings"
)

type Workloads map[string]*Workload

func (ws Workloads) Markdown(labelAllowlist []string) string {
	assignees := make([]string, 0, len(ws))
	for assignee := range ws {
		assignees = append(assignees, assignee)
	}
	sort.Strings(assignees)

	var b strings.Builder
	for _, assignee := range assignees {
		b.WriteString(ws[assignee].Markdown(labelAllowlist))
	}

	return b.String()
}
