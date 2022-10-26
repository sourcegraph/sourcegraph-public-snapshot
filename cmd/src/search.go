package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"

	isatty "github.com/mattn/go-isatty"
	"jaytaylor.com/html2text"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
	"github.com/sourcegraph/src-cli/internal/streaming"
)

var dateRegex = regexp.MustCompile(`(\w{4}-\w{2}-\w{2})`)

func init() {
	usage := `
Examples:

  Perform a search and get results:

    	$ src search 'repogroup:sample error'

  Perform a search and get results as JSON:

    	$ src search -json 'repogroup:sample error'

Other tips:

  Make 'type:diff' searches have colored diffs by installing https://colordiff.org
    - Ubuntu/Debian: $ sudo apt-get install colordiff
    - Mac OS:        $ brew install colordiff
    - Windows:       $ npm install -g colordiff

  Disable color output by setting NO_COLOR=t (see https://no-color.org).

  Force color output on (not on by default when piped to other programs) by setting COLOR=t

  Query syntax: https://docs.sourcegraph.com/code_search/reference/queries

  Be careful with search strings including negation: a search with an initial
  negated term may be parsed as a flag rather than as a search string. You can
  use -- to ensure that src parses this correctly, eg:

    	$ src search -- '-repo:github.com/foo/bar error'
`

	flagSet := flag.NewFlagSet("search", flag.ExitOnError)
	var (
		jsonFlag        = flagSet.Bool("json", false, "Whether or not to output results as JSON.")
		explainJSONFlag = flagSet.Bool("explain-json", false, "Explain the JSON output schema and exit.")
		apiFlags        = api.NewFlags(flagSet)
		lessFlag        = flagSet.Bool("less", true, "Pipe output to 'less -R' (only if stdout is terminal, and not json flag).")
		streamFlag      = flagSet.Bool("stream", false, "Consume results as stream. Streaming search only supports a subset of flags and parameters: trace, insecure-skip-verify, display, json.")
		display         = flagSet.Int("display", -1, "Limit the number of results that are displayed. Only supported together with stream flag. Statistics continue to report all results.")
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *streamFlag {
			opts := streaming.Opts{
				Display: *display,
				Trace:   apiFlags.Trace(),
				Json:    *jsonFlag,
			}
			client := cfg.apiClient(apiFlags, flagSet.Output())
			return streamSearch(flagSet.Arg(0), opts, client, os.Stdout)
		}

		if *explainJSONFlag {
			fmt.Printf("%s\n", searchJSONExplanation)
			return nil
		}

		if flagSet.NArg() != 1 {
			return cmderrors.Usage("expected exactly one argument: the search query")
		}
		queryString := flagSet.Arg(0)

		// For pagination, pipe our own output to 'less -R'
		if *lessFlag && !*jsonFlag {
			// But first we check whether we can use `less`. (Instead of
			// combining the conditions here into one, we use a 2nd conditional
			// so we don't need to do `exec.LookPath` if flags disable `less`)
			_, err := exec.LookPath("less")
			if err == nil && isatty.IsTerminal(os.Stdout.Fd()) {
				cmdPath, err := os.Executable()
				if err != nil {
					return err
				}

				srcCmd := exec.Command(cmdPath, append([]string{"search"}, args...)...)

				// Because we do not want the default "no color when piping" behavior to take place.
				srcCmd.Env = envSetDefault(os.Environ(), "COLOR", "t")

				srcStderr, err := srcCmd.StderrPipe()
				if err != nil {
					return err
				}
				srcStdout, err := srcCmd.StdoutPipe()
				if err != nil {
					return err
				}
				if err := srcCmd.Start(); err != nil {
					return err
				}

				lessCmd := exec.Command("less", "-R")
				lessCmd.Stdin = io.MultiReader(srcStdout, srcStderr)
				lessCmd.Stderr = os.Stderr
				lessCmd.Stdout = os.Stdout
				return lessCmd.Run()
			}
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `fragment FileMatchFields on FileMatch {
				repository {
					name
					url
				}
				file {
					name
					path
					url
					content
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
				...SearchResultsAlertFields
			  }
			}
		  }
		` + searchResultsAlertFragment

		var result struct {
			Site struct {
				BuildVersion string
			}
			Search struct {
				Results searchResults
			}
		}

		if ok, err := client.NewRequest(query, map[string]interface{}{
			"query": api.NullString(queryString),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

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
		return execTemplate(tmpl, improved)
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

// searchResults represents the data we get back from the GraphQL search request.
type searchResults struct {
	Results                    []map[string]interface{}
	LimitHit                   bool
	Cloning, Missing, Timedout []map[string]interface{}
	ResultCount                int
	ElapsedMilliseconds        int
	Alert                      searchResultsAlert
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
	diff := searchHighlightPreview(diffPreview, uniqueStartOfMatchToken, uniqueEndOfMatchToken)

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
		line = strings.ReplaceAll(line, uniqueStartOfMatchToken, ansiColors["search-match"])

		// Replace our end-of-match token with the color terminator,
		// and start all colors that were previously started to the left.
		line = strings.ReplaceAll(line, uniqueEndOfMatchToken, ansiColors["nc"]+strings.Join(left, ""))

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

// applyHighlightsForFile expects highlight information that is
// relative to the whole file, and can add lines of context around the
// highlighted range. It makes no assumptions about the preview field in
// LineMatches (the preview field is not used).
func applyHighlightsForFile(fileContent string, highlights []highlight) string {
	var result []rune
	start := ansiColors["search-match"]
	end := ansiColors["nc"]
	lines := strings.Split(fileContent, "\n")
	for _, highlight := range highlights {
		line := lines[highlight.line]
		for characterIndex, character := range []rune(line + "\n") {
			if characterIndex == highlight.character {
				result = append(result, []rune(start)...)
			} else if characterIndex == highlight.character+highlight.length {
				result = append(result, []rune(end)...)
			}
			result = append(result, character)
		}
	}
	return string(result)
}

// applyHighlights expects highlight information that applies relative to lines in
// the input string, where the input string corresponds to the preview field in LineMatches.
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

// convertMatchToHighlights converts a FileMatch m to a highlight data type.
// When isPreview is true, it is assumed that the result to highlight is only on
// one line, and the offets are relative to this line. When isPreview is false,
// the lineNumber from the FileMatch data is used, which is relative to the file
// content.
func convertMatchToHighlights(m map[string]interface{}, isPreview bool) (highlights []highlight) {
	var line int
	for _, offsetAndLength := range m["offsetAndLengths"].([]interface{}) {
		ol := offsetAndLength.([]interface{})
		offset := int(ol[0].(float64))
		length := int(ol[1].(float64))
		if isPreview {
			line = 1
		} else {
			line = int(m["lineNumber"].(float64))
		}
		highlights = append(highlights, highlight{line: line, character: offset, length: length})
	}
	return highlights
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
	"searchHighlightMatch": func(content, query, match interface{}) string {
		m := match.(map[string]interface{})
		q := query.(string)
		var highlights []highlight
		if strings.Contains(q, "patterntype:structural") {
			highlights = convertMatchToHighlights(m, false)
			return applyHighlightsForFile(content.(string), highlights)
		} else {
			preview := m["preview"].(string)
			highlights = convertMatchToHighlights(m, true)
			return applyHighlights(preview, highlights, ansiColors["search-match"], ansiColors["nc"])
		}
	},
	"searchHighlightPreview": func(preview interface{}) string {
		return searchHighlightPreview(preview, "", "")
	},
	"searchHighlightDiffPreview": searchHighlightDiffPreview,
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
	"htmlToPlainText":                   htmlToPlainText,
	"buildVersionHasNewSearchInterface": buildVersionHasNewSearchInterface,
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

{{- /* Any alert returned from the search */ -}}
	{{- searchAlertRender .Alert -}}

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
			{{- $content := .file.content -}}
			{{- range $index, $match := $lineMatches -}}
				{{- if not (searchSequentialLineNumber $lineMatches $index) -}}
					{{- color "search-border"}}{{"  ------------------------------------------------------------------------------\n"}}{{color "nc"}}
				{{- end -}}
				{{- "  "}}{{color "search-line-numbers"}}{{pad (addFloat $match.lineNumber 1) 6 " "}}{{color "nc" -}}
				{{- color "search-border"}}{{" |  "}}{{color "nc"}}{{searchHighlightMatch $content $.Query $match}}
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

const searchJSONExplanation = `Explanation of 'src search -json' output:

'src search -json' outputs the exact same results that are retrieved from
Sourcegraph's GraphQL API (see https://about.sourcegraph.com/docs/features/api/)

At a high-level there are three result types:

- 'FileMatch': the type of result you get without any 'type:' modifiers.
- 'CommitSearchResult': the type of result you get with a 'type:commit' or
  'type:diff' modifier.
- 'Repository': the type of result you get with a 'type:repo' modifier.

All three of these result types have different fields available. They can be
differentiated by using the '__typename' field.

The link below shows the GraphQL query that this program internally
executes when querying for search results. On this page, you can hover over
any field in the GraphQL panel on the left to get documentation about the field
itself.

If you have any questions, feedback, or suggestions, please contact us
(support@sourcegraph.com) or file an issue! :)

https://sourcegraph.com/api/console#%7B%22query%22%3A%22fragment%20FileMatchFields%20on%20FileMatch%20%7B%5Cn%20%20repository%20%7B%5Cn%20%20%20%20name%5Cn%20%20%20%20url%5Cn%20%20%7D%5Cn%20%20file%20%7B%5Cn%20%20%20%20name%5Cn%20%20%20%20path%5Cn%20%20%20%20url%5Cn%20%20%20%20commit%20%7B%5Cn%20%20%20%20%20%20oid%5Cn%20%20%20%20%7D%5Cn%20%20%7D%5Cn%20%20lineMatches%20%7B%5Cn%20%20%20%20preview%5Cn%20%20%20%20lineNumber%5Cn%20%20%20%20offsetAndLengths%5Cn%20%20%20%20limitHit%5Cn%20%20%7D%5Cn%7D%5Cn%5Cnfragment%20CommitSearchResultFields%20on%20CommitSearchResult%20%7B%5Cn%20%20messagePreview%20%7B%5Cn%20%20%20%20value%5Cn%20%20%20%20highlights%20%7B%5Cn%20%20%20%20%20%20line%5Cn%20%20%20%20%20%20character%5Cn%20%20%20%20%20%20length%5Cn%20%20%20%20%7D%5Cn%20%20%7D%5Cn%20%20diffPreview%20%7B%5Cn%20%20%20%20value%5Cn%20%20%20%20highlights%20%7B%5Cn%20%20%20%20%20%20line%5Cn%20%20%20%20%20%20character%5Cn%20%20%20%20%20%20length%5Cn%20%20%20%20%7D%5Cn%20%20%7D%5Cn%20%20commit%20%7B%5Cn%20%20%20%20repository%20%7B%5Cn%20%20%20%20%20%20name%5Cn%20%20%20%20%7D%5Cn%20%20%20%20oid%5Cn%20%20%20%20url%5Cn%20%20%20%20subject%5Cn%20%20%20%20author%20%7B%5Cn%20%20%20%20%20%20date%5Cn%20%20%20%20%20%20person%20%7B%5Cn%20%20%20%20%20%20%20%20displayName%5Cn%20%20%20%20%20%20%7D%5Cn%20%20%20%20%7D%5Cn%20%20%7D%5Cn%7D%5Cn%5Cnfragment%20RepositoryFields%20on%20Repository%20%7B%5Cn%20%20name%5Cn%20%20url%5Cn%20%20externalURLs%20%7B%5Cn%20%20%20%20serviceType%5Cn%20%20%20%20url%5Cn%20%20%7D%5Cn%7D%5Cn%5Cnquery%20(%24query%3A%20String!)%20%7B%5Cn%20%20search(query%3A%20%24query)%20%7B%5Cn%20%20%20%20results%20%7B%5Cn%20%20%20%20%20%20results%20%7B%5Cn%20%20%20%20%20%20%20%20__typename%5Cn%20%20%20%20%20%20%20%20...%20on%20FileMatch%20%7B%5Cn%20%20%20%20%20%20%20%20%20%20...FileMatchFields%5Cn%20%20%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20%20%20...%20on%20CommitSearchResult%20%7B%5Cn%20%20%20%20%20%20%20%20%20%20...CommitSearchResultFields%5Cn%20%20%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20%20%20...%20on%20Repository%20%7B%5Cn%20%20%20%20%20%20%20%20%20%20...RepositoryFields%5Cn%20%20%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20limitHit%5Cn%20%20%20%20%20%20cloning%20%7B%5Cn%20%20%20%20%20%20%20%20name%5Cn%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20missing%20%7B%5Cn%20%20%20%20%20%20%20%20name%5Cn%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20timedout%20%7B%5Cn%20%20%20%20%20%20%20%20name%5Cn%20%20%20%20%20%20%7D%5Cn%20%20%20%20%20%20resultCount%5Cn%20%20%20%20%20%20elapsedMilliseconds%5Cn%20%20%20%20%7D%5Cn%20%20%7D%5Cn%7D%5Cn%22%2C%22variables%22%3A%22%7B%5Cn%20%20%5C%22query%5C%22%3A%20%5C%22repogroup%3Asample%20error%5C%22%5Cn%7D%22%7D
`
