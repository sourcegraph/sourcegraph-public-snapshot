package squirrel

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Breadcrumb is an arbitrary annotation on a token in a file. It's used as a way to log where Squirrel
// has been traversing through trees and files for debugging.
type Breadcrumb struct {
	types.RepoCommitPathRange
	length  int
	message string
}

// addBreadcrumb adds a breadcrumb to the given slice.
func addBreadcrumb(breadcrumbs *[]Breadcrumb, node Node, message string) {
	*breadcrumbs = append(*breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: node.RepoCommitPath,
			Range:          nodeToRange(node.Node),
		},
		length:  nodeLength(node.Node),
		message: message,
	})
}

// Prints breadcrumbs like this:
//
//             v some breadcrumb
//               vvv other breadcrumb
// 78 | func f(f Foo) {
func prettyPrintBreadcrumbs(w *strings.Builder, breadcrumbs []Breadcrumb, readFile ReadFileFunc) {
	// First collect all the breadcrumbs in a map (path -> line -> breadcrumb) for easier printing.
	pathToLineToBreadcrumbs := map[types.RepoCommitPath]map[int][]Breadcrumb{}
	for _, breadcrumb := range breadcrumbs {
		path := breadcrumb.RepoCommitPath

		if _, ok := pathToLineToBreadcrumbs[path]; !ok {
			pathToLineToBreadcrumbs[path] = map[int][]Breadcrumb{}
		}

		pathToLineToBreadcrumbs[path][int(breadcrumb.Row)] = append(pathToLineToBreadcrumbs[path][int(breadcrumb.Row)], breadcrumb)
	}

	// Loop over each path, printing the breadcrumbs for each line.
	for repoCommitPath, lineToBreadcrumb := range pathToLineToBreadcrumbs {
		// Print the path header.
		blue := color.New(color.FgBlue).SprintFunc()
		grey := color.New(color.FgBlack).SprintFunc()
		fmt.Fprintf(w, blue("repo %s, commit %s, path %s"), repoCommitPath.Repo, repoCommitPath.Commit, repoCommitPath.Path)
		fmt.Fprintln(w)

		// Read the file.
		contents, err := readFile(context.Background(), repoCommitPath)
		if err != nil {
			fmt.Println("Error reading file: ", err)
			return
		}

		// Print the breadcrumbs for each line.
		for lineNumber, line := range strings.Split(string(contents), "\n") {
			breadcrumbs, ok := lineToBreadcrumb[lineNumber]
			if !ok {
				// No breadcrumbs on this line.
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

// Returns breadcrumbs that have one of the given messages.
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

// Returns the color to be used to print a message.
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
