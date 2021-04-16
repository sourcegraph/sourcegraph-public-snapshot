package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/streaming"
)

var labelRegexp *regexp.Regexp

func init() {
	labelRegexp, _ = regexp.Compile("(?:\\[)(.*?)(?:])")
}

func streamSearch(query string, opts streaming.Opts, client api.Client, w io.Writer) error {
	var d streaming.Decoder
	if opts.Json {
		d = jsonDecoder(w)
	} else {
		t, err := parseTemplate(streamingTemplate)
		if err != nil {
			return err
		}
		d = textDecoder(query, t, w)
	}
	return streaming.Search(query, opts, client, d)
}

// jsonDecoder streams results as JSON to w.
func jsonDecoder(w io.Writer) streaming.Decoder {
	// write json.Marshals data and writes it as one line to w plus a newline.
	write := func(data interface{}) error {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("\n"))
		if err != nil {
			return err
		}
		return nil
	}

	return streaming.Decoder{
		OnProgress: func(progress *streaming.Progress) {
			if !progress.Done {
				return
			}
			err := write(progress)
			if err != nil {
				logError(err.Error())
			}
		},
		OnMatches: func(matches []streaming.EventMatch) {
			for _, match := range matches {
				err := write(match)
				if err != nil {
					logError(err.Error())
				}
			}
		},
		OnAlert: func(alert *streaming.EventAlert) {
			err := write(alert)
			if err != nil {
				logError(err.Error())
			}
		},
		OnError: func(eventError *streaming.EventError) {
			// Errors are just written to stderr.
			logError(eventError.Message)
		},
	}
}

func textDecoder(query string, t *template.Template, w io.Writer) streaming.Decoder {
	return streaming.Decoder{
		OnProgress: func(progress *streaming.Progress) {
			// We only show the final progress.
			if !progress.Done {
				return
			}

			templateData := struct {
				streaming.Progress
				LimitHit bool
			}{
				Progress: *progress,
				LimitHit: isLimitHit(progress),
			}

			err := t.ExecuteTemplate(w, "progress", &templateData)
			if err != nil {
				logError(fmt.Sprintf("error when executing template: %s\n", err))
			}
			return
		},
		OnError: func(eventError *streaming.EventError) {
			fmt.Printf("ERR: %s", eventError.Message)
		},
		OnAlert: func(alert *streaming.EventAlert) {
			proposedQueries := make([]ProposedQuery, len(alert.ProposedQueries))
			for _, pq := range alert.ProposedQueries {
				proposedQueries = append(proposedQueries, ProposedQuery{
					Description: pq.Description,
					Query:       pq.Query,
				})
			}

			err := t.ExecuteTemplate(w, "alert", searchResultsAlert{
				Title:           alert.Title,
				Description:     alert.Description,
				ProposedQueries: proposedQueries,
			})
			if err != nil {
				logError(fmt.Sprintf("error when executing template: %s\n", err))
				return
			}
		},
		OnMatches: func(matches []streaming.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *streaming.EventFileMatch:
					err := t.ExecuteTemplate(w, "file", struct {
						Query string
						*streaming.EventFileMatch
					}{
						Query:          query,
						EventFileMatch: match,
					},
					)
					if err != nil {
						logError(fmt.Sprintf("error when executing template: %s\n", err))
						return
					}
				case *streaming.EventRepoMatch:
					err := t.ExecuteTemplate(w, "repo", struct {
						SourcegraphEndpoint string
						*streaming.EventRepoMatch
					}{
						SourcegraphEndpoint: cfg.Endpoint,
						EventRepoMatch:      match,
					})
					if err != nil {
						logError(fmt.Sprintf("error when executing template: %s\n", err))
						return
					}
				case *streaming.EventCommitMatch:
					err := t.ExecuteTemplate(w, "commit", struct {
						SourcegraphEndpoint string
						*streaming.EventCommitMatch
					}{
						SourcegraphEndpoint: cfg.Endpoint,
						EventCommitMatch:    match,
					})
					if err != nil {
						logError(fmt.Sprintf("error when executing template: %s\n", err))
						return
					}
				case *streaming.EventSymbolMatch:
					err := t.ExecuteTemplate(w, "symbol", struct {
						SourcegraphEndpoint string
						*streaming.EventSymbolMatch
					}{
						SourcegraphEndpoint: cfg.Endpoint,
						EventSymbolMatch:    match,
					},
					)
					if err != nil {
						logError(fmt.Sprintf("error when executing template: %s\n", err))
						return
					}
				}
			}
		},
	}
}

// isLimitHit returns true if any of the skipped reasons indicate a limit was
// hit. This is the same logic we use in the webapp.
func isLimitHit(progress *streaming.Progress) bool {
	for _, p := range progress.Skipped {
		if strings.Contains(string(p.Reason), "-limit") {
			return true
		}
	}
	return false
}

const streamingTemplate = `
{{define "file"}}
	{{- /* Repository and file name */ -}}
	{{- color "search-repository"}}{{.Repository}}{{color "nc" -}}
	{{- " › " -}}
	{{- color "search-filename"}}{{.Path}}{{color "nc" -}}
	{{- color "success"}}{{matchOrMatches (len .LineMatches)}}{{color "nc" -}}
	{{- "\n" -}}
	{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}
	
	{{- /* Line matches */ -}}
	{{- $lineMatches := .LineMatches -}}
	{{- range $index, $match := $lineMatches -}}
		{{- if not (streamSearchSequentialLineNumber $lineMatches $index) -}}
			{{- color "search-border"}}{{"  ------------------------------------------------------------------------------\n"}}{{color "nc"}}
		{{- end -}}
		{{- "  "}}{{color "search-line-numbers"}}{{pad (addInt32 $match.LineNumber 1) 6 " "}}{{color "nc" -}}
		{{- color "search-border"}}{{" |  "}}{{color "nc"}}{{streamSearchHighlightMatch $.Query $match }}
	{{- end -}}
	{{- "\n" -}}
{{- end -}}

{{define "symbol"}}
	{{- /* Repository and file name */ -}}
	{{- color "search-repository"}}{{.Repository}}{{color "nc" -}}
	{{- " › " -}}
	{{- color "search-filename"}}{{.Path}}{{color "nc" -}}
	{{- color "success"}}{{matchOrMatches (len .Symbols)}}{{color "nc" -}}
	{{- "\n" -}}
	{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}
	
	{{- /* Symbols */ -}}
	{{- $symbols := .Symbols -}}
	{{- range $index, $match := $symbols -}}
		{{- color "success"}}{{.Name}}{{color "nc" -}} ({{.Kind}}{{if .ContainerName}}{{printf ", %s" .ContainerName}}{{end}})
		{{- color "search-border"}}{{" ("}}{{color "nc" -}}
		{{- color "search-repository"}}{{$.SourcegraphEndpoint}}/{{$match.URL}}{{color "nc" -}}
		{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
	{{- end -}}
	{{- "\n" -}}
{{- end -}}

{{define "repo"}}
	{{- /* Link to the result */ -}}
	{{- color "success"}}{{.Repository}}{{color "nc" -}}
	{{- color "search-border"}}{{" ("}}{{color "nc" -}}
	{{- color "search-repository"}}{{$.SourcegraphEndpoint}}/{{.Repository}}{{color "nc" -}}
	{{- color "search-border"}}{{")"}}{{color "nc" -}}
	{{- color "success"}}{{" ("}}{{"1 match)"}}{{color "nc" -}}
	{{- "\n" -}}
{{- end -}}

{{define "commit"}}
	{{- /* Link to the result */ -}}
	{{- color "search-border"}}{{"("}}{{color "nc" -}}
	{{- color "search-link"}}{{$.SourcegraphEndpoint}}{{.URL}}{{color "nc" -}}
	{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
	{{- color "nc" -}}
	
	{{- /* Repository > author name "commit subject" (time ago) */ -}}
	{{- color "search-commit-subject"}}{{(streamSearchRenderCommitLabel .Label)}}{{color "nc" -}}
	{{- color "success" -}}
		{{- if (len .Ranges) -}}
			{{matchOrMatches (len .Ranges)}}
		{{- else -}}
			{{matchOrMatches 1}}
		{{- end -}}
	{{- color "nc" -}}
	{{- "\n" -}}
	{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}
	{{- color "search-border"}}{{color "nc"}}{{indent (streamSearchHighlightCommit .Content .Ranges) "  "}}
{{end}}

{{define "alert"}}
	{{- searchAlertRender . -}}
{{end}}

{{define "progress"}}
	{{- color "logo" -}}✱{{- color "nc" -}}
	{{- " " -}}
	{{- if eq .MatchCount 0 -}}
		{{- color "warning" -}}
	{{- else -}}
		{{- color "success" -}}
	{{- end -}}
	{{- .MatchCount -}}{{if .LimitHit}}+{{end}} results{{- color "nc" -}}
	{{- " in " -}}{{color "success"}}{{msDuration .DurationMs}}{{if .RepositoriesCount}}{{- color "nc" -}}
	{{- " from " -}}{{color "success"}}{{.RepositoriesCount}}{{- " Repositories" -}}{{- color "nc" -}}{{end}}
	{{- "\n" -}}
	{{if len .Skipped}}
		{{- "\n" -}}
		{{- "Some results excluded:" -}}
		{{- "\n" -}}
		{{- range $index, $skipped := $.Skipped -}}
			{{indent $skipped.Title "    "}}{{- "\n" -}}
		{{- end -}}
	{{- end -}}
{{- end -}}
`

var streamSearchTemplateFuncs = map[string]interface{}{
	"streamSearchHighlightMatch": func(query string, match streaming.EventLineMatch) string {
		var highlights []highlight
		if strings.Contains(query, "patterntype:structural") {
			highlights = streamConvertMatchToHighlights(match, false)
			return applyHighlightsForFile(match.Line, highlights)
		}

		highlights = streamConvertMatchToHighlights(match, true)
		return applyHighlights(match.Line, highlights, ansiColors["search-match"], ansiColors["nc"])
	},

	"streamSearchSequentialLineNumber": func(lineMatches []streaming.EventLineMatch, index int) bool {
		prevIndex := index - 1
		if prevIndex < 0 {
			return true
		}
		prevLineNumber := lineMatches[prevIndex].LineNumber
		lineNumber := lineMatches[index].LineNumber
		return prevLineNumber == lineNumber-1
	},

	"streamSearchHighlightCommit": func(content string, ranges [][3]int32) string {
		highlights := make([]highlight, len(ranges))
		for _, r := range ranges {
			highlights = append(highlights, highlight{
				line:      int(r[0]),
				character: int(r[1]),
				length:    int(r[2]),
			})
		}
		if strings.HasPrefix(content, "```diff") {
			return streamSearchHighlightDiffPreview(content, highlights)
		}
		return applyHighlights(stripMarkdownMarkers(content), highlights, ansiColors["search-match"], ansiColors["nc"])
	},

	"streamSearchRenderCommitLabel": func(label string) string {
		m := labelRegexp.FindAllStringSubmatch(label, -1)
		if len(m) != 3 || len(m[0]) < 2 || len(m[1]) < 2 || len(m[2]) < 2 {
			return label
		}
		return m[0][1] + " > " + m[1][1] + " : " + m[2][1]
	},

	"matchOrMatches": func(i int) string {
		if i == 1 {
			return " (1 match)"
		}
		return fmt.Sprintf(" (%d matches)", i)
	},
}

func streamSearchHighlightDiffPreview(diffPreview string, highlights []highlight) string {
	useColordiff, err := strconv.ParseBool(os.Getenv("COLORDIFF"))
	if err != nil {
		useColordiff = true
	}
	if colorDisabled || !useColordiff {
		// Only highlight the matches.
		return applyHighlights(stripMarkdownMarkers(diffPreview), highlights, ansiColors["search-match"], ansiColors["nc"])
	}
	path, err := exec.LookPath("colordiff")
	if err != nil {
		// colordiff not installed; only highlight the matches.
		return applyHighlights(stripMarkdownMarkers(diffPreview), highlights, ansiColors["search-match"], ansiColors["nc"])
	}

	// First highlight the matches, but use a special "end of match" token
	// instead of no color (so that we don'streamingTemplate terminate colors that colordiff
	// adds).
	uniqueStartOfMatchToken := "pXRdMhZbgnPL355429nsO4qFgX86LfXTSmqH4Nr3#*(@)!*#()@!APPJB8ZRutvZ5fdL01273i6OdzLDm0UMC9372891skfJTl2c52yR1v"
	uniqueEndOfMatchToken := "v1Ry25c2lTJfks1982739CMU0mDLzdO6i37210Ldf5ZvtuRZ8BJPPA!@)(#*!)@(*#3rN4HqmSTXfL68XgFq4Osn924553LPngbZhMdRXp"
	diff := applyHighlights(stripMarkdownMarkers(diffPreview), highlights, uniqueStartOfMatchToken, uniqueEndOfMatchToken)

	// Now highlight our diff with colordiff.
	var buf bytes.Buffer
	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader(diff)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		fmt.Println("warning: colordiff failed to colorize diff:", err)
		return diff
	}
	colorized := buf.String()
	var final []string
	for _, line := range strings.Split(colorized, "\n") {
		// fmt.Println("LINE", line)
		// Find where the start-of-match token is in the line.
		somToken := strings.Index(line, uniqueStartOfMatchToken)

		// Find which ANSI codes are to the left of our start-of-match token.
		indices := ansiRegexp.FindAllStringIndex(line, -1)
		matches := ansiRegexp.FindAllString(line, -1)
		var left []string
		for k, index := range indices {
			if index[0] < somToken && index[1] < somToken {
				left = append(left, matches[k])
			}
		}

		// Replace our start-of-match token with the color we wish.
		line = strings.Replace(line, uniqueStartOfMatchToken, ansiColors["search-match"], -1)

		// Replace our end-of-match token with the color terminator,
		// and start all colors that were previously started to the left.
		line = strings.Replace(line, uniqueEndOfMatchToken, ansiColors["nc"]+strings.Join(left, ""), -1)

		final = append(final, line)
	}
	return strings.Join(final, "\n")
}

func stripMarkdownMarkers(content string) string {
	content = strings.TrimLeft(content, "```COMMIT_EDITMSG\n")
	content = strings.TrimLeft(content, "```diff\n")
	return strings.TrimRight(content, "\n```")
}

// convertMatchToHighlights converts a FileMatch m to a highlight data type.
// When isPreview is true, it is assumed that the result to highlight is only on
// one line, and the offsets are relative to this line. When isPreview is false,
// the lineNumber from the FileMatch data is used, which is relative to the file
// content.
func streamConvertMatchToHighlights(m streaming.EventLineMatch, isPreview bool) (highlights []highlight) {
	var line int
	for _, offsetAndLength := range m.OffsetAndLengths {
		ol := offsetAndLength
		offset := int(ol[0])
		length := int(ol[1])
		if isPreview {
			line = 1
		} else {
			line = int(m.LineNumber)
		}
		highlights = append(highlights, highlight{line: line, character: offset, length: length})
	}
	return highlights
}

func logError(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, msg)
}
