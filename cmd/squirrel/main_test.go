// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
)

func TestFixturesAreValid(t *testing.T) {
	pathToValidateCmd := map[string]string{
		"simple": "go build . && rm -f simple",
	}

	for path, validateCmd := range pathToValidateCmd {
		cmd := exec.Command("bash", "-c", validateCmd)
		cwd := filepath.Join(TEST_REPOS_DIR, path)
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("when running %q in %s: %s, combined output:\n\n%s", validateCmd, cwd, err, string(output))
		}
	}
}

func TestDefinition(t *testing.T) {
	prevPath := ""
	prevContents := []byte{}
	readFile := func(path RepoCommitPath) ([]byte, error) {
		if path.Path == prevPath {
			return prevContents, nil
		}
		prevPath = path.Path
		contents, err := os.ReadFile(path.Path)
		prevContents = contents
		return contents, err
	}

	squirrel := NewSquirrel(readFile)
	annotations := collectAnnotations(t)
	pathToAnnotations := groupAnnotationsByPath(annotations)
	symbolToAnnotations := groupAnnotationsBysymbol(annotations)
	for _, annotations := range pathToAnnotations {
	nextAnnotation:
		for _, annotation := range annotations {
			if annotation.kind == "ref" {
				var want *Location
				for _, ann := range symbolToAnnotations[annotation.symbol] {
					if ann.kind == "def" {
						annCopy := ann // avoid capturing the loop variable
						want = &annCopy.location
						break
					}
				}

				if want == nil {
					t.Fatalf("tests are missing a definition for %s", annotation.symbol)
					continue
				}

				got, breadcrumbs, err := squirrel.definition(annotation.location)
				if err != nil {
					breadcrumbs = append(breadcrumbs, Breadcrumb{Location: *want, length: 1, message: "correct"})
					prettyPrintBreadcrumbs(t, breadcrumbs, readFile)
					t.Fatal(err)
				}

				if got == nil {
					t.Fatalf("definition(%+v) returned nil", annotation.location)
				}

				for _, ann := range symbolToAnnotations[annotation.symbol] {
					if ann.kind == "def" && ann.location == *got {
						// ✅ Found the definition.
						continue nextAnnotation
					}
				}

				// ❌ Could not find the definition.
				if diff := cmp.Diff(want, *got); diff != "" {
					t.Fatalf("definition(%+v) returned an incorrect location, -got +want:\n\n%s", annotation.location, diff)
				}
			}
		}
	}
}

type annotation struct {
	location Location
	symbol   string
	kind     string
}

func collectAnnotations(t *testing.T) []annotation {
	annotations := []annotation{}

	filepath.WalkDir(TEST_REPOS_DIR, func(path string, file os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if file.IsDir() {
			return nil
		}

		contents, err := os.ReadFile(path)
		fatalIfError(t, err)

		lines := strings.Split(strings.TrimSpace(string(contents)), "\n")

		// Annotation at the end of the line
		for i, line := range lines {
			matches := regexp.MustCompile(`^([^<]+)< "([^"]+)" ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(line)
			if matches == nil {
				continue
			}

			substr, symbol, kind := matches[2], matches[3], matches[4]

			annotations = append(annotations, annotation{
				location: Location{
					RepoCommitPath: RepoCommitPath{
						Repo:   "foo",
						Commit: "bar",
						Path:   path,
					},
					Row:    uint32(i),
					Column: uint32(strings.Index(line, substr)),
				},
				symbol: symbol,
				kind:   kind,
			})
		}

		// Annotations below source lines
	nextSourceLine:
		for sourceLine := 0; ; {
			for annLine := sourceLine + 1; ; annLine++ {
				if annLine >= len(lines) {
					break nextSourceLine
				}

				matches := regexp.MustCompile(`([^^]*)\^+ ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[annLine])
				if matches == nil {
					sourceLine = annLine
					continue nextSourceLine
				}

				prefix, symbol, kind := matches[1], matches[2], matches[3]

				annotations = append(annotations, annotation{
					location: Location{
						RepoCommitPath: RepoCommitPath{
							Repo:   "foo",
							Commit: "bar",
							Path:   path,
						},
						Row:    uint32(sourceLine),
						Column: uint32(spacesToColumn(lines[sourceLine], lengthInSpaces(prefix))),
					},
					symbol: symbol,
					kind:   kind,
				})
			}
		}

		// Annotations above source lines
	previousSourceLine:
		for sourceLine := len(lines) - 1; ; {
			for annLine := sourceLine - 1; ; annLine-- {
				if annLine <= 0 {
					break previousSourceLine
				}

				matches := regexp.MustCompile(`([^v]*)v+ ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[annLine])
				if matches == nil {
					sourceLine = annLine
					continue previousSourceLine
				}

				prefix, symbol, kind := matches[1], matches[2], matches[3]

				annotations = append(annotations, annotation{
					location: Location{
						RepoCommitPath: RepoCommitPath{
							Repo:   "foo",
							Commit: "bar",
							Path:   path,
						},
						Row:    uint32(sourceLine),
						Column: uint32(spacesToColumn(lines[sourceLine], lengthInSpaces(prefix))),
					},
					symbol: symbol,
					kind:   kind,
				})
			}
		}

		return nil
	})

	return annotations
}

func groupAnnotationsByPath(annotations []annotation) map[string][]annotation {
	pathToAnnotations := map[string][]annotation{}
	for _, annotation := range annotations {
		pathToAnnotations[annotation.location.Path] = append(pathToAnnotations[annotation.location.Path], annotation)
	}
	return pathToAnnotations
}

func groupAnnotationsBysymbol(annotations []annotation) map[string][]annotation {
	symbolToAnnotations := map[string][]annotation{}
	for _, annotation := range annotations {
		symbolToAnnotations[annotation.symbol] = append(symbolToAnnotations[annotation.symbol], annotation)
	}
	return symbolToAnnotations
}

func lengthInSpaces(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

func spacesToColumn(s string, ix int) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if total >= ix {
			return i
		}

		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

const TEST_REPOS_DIR = "test_repos"

func prettyPrintBreadcrumbs(t *testing.T, breadcrumbs []Breadcrumb, readFile ReadFileFunc) {
	sb := &strings.Builder{}

	m := map[RepoCommitPath]map[int][]Breadcrumb{}
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
		fmt.Fprintf(sb, blue("repo %s, commit %s, path %s"), repoCommitPath.Repo, repoCommitPath.Commit, repoCommitPath.Path)
		fmt.Fprintln(sb)

		contents, err := readFile(repoCommitPath)
		fatalIfError(t, err)
		lines := strings.Split(string(contents), "\n")
		for lineNumber, line := range lines {
			breadcrumbs, ok := lineToBreadcrumb[lineNumber]
			if !ok {
				continue
			}

			fmt.Fprintln(sb)

			gutter := fmt.Sprintf("%5d | ", lineNumber)

			columnToMessage := map[int]string{}
			for _, breadcrumb := range breadcrumbs {
				for column := int(breadcrumb.Column); column < int(breadcrumb.Column)+breadcrumb.length; column++ {
					columnToMessage[lengthInSpaces(line[:column])] = breadcrumb.message
				}

				gutterPadding := strings.Repeat(" ", len(gutter))

				space := strings.Repeat(" ", lengthInSpaces(line[:breadcrumb.Column]))

				arrows := messageColor(breadcrumb.message)(strings.Repeat("v", breadcrumb.length))

				fmt.Fprintf(sb, "%s%s%s %s\n", gutterPadding, space, arrows, messageColor(breadcrumb.message)(breadcrumb.message))
			}

			fmt.Fprint(sb, grey(gutter))
			lineWithSpaces := strings.ReplaceAll(line, "\t", "    ")
			for c := 0; c < len(lineWithSpaces); c++ {
				if message, ok := columnToMessage[c]; ok {
					fmt.Fprint(sb, messageColor(message)(string(lineWithSpaces[c])))
				} else {
					fmt.Fprint(sb, grey(string(lineWithSpaces[c])))
				}
			}
			fmt.Fprintln(sb)
		}
	}

	fmt.Println(bracket(sb.String()))
}

type colorSprintfFunc func(a ...interface{}) string

func messageColor(message string) colorSprintfFunc {
	switch message {
	case "start":
		return color.New(color.FgHiCyan).SprintFunc()
	case "found":
		return color.New(color.FgRed).SprintFunc()
	case "correct":
		return color.New(color.FgGreen).SprintFunc()
	default:
		return color.New(color.FgHiMagenta).SprintFunc()
	}
}

func bracket(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 1 {
		return "- " + text
	}

	for i, line := range lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i < len(lines)-1 {
			lines[i] = "│ " + line
		} else {
			lines[i] = "└ " + line
		}
	}

	return strings.Join(lines, "\n")
}
