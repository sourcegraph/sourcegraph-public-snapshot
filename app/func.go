package app

import (
	"bytes"
	"fmt"
	htmpl "html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
)

func pluralizeWord(noun string, n int) string {
	if n == 1 {
		return noun
	}
	var plnoun string
	if noun == "All" {
		return noun
	} else if strings.HasSuffix(noun, "person") {
		plnoun = noun[:len(noun)-len("person")] + "people"
	} else if strings.HasSuffix(noun, "y") {
		plnoun = noun[:len(noun)-1] + "ies"
	} else if strings.HasSuffix(noun, "ss") || strings.HasSuffix(noun, "ch") {
		plnoun = noun + "es"
	} else {
		plnoun = noun + "s"
	}
	return plnoun
}

func pluralize(noun string, n int) string {
	return num(n) + " " + pluralizeWord(noun, n)
}

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

func number(n int) string { return humanize.Comma(int64(n)) }

// num abbreviates and rounds n. Examples: 150, 13.2K, 1.5K.
func num(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	} else if n < 30000 {
		s := fmt.Sprintf("%.1fk", float64(n)/1000)
		return strings.Replace(s, ".0k", "k", 1)
	} else if n < 500000 {
		return strconv.Itoa(n/1000) + "k"
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
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

// capFirst capitalizes the first letter (if any) of s.
func capFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[0:1]) + s[1:]
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

func joinEnglish(nouns []string) string {
	if len(nouns) >= 2 {
		return strings.Join(nouns[:len(nouns)-1], ", ") + " and " + nouns[len(nouns)-1]
	}
	return strings.Join(nouns, " and ")
}

// ifTemplate will look up the template name in the specified file, pass it
// the given data and return the result. If the template does not exist or
// any error occurs, ifTemplate returns an empty string.
// The file parameter can be obtained in the HTML templates via $.Common.TemplateName
func ifTemplate(file, name string, data interface{}) htmpl.HTML {
	f := tmpl.Get(file)
	if f == nil {
		return ""
	}
	t := f.Lookup(name)
	if t == nil {
		return ""
	}
	var buf bytes.Buffer
	t.Execute(&buf, data)
	return htmpl.HTML(buf.String())
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
