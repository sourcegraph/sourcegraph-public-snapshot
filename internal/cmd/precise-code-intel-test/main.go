package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/cmd/precise-code-intel-test/internal"
)

const MaxConcurrency = 3
const MaxIndexConcurrency = 2
const MaxQueryConcurrency = 25
const BaseURL = "http://localhost:3080"
const Token = "fd1a128890e12bc6e9d00a8a33d68b545b1adf0f"
const CacheDir = "/tmp/precise-code-intel-test"
const DataDir = "./internal/cmd/precise-code-intel-test/data"
const CheckQueryResult = true
const QueryReferenceReferences = false

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

	if err := command(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func clone() error {
	if err := os.MkdirAll(filepath.Join(CacheDir, "repos"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.CloneParallel(CacheDir, MaxConcurrency, repos)
}

func index() error {
	if err := os.MkdirAll(filepath.Join(CacheDir, "indexes"), os.ModePerm); err != nil {
		return err
	}

	repos, err := internal.ReadReposCSV(filepath.Join(DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	return internal.IndexParallel(CacheDir, MaxIndexConcurrency, repos)
}

func upload() error {
	repos, err := internal.ReadReposCSV(filepath.Join(DataDir, "repos.csv"))
	if err != nil {
		return err
	}

	start := time.Now()

	ids, err := internal.UploadParallel(CacheDir, BaseURL, MaxConcurrency, repos)
	if err != nil {
		return err
	}

	if err := internal.WaitForSuccessAll(BaseURL, Token, ids); err != nil {
		return err
	}

	fmt.Printf("All uploads completed processing in %s\n", time.Since(start))
	return nil
}

func query() error {
	testCases, err := internal.ReadTestCaseCSV(DataDir, filepath.Join(DataDir, "test-cases.csv"))
	if err != nil {
		return err
	}

	var fns []internal.FnPair
	for _, testCase := range testCases {
		definition := testCase.Definition
		expectedReferences := testCase.References

		fns = append(fns, internal.FnPair{
			Fn: func() error {
				references, err := internal.QueryReferences(BaseURL, Token, definition)
				if err != nil {
					return err
				}
				if CheckQueryResult {
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
					definitions, err := internal.QueryDefinitions(BaseURL, Token, localReference)
					if err != nil {
						return err
					}

					if CheckQueryResult {
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

			if QueryReferenceReferences {
				fns = append(fns, internal.FnPair{
					Fn: func() error {
						references, err := internal.QueryReferences(BaseURL, Token, localReference)
						if err != nil {
							return err
						}

						if CheckQueryResult {
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

	if err := internal.RunParallel(MaxQueryConcurrency, fns); err != nil {
		return err
	}

	fmt.Printf("All queries completed in %s\n", time.Since(start))
	return nil
}
