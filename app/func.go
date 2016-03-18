package app

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/util/timeutil"
)

func minTime(a, b interface{}) time.Time {
	at, bt := timeutil.TimeOrNil(a), timeutil.TimeOrNil(b)
	if at == nil && bt == nil {
		return time.Time{}
	}
	if at == nil {
		return *bt
	}
	if bt == nil {
		return *at
	}
	if at.Before(*bt) {
		return *at
	}
	return *bt
}

func duration(v0, v1 interface{}) string {
	t0, t1 := timeutil.TimeOrNil(v0), timeutil.TimeOrNil(v1)
	if t0 == nil || t1 == nil {
		return "n/a"
	}
	d := t1.Sub(*t0)
	return roundToMsec(d).String()
}

func roundToMsec(d time.Duration) time.Duration {
	return (d / time.Millisecond) * time.Millisecond
}

func maxLen(maxLen int, s string) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	return reflect.ValueOf(v).IsNil()
}

// truncateCommitID truncates commit IDs to 6 chars.
func truncateCommitID(id interface{}) (string, error) {
	var idStr string
	if s, ok := id.(vcs.CommitID); ok {
		idStr = string(s)
	} else {
		idStr = id.(string)
	}
	if len(idStr) != 40 {
		return "", fmt.Errorf("truncateCommitID: got %q, expected full 40-char commit ID", idStr)
	}
	return idStr[:6], nil
}

// commitSummary returns the git commit summary from the full commit.
func commitSummary(message string) string {
	summary, _ := splitCommitMessage(message)
	return summary
}

// commitRestOfMessage returns the commit body excluding the summary (returned
// by commitSummary).
func commitRestOfMessage(message string) string {
	_, rest := splitCommitMessage(message)
	return rest
}

// splitCommitMessage splits a commit message into the summary (for commitSummary) and the rest (for commitRestOfMessage).
func splitCommitMessage(message string) (summary, rest string) {
	parts := strings.SplitN(message, "\n\n", 2)
	summary = parts[0]
	if len(parts) > 1 {
		rest = parts[1]
	}
	return summary, rest
}

// hasStructField returns true if v is a struct that contains a field with
// the given name, or is a pointer to such a struct.
func hasStructField(v interface{}, field string) bool {
	_, exists := getStructField(v, field)
	return exists
}

// getStructField returns the value of v's field with the given name
// if it exists. v must be a struct or a pointer to a struct.
func getStructField(v interface{}, field string) (fieldVal interface{}, exists bool) {
	vv := reflect.ValueOf(v)
	if !vv.IsValid() {
		return nil, false
	}
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	fv := vv.FieldByName(field)
	if !fv.IsValid() {
		return nil, false
	}
	return fv.Interface(), true
}
