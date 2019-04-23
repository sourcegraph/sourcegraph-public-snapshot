package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"jaytaylor.com/html2text"
)

var dateRegex = regexp.MustCompile("(\\w{4}-\\w{2}-\\w{2})")

func search() error {
	jsonFlag := flag.Bool("json", false, "Whether or not to output results as JSON")
	flag.Parse()
	if flag.NArg() != 1 {
		return &usageError{errors.New("expected exactly one argument: the search query")}
	}
	queryString := flag.Arg(0)

	query := `fragment FileMatchFields on FileMatch {
				repository {
					name
					url
				}
				file {
					name
					path
					url
					commit {
						oid
					}
				}
				lineMatches {
					preview
					lineNumber
					offsetAndLengths
					limitHit
				}
			}

			fragment CommitSearchResultFields on CommitSearchResult {
				messagePreview {
					value
					highlights{
						line
						character
						length
					}
				}
				diffPreview {
					value
					highlights {
						line
						character
						length
					}
				}
				label {
					html
				}
				url
				matches {
					url
					body {
						html
						text
					}
					highlights {
						character
						line
						length
					}
				}
				commit {
					repository {
						name
					}
					oid
					url
					subject
					author {
						date
						person {
							displayName
						}
					}
				}
			}

		  fragment RepositoryFields on Repository {
			name
			url
			externalURLs {
			  serviceType
			  url
			}
			label {
				html
			}
		  }

		  query ($query: String!) {
			site {
				buildVersion
			}
			search(query: $query) {
			  results {
				results{
				  __typename
				  ... on FileMatch {
					...FileMatchFields
				  }
				  ... on CommitSearchResult {
					...CommitSearchResultFields
				  }
				  ... on Repository {
					...RepositoryFields
				  }
				}
				limitHit
				cloning {
				  name
				}
				missing {
				  name
				}
				timedout {
				  name
				}
				resultCount
				elapsedMilliseconds
			  }
			}
		  }
		`

	var result struct {
		Site struct {
			BuildVersion string
		}
		Search struct {
			Results searchResults
		}
	}

	// Parse config.
	cfg, err := readConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	return (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"query": nullString(queryString),
		},
		result: &result,
		done: func() error {
			improved := searchResultsImproved{
				SourcegraphEndpoint: cfg.Endpoint,
				Query:               queryString,
				Site:                result.Site,
				searchResults:       result.Search.Results,
			}

			if *jsonFlag {
				// Print the formatted JSON.
				f, err := marshalIndent(improved)
				if err != nil {
					return err
				}
				fmt.Println(string(f))
				return nil
			}

			tmpl, err := parseTemplate(searchResultsTemplate)
			if err != nil {
				return err
			}
			if err := execTemplate(tmpl, improved); err != nil {
				return err
			}
			return nil
		},
		endpoint: cfg.Endpoint,
		accessToken: cfg.AccessToken,
	}).do()
}

// searchResults represents the data we get back from the GraphQL search request.
type searchResults struct {
	Results                    []map[string]interface{}
	LimitHit                   bool
	Cloning, Missing, Timedout []map[string]interface{}
	ResultCount                int
	ElapsedMilliseconds        int
}

// searchResultsImproved is a superset of what the GraphQL API returns. It
// contains the query responsible for producing the results, which is nice for
// most consumers.
type searchResultsImproved struct {
	SourcegraphEndpoint string
	Query               string
	Site                struct{ BuildVersion string }
	searchResults
}

func envSetDefault(env []string, key, value string) []string {
	set := false
	for _, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			set = true
			break
		}
	}
	if !set {
		env = append(env, key+"="+value)
	}
	return env
}

func searchHighlightPreview(preview interface{}, start, end string) string {
	if start == "" {
		start = ansiColors["search-match"]
	}
	if end == "" {
		end = ansiColors["nc"]
	}
	p := preview.(map[string]interface{})
	value := p["value"].(string)
	var highlights []highlight
	for _, highlightObject := range p["highlights"].([]interface{}) {
		h := highlightObject.(map[string]interface{})
		line := int(h["line"].(float64))
		character := int(h["character"].(float64))
		length := int(h["length"].(float64))
		highlights = append(highlights, highlight{line, character, length})
	}
	return applyHighlights(value, highlights, start, end)
}

func searchHighlightDiffPreview(diffPreview interface{}) string {
	p := diffPreview.(map[string]interface{})
	diff := p["value"].(string)

	useColordiff, err := strconv.ParseBool(os.Getenv("COLORDIFF"))
	if err != nil {
		useColordiff = true
	}
	if colorDisabled || !useColordiff {
		// Only highlight the matches.
		return searchHighlightPreview(diffPreview, "", "")
	}
	path, err := exec.LookPath("colordiff")
	if err != nil {
		// colordiff not installed; only highlight the matches.
		return searchHighlightPreview(diffPreview, "", "")
	}

	// First highlight the matches, but use a special "end of match" token
	// instead of no color (so that we don't terminate colors that colordiff
	// adds).
	uniqueStartOfMatchToken := "pXRdMhZbgnPL355429nsO4qFgX86LfXTSmqH4Nr3#*(@)!*#()@!APPJB8ZRutvZ5fdL01273i6OdzLDm0UMC9372891skfJTl2c52yR1v"
	uniqueEndOfMatchToken := "v1Ry25c2lTJfks1982739CMU0mDLzdO6i37210Ldf5ZvtuRZ8BJPPA!@)(#*!)@(*#3rN4HqmSTXfL68XgFq4Osn924553LPngbZhMdRXp"
	diff = searchHighlightPreview(diffPreview, uniqueStartOfMatchToken, uniqueEndOfMatchToken)

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

func htmlToPlainText(input string) string {
	text, err := html2text.FromString(input, html2text.Options{OmitLinks: true})
	text = html.UnescapeString(text)
	if err != nil {
		return input
	}
	return text
}

// Checks the Sourcegraph instance's build version date to determine if the new search result interface exists.
func buildVersionHasNewSearchInterface(buildVersion string) bool {
	buildDate := dateRegex.FindString(buildVersion)
	t, err := time.Parse("2006-01-02", buildDate)
	if err != nil {
		return false
	}
	newSearchBuildDate, err := time.Parse("2006-01-02", "2018-12-10")
	if err != nil {
		return false
	}
	isAfterNewSearchDate := t.After(newSearchBuildDate)
	return isAfterNewSearchDate
}

type highlight struct {
	line      int // the 1-indexed line number
	character int // the 1-indexed character on the line.
	length    int // the 1-indexed length of the highlight, in characters.
}

func applyHighlights(input string, highlights []highlight, start, end string) string {
	var result []rune
	lines := strings.Split(input, "\n")
	for lineNumber, line := range lines {
		lineNumber++
		for characterIndex, character := range []rune(line + "\n") {
			for _, highlight := range highlights {
				if highlight.line == lineNumber {
					if characterIndex == highlight.character {
						result = append(result, []rune(start)...)
					} else if characterIndex == highlight.character+highlight.length {
						result = append(result, []rune(end)...)
					}
				}
			}
			result = append(result, character)
		}
	}
	return string(result)
}

var searchTemplateFuncs = map[string]interface{}{
	"searchSequentialLineNumber": func(lineMatches []interface{}, index int) bool {
		prevIndex := index - 1
		if prevIndex < 0 {
			return true
		}
		prevLineNumber := lineMatches[prevIndex].(map[string]interface{})["lineNumber"]
		lineNumber := lineMatches[index].(map[string]interface{})["lineNumber"]
		return prevLineNumber.(float64) == lineNumber.(float64)-1
	},
	"searchHighlightMatch": func(match interface{}) string {
		m := match.(map[string]interface{})
		preview := m["preview"].(string)
		var highlights []highlight
		for _, offsetAndLength := range m["offsetAndLengths"].([]interface{}) {
			ol := offsetAndLength.([]interface{})
			offset := int(ol[0].(float64))
			length := int(ol[1].(float64))
			highlights = append(highlights, highlight{line: 1, character: offset, length: length})
		}
		return applyHighlights(preview, highlights, ansiColors["search-match"], ansiColors["nc"])
	},
	"searchHighlightPreview": func(preview interface{}) string {
		return searchHighlightPreview(preview, "", "")
	},
	"searchHighlightDiffPreview": func(diffPreview interface{}) string {
		return searchHighlightDiffPreview(diffPreview)
	},
	"searchMaxRepoNameLength": func(results []map[string]interface{}) int {
		max := 0
		for _, r := range results {
			if r["__typename"] != "Repository" {
				continue
			}
			if name := r["name"].(string); len(name) > max {
				max = len(name)
			}
		}
		return max
	},
	"htmlToPlainText": func(input string) string {
		return htmlToPlainText(input)
	},
	"buildVersionHasNewSearchInterface": func(input string) bool {
		return buildVersionHasNewSearchInterface(input)
	},
	"renderResult": func(searchResult map[string]interface{}) string {
		searchResultBody := searchResult["body"].(map[string]interface{})
		html := searchResultBody["html"].(string)
		markdown := searchResultBody["text"].(string)
		plainText := htmlToPlainText(html)
		highlights := searchResult["highlights"]
		isDiff := strings.HasPrefix(markdown, "```diff") && strings.HasSuffix(markdown, "```")
		if _, ok := highlights.([]interface{}); ok {
			if isDiff {
				// We special case diffs because we want to display them with color.
				return searchHighlightDiffPreview(map[string]interface{}{"value": plainText, "highlights": highlights.([]interface{})})
			}
			return searchHighlightPreview(map[string]interface{}{"value": plainText, "highlights": highlights.([]interface{})}, "", "")
		}
		return markdown
	},
}

const searchResultsTemplate = `{{- /* ignore this line for template formatting sake */ -}}

{{- /* The first results line */ -}}
	{{- color "logo" -}}✱{{- color "nc" -}}
	{{- " " -}}
	{{- if eq .ResultCount 0 -}}
		{{- color "warning" -}}
	{{- else -}}
		{{- color "success" -}}
	{{- end -}}
	{{- .ResultCount -}}{{if .LimitHit}}+{{end}} results{{- color "nc" -}}
	{{- " for " -}}{{- color "search-query"}}"{{.Query}}"{{color "nc" -}}
	{{- " in " -}}{{color "success"}}{{msDuration .ElapsedMilliseconds}}{{color "nc" -}}

{{- /* The cloning / missing / timed out repos warnings */ -}}
	{{- with .Cloning}}{{color "warning"}}{{"\n"}}({{len .}}) still cloning:{{color "nc"}} {{join (repoNames .) ", "}}{{end -}}
	{{- with .Missing}}{{color "warning"}}{{"\n"}}({{len .}}) missing:{{color "nc"}} {{join (repoNames .) ", "}}{{end -}}
	{{- with .Timedout}}{{color "warning"}}{{"\n"}}({{len .}}) timed out:{{color "nc"}} {{join (repoNames .) ", "}}{{end -}}
	{{"\n"}}

{{- /* Rendering of results */ -}}
	{{- range .Results -}}
		{{- if ne .__typename "Repository" -}}
			{{- /* The border separating results */ -}}
			{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}
		{{- end -}}

		{{- /* File match rendering. */ -}}
		{{- if eq .__typename "FileMatch" -}}
			{{- /* Link to the result */ -}}
			{{- color "search-border"}}{{"("}}{{color "nc" -}}
			{{- color "search-link"}}{{$.SourcegraphEndpoint}}{{.file.url}}{{color "nc" -}}
			{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
			{{- color "nc" -}}

			{{- /* Repository and file name */ -}}
			{{- color "search-repository"}}{{.repository.name}}{{color "nc" -}}
			{{- " › " -}}
			{{- color "search-filename"}}{{.file.name}}{{color "nc" -}}
			{{- color "success"}}{{" ("}}{{len .lineMatches}}{{" matches)"}}{{color "nc" -}}
			{{- "\n" -}}
			{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}

			{{- /* Line matches */ -}}
			{{- $lineMatches := .lineMatches -}}
			{{- range $index, $match := $lineMatches -}}
				{{- if not (searchSequentialLineNumber $lineMatches $index) -}}
					{{- color "search-border"}}{{"  ------------------------------------------------------------------------------\n"}}{{color "nc"}}
				{{- end -}}
				{{- "  "}}{{color "search-line-numbers"}}{{pad (addFloat $match.lineNumber 1) 6 " "}}{{color "nc" -}}
				{{- color "search-border"}}{{" |  "}}{{color "nc"}}{{searchHighlightMatch $match}}
			{{- end -}}
		{{- end -}}

		{{- /* Commit (type:diff, type:commit) result rendering for Sourcegraph instances after 2.13.x. */ -}}
		{{- if and (eq .__typename "CommitSearchResult") (buildVersionHasNewSearchInterface $.Site.BuildVersion) -}}
			{{- /* Link to the result */ -}}
			{{- color "search-border"}}{{"("}}{{color "nc" -}}
			{{- color "search-link"}}{{$.SourcegraphEndpoint}}{{.url}}{{color "nc" -}}
			{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
			{{- color "nc" -}}

			{{- /* Repository > author name "commit subject" (time ago) */ -}}
			{{- color "search-commit-subject"}}{{(htmlToPlainText .label.html)}}{{color "nc" -}}
			{{- "\n" -}}
			{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}
			{{- $matches := .matches -}}
			{{- range $index, $match := $matches -}}
				{{- color "search-border"}}{{color "nc"}}{{indent (renderResult $match) "  "}}
			{{- end -}}
		{{- end -}}

		{{- /* Commit (type:diff, type:commit) result rendering for Sourcegraph instances on and before 2.13.x. */ -}}
		{{- if and (eq .__typename "CommitSearchResult") (not (buildVersionHasNewSearchInterface $.Site.BuildVersion)) -}}
			{{- /* Link to the result */ -}}
			{{- color "search-border"}}{{"("}}{{color "nc" -}}
			{{- color "search-link"}}{{$.SourcegraphEndpoint}}{{.commit.url}}{{color "nc" -}}
			{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
			{{- color "nc" -}}

			{{- /* Repository > author name "commit subject" (time ago) */ -}}
			{{- color "search-repository"}}{{.commit.repository.name}}{{color "nc" -}}
			{{- " › " -}}
			{{- color "search-commit-author"}}{{.commit.author.person.displayName}}{{color "nc" -}}
			{{- " " -}}
			{{- color "search-commit-subject"}}"{{.commit.subject}}"{{color "nc" -}}
			{{- " "}}
			{{- color "search-commit-date"}}{{"("}}{{humanizeRFC3339 .commit.author.date}}{{")" -}}{{color "nc" -}}
			{{- "\n" -}}
			{{- color "search-border"}}{{"--------------------------------------------------------------------------------\n"}}{{color "nc"}}

			{{- if .messagePreview -}}
				{{- /* type:commit rendering */ -}}
				{{indent (searchHighlightPreview .messagePreview) "  "}}
			{{- end -}}
			{{- if .diffPreview -}}
				{{- /* type:diff rendering */ -}}
				{{indent (searchHighlightDiffPreview .diffPreview) "  "}}
			{{- end -}}
		{{- end -}}

		{{- /* Repository (type:repo) result rendering for Sourcegraph instances after 2.13.x. */ -}}
		{{- if and (eq .__typename "Repository") (buildVersionHasNewSearchInterface $.Site.BuildVersion) -}}
			{{- /* Link to the result */ -}}
			{{- color "success"}}{{padRight (htmlToPlainText .label.html) (searchMaxRepoNameLength $.Results) " "}}{{color "nc" -}}
			{{- color "search-border"}}{{" ("}}{{color "nc" -}}
			{{- color "search-repository"}}{{$.SourcegraphEndpoint}}{{.url}}{{color "nc" -}}
			{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
			{{- color "nc" -}}
		{{- end -}}

		{{- /* Repository (type:repo) result rendering for Sourcegraph instances on and before 2.13.x. */ -}}
		{{- if and (eq .__typename "Repository") (not (buildVersionHasNewSearchInterface $.Site.BuildVersion)) -}}
			{{- /* Link to the result */ -}}
			{{- color "success"}}{{padRight .name (searchMaxRepoNameLength $.Results) " "}}{{color "nc" -}}
			{{- color "search-border"}}{{" ("}}{{color "nc" -}}
			{{- color "search-repository"}}{{$.SourcegraphEndpoint}}{{.url}}{{color "nc" -}}
			{{- color "search-border"}}{{")\n"}}{{color "nc" -}}
			{{- color "nc" -}}
		{{- end -}}
	{{- end -}}
`
