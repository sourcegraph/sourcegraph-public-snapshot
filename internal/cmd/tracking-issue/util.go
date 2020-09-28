package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Days(estimate string) float64 {
	d, _ := strconv.ParseFloat(strings.TrimSuffix(estimate, "d"), 64)
	return d
}

func Estimate(labels []string) string {
	const prefix = "estimate/"
	for _, label := range labels {
		if strings.HasPrefix(label, prefix) {
			return label[len(prefix):]
		}
	}
	return ""
}

func Emojis(categories map[string]string) string {
	sorted := make([]string, 0, len(categories))
	length := 0

	for _, emoji := range categories {
		sorted = append(sorted, emoji)
		length += len(emoji)
	}

	sort.Strings(sorted)

	s := make([]byte, 0, length)
	for _, emoji := range sorted {
		s = append(s, emoji...)
	}

	return string(s)
}

var customerMatcher = regexp.MustCompile(`https://app\.hubspot\.com/contacts/2762526/company/\d+`)

func Customer(body string) string {
	customer := customerMatcher.FindString(body)
	if customer == "" {
		return "👩"
	}
	return "[👩](" + customer + ")"
}

func Assignee(assignees []string) string {
	if len(assignees) == 0 {
		return "Unassigned"
	}

	return assignees[0]
}

func ListOfAssignees(assignees []string) []string {
	if len(assignees) == 0 {
		return []string{"Unassigned"}
	}

	return assignees
}

func RedactLabels(labels []string) []string {
	redacted := labels[:0]
	for _, label := range labels {
		if strings.HasPrefix(label, "estimate/") || strings.HasPrefix(label, "planned/") {
			redacted = append(redacted, label)
		}
	}
	return redacted
}

func Categories(labels []string, repository, body string) map[string]string {
	categories := make(map[string]string, len(labels))

	switch repository {
	case "sourcegraph/customer":
		categories["customer"] = Customer(body)
	case "sourcegraph/security-prs":
		categories["security"] = Emoji("security")
	}

	for _, label := range labels {
		if label == "customer" {
			categories[label] = Customer(body)
		} else if emoji := Emoji(label); emoji != "" {
			categories[label] = emoji
		}
	}

	return categories
}

func Emoji(category string) string {
	switch category {
	case "roadmap":
		return "🛠️"
	case "debt":
		return "🧶"
	case "spike":
		return "🕵️"
	case "quality-of-life":
		return "🎩"
	case "bug":
		return "🐛"
	case "security":
		return "🔒"
	default:
		return ""
	}
}

func contains(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}

	return false
}

// now returns the current time for relative formatting. This
// is overwritten during tests to ensure that our output can be
// byte-for-byte compared against the golden output file.
var now = time.Now

func formatTimeSince(t time.Time) string {
	days := now().UTC().Sub(t.UTC()) / time.Hour / 24

	switch days {
	case 0:
		return "today"
	case 1:
		return "1 day ago"
	default:
		return fmt.Sprintf("%d days ago", days)
	}
}
