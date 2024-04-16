package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/sourcegraph/jsonx"
)

func parseTemplate(text string) (*template.Template, error) {
	tmpl := template.New("")
	tmpl.Funcs(map[string]interface{}{
		"join": strings.Join,
		"json": func(v interface{}) (string, error) {
			b, err := marshalIndent(v)
			return string(b), err
		},
		"jsonIndent": func(jsonStr string) (string, error) {
			return jsonx.ApplyEdits(jsonStr, jsonx.Format(jsonStr, jsonx.FormatOptions{TabSize: 2})...)
		},
		"msDuration": func(ms int) time.Duration {
			return time.Duration(ms) * time.Millisecond
		},
		"repoNames": func(repos []map[string]interface{}) (names []string) {
			for _, r := range repos {
				names = append(names, r["name"].(string))
			}
			return
		},
		"pad": func(value interface{}, padding int, padCharacter string) string {
			val := fmt.Sprint(value)
			repeat := padding - len(val)
			if repeat < 0 {
				repeat = 0
			}
			return strings.Repeat(padCharacter, repeat) + val
		},
		"padRight": func(value interface{}, padding int, padCharacter string) string {
			val := fmt.Sprint(value)
			repeat := padding - len(val)
			if repeat < 0 {
				repeat = 0
			}
			return val + strings.Repeat(padCharacter, repeat)
		},
		"indent": func(lines, indention string) string {
			split := strings.Split(lines, "\n")
			for i, l := range split {
				if l != "" {
					split[i] = indention + l
				}
			}
			return strings.Join(split, "\n")
		},
		"addFloat": func(x, y float64) float64 {
			return x + y
		},
		"debug": func(v interface{}) string {
			data, _ := marshalIndent(v)
			fmt.Println(string(data))

			// Template functions must return something. In our case, it is
			// useful to actually print the string above now as the template
			// could fail due to e.g. syntax errors that someone is trying to
			// debug,and we want the spew above to show regardless.
			return ""
		},
		"color": func(name string) string {
			return ansiColors[name]
		},
		"humanizeRFC3339": func(date string) (string, error) {
			t, err := time.Parse(time.RFC3339, date)
			if err != nil {
				return "", err
			}
			return humanize.Time(t), nil
		},

		// Register search-specific template functions
		"searchSequentialLineNumber":        searchTemplateFuncs["searchSequentialLineNumber"],
		"searchHighlightMatch":              searchTemplateFuncs["searchHighlightMatch"],
		"searchHighlightPreview":            searchTemplateFuncs["searchHighlightPreview"],
		"searchHighlightDiffPreview":        searchTemplateFuncs["searchHighlightDiffPreview"],
		"searchMaxRepoNameLength":           searchTemplateFuncs["searchMaxRepoNameLength"],
		"htmlToPlainText":                   searchTemplateFuncs["htmlToPlainText"],
		"buildVersionHasNewSearchInterface": searchTemplateFuncs["buildVersionHasNewSearchInterface"],
		"renderResult":                      searchTemplateFuncs["renderResult"],

		// Register stream-search specific template functions.
		"streamSearchHighlightMatch":    streamSearchTemplateFuncs["streamSearchHighlightMatch"],
		"streamSearchHighlightCommit":   streamSearchTemplateFuncs["streamSearchHighlightCommit"],
		"streamSearchRenderCommitLabel": streamSearchTemplateFuncs["streamSearchRenderCommitLabel"],
		"matchOrMatches":                streamSearchTemplateFuncs["matchOrMatches"],
		"countMatches":                  streamSearchTemplateFuncs["countMatches"],

		// Alert rendering
		"searchAlertRender": func(alert searchResultsAlert) string {
			if content, err := alert.Render(); err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering search alert: %v\n", err)
				return ""
			} else {
				return content
			}
		},
	})
	return tmpl.Parse(text)
}

func execTemplate(tmpl *template.Template, data interface{}) error {
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// json.MarshalIndent, but with defaults.
func marshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
