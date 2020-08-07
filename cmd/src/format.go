package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
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

		// `src campaign patchset create-from-patches`
		"friendlyPatchSetCreatedMessage": func(patchSet PatchSet) string {
			var buf bytes.Buffer
			fmt.Fprintln(&buf)
			fmt.Fprintln(&buf, color.HiGreenString("✔  Patch set saved."), "\n\nPreview and create a campaign on Sourcegraph using one of the following options:")
			fmt.Fprintln(&buf)
			fmt.Fprintln(&buf, " ", color.HiCyanString("▶ Web:"), patchSet.PreviewURL, color.HiBlackString("or"))
			cliCommand := fmt.Sprintf("src campaigns create -patchset=%s -branch=DESIRED-BRANCH-NAME", patchSet.ID)
			fmt.Fprintln(&buf, " ", color.HiCyanString("▶ CLI:"), cliCommand)

			// Hacky to do this in a formatting helper, but better than
			// globally querying the version and only using it here for now.
			version, err := getSourcegraphVersion(context.Background(), cfg.apiClient(nil, os.Stdout))
			if err != nil {
				// We ignore the error and return what we have
				return buf.String()
			}

			supportsUpdatingPatchSet, err := sourcegraphVersionCheck(version, ">= 3.13-0", "2020-02-14")
			if err != nil {
				return buf.String()
			}

			if supportsUpdatingPatchSet {
				fmt.Fprintln(&buf, "\nTo update an existing campaign using this patch set:")
				fmt.Fprintln(&buf, "\n ", color.HiCyanString("▶ Web:"), strings.Replace(patchSet.PreviewURL, "/new", "/update", 1))
			}

			return buf.String()
		},

		// `src campaign create`
		"friendlyCampaignCreatedMessage": func(campaign Campaign) string {
			var buf bytes.Buffer
			fmt.Fprintln(&buf)

			message := "See the progress of changeset creation on code hosts:"
			if campaign.PublishedAt.IsZero() {
				message = "Publish the campaign and all of its changesets or single changesets individually to create pull requests on code hosts:"
			}

			fmt.Fprintln(&buf, color.HiGreenString("✔  Campaign created."), message)
			fmt.Fprintln(&buf)

			u, err := resolveURL(cfg.Endpoint, campaign.URL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to resolve campaign URL: %s\n", err)
				return buf.String()
			}

			fmt.Fprintln(&buf, " ", color.HiCyanString("▶ Web:"), u)

			return buf.String()
		},

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

func resolveURL(endpoint, u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	base, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(parsed).String(), nil
}
