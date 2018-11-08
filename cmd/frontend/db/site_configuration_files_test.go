package db

import (
	"strings"
	"testing"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
)

func TestSiteConfigurationFiles_GetLatestEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	latestFile, err := SiteConfigurationFiles.GetLatest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if latestFile != nil {
		t.Errorf("expected nil latestFile since no site configuration was created, got: %+v", latestFile)
	}
}

func TestSiteConfigurationFiles_Create_RejectInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	malformedJSON := "[This is malformed.}"

	_, err := SiteConfigurationFiles.CreateIfUpToDate(ctx, nil, malformedJSON)

	if err == nil || !strings.Contains(err.Error(), "invalid settings JSON") {
		t.Fatalf("expected parse error after creating site configuration with malformed JSON, got: %+v", err)
	}
}

func TestSiteConfigurationFiles_Create(t *testing.T) {
	type input struct {
		lastID   int32
		contents string
	}

	type output struct {
		contents string
	}

	type pair struct {
		input    input
		expected output
	}

	type test struct {
		description string
		sequence    []pair
	}

	for _, test := range []test{
		test{
			description: "create one site configuration file",
			sequence: []pair{
				pair{
					input{
						lastID:   0,
						contents: `"This is a test."`,
					},
					output{
						contents: `"This is a test."`,
					},
				},
			},
		},
		test{
			description: "create two site configuration files",
			sequence: []pair{
				pair{
					input{
						lastID:   0,
						contents: `"This is the first one."`,
					},
					output{
						contents: `"This is the first one."`,
					},
				},
				pair{
					input{
						lastID:   1,
						contents: `"This is the second one."`,
					},
					output{
						contents: `"This is the second one."`,
					},
				},
			},
		},
		test{
			description: "shouldn't update the configuration if the caller is out of date",
			sequence: []pair{
				pair{
					input{
						lastID:   0,
						contents: `"This is the first one."`,
					},
					output{
						contents: `"This is the first one."`,
					},
				},
				pair{
					input{
						lastID:   0,
						contents: `"This configuration is now behind the first one, so it shouldn't be saved."`,
					},
					output{
						contents: `"This is the first one."`,
					},
				},
			},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			ctx := dbtesting.TestContext(t)
			for _, p := range test.sequence {
				output, err := SiteConfigurationFiles.CreateIfUpToDate(ctx, &p.input.lastID, p.input.contents)
				if err != nil {
					t.Fatal(err)
				}

				if output == nil {
					t.Fatal("got unexpected nil site configuration file after creation")
				}

				if output.Contents != p.expected.contents {
					t.Fatalf("returned site configuration file contents after creation - expected: %q, got:%q", p.expected.contents, output.Contents)
				}

				latestFile, err := SiteConfigurationFiles.GetLatest(ctx)
				if err != nil {
					t.Fatal(err)
				}

				if latestFile == nil {
					t.Errorf("got unexpected nil site configuration file after GetLatest")
				}

				if latestFile.Contents != p.expected.contents {
					t.Fatalf("returned site configuration file contents after GetLatest - expected: %q, got:%q", p.expected.contents, latestFile.Contents)
				}
			}
		})
	}
}
