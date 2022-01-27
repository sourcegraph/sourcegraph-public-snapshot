package main

import (
	"regexp"
	"strings"
)

type acceptanceResult struct {
	Checked     bool
	Explanation string
}

var markdownCommentRegexp = regexp.MustCompile("<!--((.|\n)*?)-->(\n)*")

func checkAcceptance(body string) *acceptanceResult {
	const (
		acceptanceHeader        = "## Acceptance checklist"
		acceptanceChecklistItem = "I have gone through the [acceptance checklist]"
	)

	sections := strings.Split(body, acceptanceHeader)
	if len(sections) < 2 {
		return &acceptanceResult{Checked: false}
	}
	acceptanceSection := sections[1]
	if strings.Contains(acceptanceSection, acceptanceChecklistItem) && strings.Contains(acceptanceSection, "- [x] ") {
		return &acceptanceResult{Checked: true}
	}

	acceptanceLines := strings.Split(acceptanceSection, "\n")
	var explanation []string
	for _, l := range acceptanceLines {
		line := strings.TrimSpace(l)
		if !strings.Contains(line, acceptanceChecklistItem) {
			explanation = append(explanation, line)
		}
	}

	// Merge into single string
	fullExplanation := strings.Join(explanation, "\n")
	// Remove comments
	fullExplanation = markdownCommentRegexp.ReplaceAllString(fullExplanation, "")
	// Remove whitespace
	fullExplanation = strings.TrimSpace(fullExplanation)
	return &acceptanceResult{Checked: false, Explanation: fullExplanation}
}
