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
	BaseURL                     = flag.String("baseUrl", "https://sourcegraph.test:3443", "A Sourcegraph URL")
	Token                       = flag.String("token", "", "A Sourcegraph access token")
	MaxConcurrency              = flag.Int("maxConcurrency", 5, "The maximum number of concurrent operations")
	CheckQueryResult            = flag.Bool("checkQueryResult", true, "Whether to confirm query results are correct")
	QueryReferencesOfReferences = flag.Bool("queryReferencesOfReferences", false, "Whether to perform reference operations on test case references")

	// Assumes running from the root of the repo
	CacheDir = flag.String("cacheDir", "/tmp/precise-code-intel-test", "The location of the cache directory")
	DataDir  = flag.String("dataDir", "./internal/cmd/precise-code-intel-test/data", "The location of the data directory")
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
	if err := os.MkdirAll(filepath.Join(*CacheDir, "repos"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(*DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.CloneParallel(*CacheDir, *MaxConcurrency, repos)
}

func index() error {
	if err := os.MkdirAll(filepath.Join(*CacheDir, "indexes"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(*DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.IndexParallel(*CacheDir, *MaxConcurrency, repos)
}

func upload() error {
	repos, err := internal.ReadReposCSV(filepath.Join(*DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	start := time.Now()

	ids, err := internal.UploadParallel(*CacheDir, *BaseURL, *MaxConcurrency, repos)
	if err != nil {
		return err
	}

	if err := internal.WaitForSuccessAll(*BaseURL, *Token, ids); err != nil {
		return err
	}

	fmt.Printf("All uploads completed processing in %s\n", time.Since(start))
	return nil
}

func query() error {
	testCases, err := internal.ReadTestCaseCSV(*DataDir, filepath.Join(*DataDir, "test-cases.csv"))
	if err != nil {
		return err
	}

	var fns []internal.FnPair
	for _, testCase := range testCases {
		definition := testCase.Definition
		expectedReferences := testCase.References

		fns = append(fns, internal.FnPair{
			Fn: func() error {
				references, err := internal.QueryReferences(*BaseURL, *Token, definition)
				if err != nil {
					return err
				}
				if *CheckQueryResult {
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
					definitions, err := internal.QueryDefinitions(*BaseURL, *Token, localReference)
					if err != nil {
						return err
					}

					if *CheckQueryResult {
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

			if *QueryReferencesOfReferences {
				fns = append(fns, internal.FnPair{
					Fn: func() error {
						references, err := internal.QueryReferences(*BaseURL, *Token, localReference)
						if err != nil {
							return err
						}

						if *CheckQueryResult {
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

	if err := internal.RunParallel(*MaxConcurrency, fns); err != nil {
		return err
	}

	fmt.Printf("All queries completed in %s\n", time.Since(start))
	return nil
}
