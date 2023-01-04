package database

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSiteGetLatestDefault(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Background()
	latest, err := db.Conf().SiteGetLatest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if latest == nil {
		t.Errorf("expected non-nil latest config since default config should be created, got: %+v", latest)
	}
}

func TestSiteCreate_RejectInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	malformedJSON := "[This is malformed.}"

	_, err := db.Conf().SiteCreateIfUpToDate(ctx, nil, malformedJSON, false)

	if err == nil || !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Fatalf("expected parse error after creating configuration with malformed JSON, got: %+v", err)
	}
}

func TestSiteCreateIfUpToDate(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)

	type input struct {
		lastID   int32
		contents string
	}

	type output struct {
		ID       int32
		contents string
		err      error
	}

	type pair struct {
		input    input
		expected output
	}

	type test struct {
		name     string
		sequence []pair
	}

	for _, test := range []test{
		{
			name: "create_one",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
					output{
						ID:       2,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
				},
			},
		},
		{
			name: "create_two",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
					output{
						ID:       2,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
				},
				{
					input{
						lastID:   2,
						contents: `{"defaultRateLimit": 1,"auth.providers": []}`,
					},
					output{
						ID:       3,
						contents: `{"defaultRateLimit": 1,"auth.providers": []}`,
					},
				},
			},
		},
		{
			name: "do_not_update_if_outdated",
			sequence: []pair{
				{
					input{
						lastID:   0,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
					output{
						ID:       2,
						contents: `{"defaultRateLimit": 0,"auth.providers": []}`,
					},
				},
				{
					input{
						lastID: 0,
						// This configuration is now behind the first one, so it shouldn't be saved
						contents: `{"defaultRateLimit": 1,"auth.providers": []}`,
					},
					output{
						ID:       2,
						contents: `{"defaultRateLimit": 1,"auth.providers": []}`,
						err:      errors.Append(ErrNewerEdit),
					},
				},
			},
		},
		{
			name: "maintain_commments_and_whitespace",
			sequence: []pair{
				{
					input{
						lastID: 0,
						contents: `{"disableAutoGitUpdates": true,

// This is a comment.
             "defaultRateLimit": 42,
             "auth.providers": [],
						}`,
					},
					output{
						ID: 2,
						contents: `{"disableAutoGitUpdates": true,

// This is a comment.
             "defaultRateLimit": 42,
             "auth.providers": [],
						}`,
					},
				},
			},
		},
	} {
		// we were running the same test all the time, see this gist for more information
		// https://gist.github.com/posener/92a55c4cd441fc5e5e85f27bca008721
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db := NewDB(logger, dbtest.NewDB(logger, t))
			ctx := context.Background()
			for _, p := range test.sequence {
				output, err := db.Conf().SiteCreateIfUpToDate(ctx, &p.input.lastID, p.input.contents, false)
				if err != nil {
					if errors.Is(err, p.expected.err) {
						continue
					}
					t.Fatal(err)
				}

				if output == nil {
					t.Fatal("got unexpected nil configuration after creation")
				}

				if output.Contents != p.expected.contents {
					t.Fatalf("returned configuration contents after creation - expected: %q, got:%q", p.expected.contents, output.Contents)
				}
				if output.ID != p.expected.ID {
					t.Fatalf("returned configuration ID after creation - expected: %v, got:%v", p.expected.ID, output.ID)
				}

				latest, err := db.Conf().SiteGetLatest(ctx)
				if err != nil {
					t.Fatal(err)
				}

				if latest == nil {
					t.Fatalf("got unexpected nil configuration after GetLatest")
				}

				if latest.Contents != p.expected.contents {
					t.Fatalf("returned configuration contents after GetLatest - expected: %q, got:%q", p.expected.contents, latest.Contents)
				}
				if latest.ID != p.expected.ID {
					t.Fatalf("returned configuration ID after GetLatest - expected: %v, got:%v", p.expected.ID, latest.ID)
				}
			}
		})
	}
}
