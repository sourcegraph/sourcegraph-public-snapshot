package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: md2mdx <path>")
		os.Exit(1)
	}

	path := os.Args[1]

	if filepath.Ext(path) != ".md" {
		println("Must pass a markdown file")
		os.Exit(1)
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	out, err := convert(f, path)
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}

	if _, err := os.Stdout.Write([]byte(out)); err != nil {
		panic(err)
	}
}

var htmlCommentsRe = regexp.MustCompile(`<!-- (.*) -->`)
var relativeLinksRe = regexp.MustCompile(`\[(.*)\]\((.*\.md)\)`)

func convert(r io.Reader, path string) (string, error) {
	res := []string{}
	s := bufio.NewScanner(r)
	_, parentFolder := filepath.Split(filepath.Dir(path))

	var withinTripleBackticks bool

	for s.Scan() {
		line := s.Text()

		if withinTripleBackticks && strings.Contains(line, "```") {
			withinTripleBackticks = false
			goto APPEND
		}

		if withinTripleBackticks {
			goto APPEND
		}

		// Yes, the first part of the predicate is useless, but it's easier to read.
		if !withinTripleBackticks && strings.Contains(line, "```") {
			withinTripleBackticks = true
			goto APPEND
		}

		line = htmlCommentsRe.ReplaceAllString(line, `{/* $1 */}`)
		line = replaceTagsOrLtGt(line)
		line = replaceLinks(line, parentFolder)
		line = replaceCurlies(line)
		line = replacePipes(line)

	APPEND:
		res = append(res, line)
	}

	if err := s.Err(); err != nil {
		return "", err
	}

	return strings.Join(res, "\n"), nil
}

func replaceLinks(s string, parentFolder string) string {
	matches := relativeLinksRe.FindStringSubmatch(s)
	if len(matches) == 0 {
		return s
	}

	content := matches[0]
	text := matches[1]
	link := matches[2]

	if strings.HasSuffix(link, "/index.md") {
		link = strings.TrimSuffix(link, "/index.md")
	} else if strings.HasSuffix(link, ".md") {
		link = link + "x"
	}

	return strings.Replace(s, content, fmt.Sprintf("[%s](%s/%s)", text, parentFolder, link), -1)
}

var tagsRe = regexp.MustCompile(`<([^>]+)>`)

func replaceTagsOrLtGt(line string) string {
	if !tagsRe.MatchString(line) {
		line = strings.ReplaceAll(line, "<", "&lt;")
		line = strings.ReplaceAll(line, ">", "&gt;")
	}
	return line
}

func replaceCurlies(s string) string {
	singleQuoteRe := regexp.MustCompile(`'([^']+)'`)
	doubleQuoteRe := regexp.MustCompile(`"([^']+)"`)

	replace := func(s string, r *regexp.Regexp) string {
		chunks := r.FindAllString(s, -1)
		if len(chunks) == 0 {
			return s
		}
		for _, content := range chunks {
			newContent := content
			newContent = strings.Replace(newContent, "{", "\\{", -1)
			newContent = strings.Replace(newContent, "}", "\\}", -1)
			newContent = strings.Replace(newContent, "|", "\\|", -1)

			s = strings.ReplaceAll(s, content, newContent)
		}
		return s
	}

	for _, r := range []*regexp.Regexp{singleQuoteRe, doubleQuoteRe} {
		s = replace(s, r)
	}

	return s
}

func replacePipes(s string) string {
	backticksRe := regexp.MustCompile("`([^`]+)`")

	chunks := backticksRe.FindAllString(s, -1)
	if len(chunks) == 0 {
		return s
	}
	for _, content := range chunks {
		newContent := content
		newContent = strings.Replace(newContent, "|", "\\|", -1)

		s = strings.ReplaceAll(s, content, newContent)
	}
	return s
}
