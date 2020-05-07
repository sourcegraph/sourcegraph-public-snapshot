package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/cmd/precise-code-intel-test/internal"
)

var (
	baseURL                     = flag.String("baseUrl", "http://127.0.0.1:3080", "A Sourcegraph URL")
	token                       = flag.String("token", "", "A Sourcegraph access token")
	maxConcurrency              = flag.Int("maxConcurrency", 5, "The maximum number of concurrent operations")
	checkQueryResult            = flag.Bool("checkQueryResult", true, "Whether to confirm query results are correct")
	queryReferencesOfReferences = flag.Bool("queryReferencesOfReferences", false, "Whether to perform reference operations on test case references")

	// Assumes running from the root of the repo
	cacheDir = flag.String("cacheDir", "/tmp/precise-code-intel-test", "The location of the cache directory")
	dataDir  = flag.String("dataDir", "./internal/cmd/precise-code-intel-test/data", "The location of the data directory")
)

func main() {
	commands := map[string]func() error{
		"clone":  clone,
		"index":  index,
		"upload": upload,
		"query":  query,
	}

	if len(os.Args) < 2 {
		fmt.Println("subcommand (clone, index, upload, or query) is required")
		os.Exit(1)
	}

	command, ok := commands[os.Args[1]]
	if !ok {
		fmt.Println("subcommand (clone, index, upload, or query) is required")
		os.Exit(1)
	}

	if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	if err := command(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func clone() error {
	if err := os.MkdirAll(filepath.Join(*cacheDir, "repos"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(*dataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.CloneParallel(*cacheDir, *maxConcurrency, repos)
}

func index() error {
	if err := os.MkdirAll(filepath.Join(*cacheDir, "indexes"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(*dataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.IndexParallel(*cacheDir, *maxConcurrency, repos)
}

func upload() error {
	repos, err := internal.ReadReposCSV(filepath.Join(*dataDir, "repos.csv"))
	if err != nil {
		return err
	}

	start := time.Now()

	ids, err := internal.UploadParallel(*cacheDir, *baseURL, *maxConcurrency, repos)
	if err != nil {
		return err
	}

	if err := internal.WaitForSuccessAll(*baseURL, *token, ids); err != nil {
		return err
	}

	fmt.Printf("All uploads completed processing in %s\n", time.Since(start))
	return nil
}

func query() error {
	testCases, err := internal.ReadTestCaseCSV(*dataDir, filepath.Join(*dataDir, "test-cases.csv"))
	if err != nil {
		return err
	}

	var fns []internal.FnPair
	for _, testCase := range testCases {
		definition := testCase.Definition
		expectedReferences := testCase.References

		fns = append(fns, internal.FnPair{
			Fn: func() error {
				references, err := internal.QueryReferences(*baseURL, *token, definition)
				if err != nil {
					return err
				}
				if *checkQueryResult {
					if diff := cmp.Diff(expectedReferences, references); diff != "" {
						return fmt.Errorf("unexpected references (-want +got):\n%s\n", diff)
					}
				}
				return nil
			},
			Description: fmt.Sprintf(
				"Checking references for definition %s@%s %s %d:%d",
				definition.Repo,
				definition.Rev[:6],
				definition.Path,
				definition.Line,
				definition.Character,
			),
		})

		for _, reference := range expectedReferences {
			localReference := reference

			fns = append(fns, internal.FnPair{
				Fn: func() error {
					definitions, err := internal.QueryDefinitions(*baseURL, *token, localReference)
					if err != nil {
						return err
					}

					if *checkQueryResult {
						if diff := cmp.Diff([]internal.Location{definition}, definitions); diff != "" {
							return fmt.Errorf("unexpected definitions (-want +got):\n%s\n", diff)
						}
					}
					return nil
				},
				Description: fmt.Sprintf(
					"Checking definitions for %s@%s %s %d:%d",
					reference.Repo,
					reference.Rev[:6],
					reference.Path,
					reference.Line,
					reference.Character,
				),
			})

			if *queryReferencesOfReferences {
				fns = append(fns, internal.FnPair{
					Fn: func() error {
						references, err := internal.QueryReferences(*baseURL, *token, localReference)
						if err != nil {
							return err
						}

						if *checkQueryResult {
							if diff := cmp.Diff(expectedReferences, references); diff != "" {
								return fmt.Errorf("unexpected references (-want +got):\n%s\n", diff)
							}
						}
						return nil
					},
					Description: fmt.Sprintf(
						"Checking references for %s@%s %s %d:%d",
						reference.Repo,
						reference.Rev[:6],
						reference.Path,
						reference.Line,
						reference.Character,
					),
				})
			}
		}
	}

	start := time.Now()

	if err := internal.RunParallel(*maxConcurrency, fns); err != nil {
		return err
	}

	fmt.Printf("All queries completed in %s\n", time.Since(start))
	return nil
}
