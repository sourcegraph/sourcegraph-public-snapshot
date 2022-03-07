package squirrel

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Breadcrumb struct {
	types.RepoCommitPathRange
	length  int
	message string
}

func prettyPrintBreadcrumbs(w *strings.Builder, breadcrumbs []Breadcrumb, readFile ReadFileFunc) {
	m := map[types.RepoCommitPath]map[int][]Breadcrumb{}
	for _, breadcrumb := range breadcrumbs {
		path := breadcrumb.RepoCommitPath

		if _, ok := m[path]; !ok {
			m[path] = map[int][]Breadcrumb{}
		}

		m[path][int(breadcrumb.Row)] = append(m[path][int(breadcrumb.Row)], breadcrumb)
	}

	for repoCommitPath, lineToBreadcrumb := range m {
		blue := color.New(color.FgBlue).SprintFunc()
		grey := color.New(color.FgBlack).SprintFunc()
		fmt.Fprintf(w, blue("repo %s, commit %s, path %s"), repoCommitPath.Repo, repoCommitPath.Commit, repoCommitPath.Path)
		fmt.Fprintln(w)

		contents, err := readFile(context.Background(), repoCommitPath)
		if err != nil {
			fmt.Println("Error reading file: ", err)
			return
		}
		lines := strings.Split(string(contents), "\n")
		for lineNumber, line := range lines {
			breadcrumbs, ok := lineToBreadcrumb[lineNumber]
			if !ok {
				continue
			}

			fmt.Fprintln(w)

			gutter := fmt.Sprintf("%5d | ", lineNumber)

			columnToMessage := map[int]string{}
			for _, breadcrumb := range breadcrumbs {
				for column := int(breadcrumb.Column); column < int(breadcrumb.Column)+breadcrumb.length; column++ {
					columnToMessage[lengthInSpaces(line[:column])] = breadcrumb.message
				}

				gutterPadding := strings.Repeat(" ", len(gutter))

				space := strings.Repeat(" ", lengthInSpaces(line[:breadcrumb.Column]))

				arrows := messageColor(breadcrumb.message)(strings.Repeat("v", breadcrumb.length))

				fmt.Fprintf(w, "%s%s%s %s\n", gutterPadding, space, arrows, messageColor(breadcrumb.message)(breadcrumb.message))
			}

			fmt.Fprint(w, grey(gutter))
			lineWithSpaces := strings.ReplaceAll(line, "\t", "    ")
			for c := 0; c < len(lineWithSpaces); c++ {
				if message, ok := columnToMessage[c]; ok {
					fmt.Fprint(w, messageColor(message)(string(lineWithSpaces[c])))
				} else {
					fmt.Fprint(w, grey(string(lineWithSpaces[c])))
				}
			}
			fmt.Fprintln(w)
		}
	}
}

func pickBreadcrumbs(breadcrumbs []Breadcrumb, messages []string) []Breadcrumb {
	var picked []Breadcrumb
	for _, breadcrumb := range breadcrumbs {
		for _, message := range messages {
			if strings.Contains(breadcrumb.message, message) {
				picked = append(picked, breadcrumb)
				break
			}
		}
	}
	return picked
}

func messageColor(message string) colorSprintfFunc {
	if message == "start" {
		return color.New(color.FgHiCyan).SprintFunc()
	} else if message == "found" {
		return color.New(color.FgRed).SprintFunc()
	} else if message == "correct" {
		return color.New(color.FgGreen).SprintFunc()
	} else if strings.Contains(message, "scope") {
		return color.New(color.FgHiYellow).SprintFunc()
	} else {
		return color.New(color.FgHiMagenta).SprintFunc()
	}
}
