// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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

				got, err := squirrel.definition(annotation.location)
				fatalIfError(t, err)

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
	nextSourceLine:
		for sourceLine := 0; sourceLine < len(lines); {
			if matches := regexp.MustCompile(`^([^<]+)< "([^"]+)" ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[sourceLine]); matches != nil {
				// After line annotation
				column := strings.Index(lines[sourceLine], matches[2])
				annotations = append(annotations, annotation{
					location: Location{
						RepoCommitPath: RepoCommitPath{
							Repo:   "foo",
							Commit: "bar",
							Path:   path,
						},
						Row:    uint32(sourceLine),
						Column: uint32(column),
					},
					symbol: matches[3],
					kind:   matches[4],
				})

				sourceLine++
				continue nextSourceLine
			} else {
				// Next line annotation
				for annLine := sourceLine + 1; ; annLine++ {
					if annLine >= len(lines) {
						break nextSourceLine
					}

					matches := regexp.MustCompile(`([^^]*)(\^+) ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[annLine])
					if matches == nil {
						sourceLine = annLine
						continue nextSourceLine
					}

					annotations = append(annotations, annotation{
						location: Location{
							RepoCommitPath: RepoCommitPath{
								Repo:   "foo",
								Commit: "bar",
								Path:   path,
							},
							Row:    uint32(sourceLine),
							Column: uint32(spacesToColumn(lines[sourceLine], lengthInSpaces(matches[1]))),
						},
						symbol: matches[3],
						kind:   matches[4],
					})
				}
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
